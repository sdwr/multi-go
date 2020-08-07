package main

import (
    "log"

    "github.com/sdwr/multi-go/logger"
    "github.com/sdwr/multi-go/socket"
    . "github.com/sdwr/multi-go/types"
)

var GlobalRoom *socket.Room
var queueRoom *socket.Room
var gameRooms map[int]*socket.Room
var GameHandler chan *Message

func InitCoordinator() *socket.Room {
    initRooms()
    go run()
    return GlobalRoom
}


func initRooms() {
    GlobalRoom = socket.NewRoom()
    go GlobalRoom.Run()
    queueRoom = socket.NewRoom()
    go queueRoom.Run()
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
	logger.Log(3, "moving client", m.Sender, " to queue")
	logger.Log(3, *queueRoom)
	if len(queueRoom.Clients) >= 4 {
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
    room := r
    queueRoom = socket.NewRoom()
    go queueRoom.Run()
   gameRooms[r.ID] = r
   logger.Log(3, "starting game in room", *room)
   logger.Log(3, "with clients", room.Clients)
   game := NewGame(19, room)
   go game.Run()
}

func addRoom() *socket.Room {
    room := socket.NewRoom()
    gameRooms[room.ID] = room
    return room
}

