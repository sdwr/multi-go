package main

import (
    "log"

    "github.com/sdwr/multi-go/socket"
    . "github.com/sdwr/multi-go/types"
)

var GlobalRoom *socket.Room
var queueRoom *socket.Room
var gameRooms map[int]*socket.Room
var GameHandler chan *Message

func InitCoordinator() *socket.Room {
    initRooms()
    return GlobalRoom
}	

func RunCoordinator() {
    run()
}

func initRooms() {
    GlobalRoom = socket.NewRoom()
    queueRoom = socket.NewRoom()
    gameRooms = make(map[int]*socket.Room)
    GameHandler = make(chan *Message, 10)
}

func run() {
    for {
        select {
	case m := <- GlobalRoom.Incoming:
            handleGlobalMessage(m)
    case m := <- queueRoom.Incoming:
            handleQueueMessage(m)
    case m := <- GameHandler:
            handleGameMessage(m)
        }
    }
}

func handleGlobalMessage(m *Message) {
    if(m.Type == "QUEUE") {
        GlobalRoom.FindClient(m.Sender).ChangeRoom(queueRoom)
        if len(queueRoom.Clients) >= 8 {
            startGame(queueRoom)
        }
    }
}

func handleQueueMessage(m *Message) {
    log.Println(m)
}

func handleGameMessage(m *Message) {
    if(m.Type == "DONE") {
        gameRoom := gameRooms[m.Sender]
        gameRoom.MoveClients(GlobalRoom)
	delete(gameRooms, gameRoom.ID)	    
    }
}

func startGame(r *socket.Room) {
   gameRoom := addRoom()
   queueRoom.MoveClients(gameRoom)
   game := NewGame(19, gameRoom)
   game.Run()
}

func addRoom() *socket.Room {
    room := socket.NewRoom()
    gameRooms[room.ID] = room
    return room
}

