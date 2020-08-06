package main

import (
    "github.com/sdwr/multi-go/socket"
    . "github.com/sdwr/multi-go/types"
)

var GlobalRoom *Room
var queueRoom *Room
var gameRooms map[int]*Room
var GameHandler chan *Message

func InitCoordinator() *Room {
    initRooms()
    return GlobalRoom
}	

func RunCoordinator() {
    run()
}

func initRooms() {
    GlobalRoom = socket.NewRoom()
    queueRoom = socket.NewRoom()
    gameRooms = make(map[int]*Room)
    GameHandler = make(chan *Message, 10)
}

func run() {
    for {
        select {
        case m <- GlobalRoom.Incoming:
            handleGlobalMessage(m)
        case m <- queueRoom.Incoming:
            handleQueueMessage(m)
        case m <- GameHandler:
            handleGameMessage(m)
        }
    }
}

func handleGlobalMessage(m *Message) {
    if(m.Type == "QUEUE") {
        m.Sender.ChangeRoom(queueRoom)
        if len(queueRoom.Clients) >= 8 {
            startGame(queueRoom)
        }
    }
}

func handleQueueMessage(m *Message) {
    Log.Println(m)
}

func handleGameMessage(m *Message) {
    if(m.Type == "DONE") {
        gameRoom := gameRooms[m.Sender.ID]
        gameRoom.MoveClients(GlobalRoom)
        delete(gameRoom)
    }
}

func startGame(r *Room) {
   gameRoom := addRoom()
   moveClients(queueRoom, gameRoom)
   game := NewGame(gameRoom)
   game.Run()
}

func addRoom() *socket.Room {
    room := NewRoom()
    gameRooms[room.ID] = room
    return room
}

