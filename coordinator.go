package main

import (
    "log"
    "time"

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
    initQueueRoom()
    gameRooms = make(map[int]*socket.Room)
    GameHandler = make(chan *Message, 10)
}

func initQueueRoom() {
    queueRoom = socket.NewRoom()
    go queueRoom.Run()
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
	sendQueueUpdateMessage(queueRoom)
	logger.Log(3, "moving client", m.Sender, " to queue")
	logger.Log(3, *queueRoom)
	if len(queueRoom.Clients) >= 4 {
           go startGame(queueRoom)
        }
    }
}

func handleQueueMessage(m *Message) {
    if(m.Type == "LEAVEQUEUE") {
	queueRoom.FindClient(m.Sender).ChangeRoom(GlobalRoom)
	sendQueueUpdateMessage(queueRoom)
	logger.Log(3, "moving client", m.Sender, " out of queue")
    }
	log.Println(m)
}

func sendQueueUpdateMessage(r *socket.Room) {
	m := Message{}
	m.Type = "QUEUEUPDATE"
	m.Payload.Queued = len(r.Clients)
	r.Outgoing <- &m
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
   initQueueRoom()
   gameRooms[r.ID] = r
   time.Sleep(2000)
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

