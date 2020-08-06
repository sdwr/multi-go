package socket

import (
    "math/rand"
    "time"
    "log"
    "net/http"
    "encoding/json"

    "github.com/gorilla/websocket"

    . "github.com/sdwr/multi-go/types"
)

type Room struct {
    ID int
    Name string
    Settings RoomSettings
    Clients map[*Client]bool

    register chan *Client
    unregister chan *Client
    Incoming chan *Message
    Outgoing chan *Message

    registerCallback func(*Client)
    unregisterCallback func(*Client)
    incomingCallback func(*Message)
}

type RoomSettings struct {
    MaxClients int
    Timeout time.Duration
}

type Client struct {
    room *Room
    connection *websocket.Conn
    send chan []byte
    Name string
    ID int
}

var pingTimer int
var pongDeadline int

var randomSource *rand.Rand
var lastID int

var rooms []*Room
var upgrader websocket.Upgrader

//**************************************************
//HELPER FUNCTIONS
//**************************************************

func GenerateID() int{
    lastID++
    return lastID
}

func findClient(id int, clients map[*Client]bool) *Client {
    for c, _ := range clients {
	if c.ID == id {
	    return c
	}
    }
    return nil
}

//**************************************************
//ROOM FUNCTIONS
//**************************************************

func clientCallback(c *Client) {}

func messageCallback(m *Message){}

func NewRoom() *Room {
	initGlobals()
	return &Room{ID:GenerateID(),
                        Name:"",
                        Settings: CreateRoomSettings(),
                        Clients:make(map[*Client]bool),
                        register:make(chan *Client),
                        unregister:make(chan *Client),
                        Incoming:make(chan *Message, 20),
                        Outgoing:make(chan *Message, 20),
                        registerCallback:clientCallback,
			unregisterCallback:clientCallback,
			incomingCallback:messageCallback,
                }
}

func CreateRoomSettings() RoomSettings {
    return RoomSettings{MaxClients:100,
                                Timeout: 60 * time.Second,}
}

func (r *Room) SetRegisterCallback(cb func(*Client)) {
    r.registerCallback = cb
}

func (r *Room) SetUnregisterCallback(cb func(*Client)) {
    r.unregisterCallback = cb
}

func (r *Room) SetIncomingCallback(cb func(*Message)) {
    r.incomingCallback = cb
}

func (r *Room) BroadcastMessage(m *Message) {
    r.Outgoing <- m
}

func (r *Room) FakeIncomingMessage(m *Message) {
    r.Incoming <- m
}

func (r *Room) sendOutgoing(m *Message) {
    encodedMessage, err := json.Marshal(m)
    if(err != nil) {
        log.Println(err)
    return
    }
    if(m.Reciever != 0) {
        findClient(m.Reciever, r.Clients).send <- encodedMessage
    } else {
        r.sendAll(encodedMessage)
    }
}

func (r *Room) sendAll(encodedMessage []byte) {
    clients := r.Clients
    for c, _ := range clients {
        select {
        case c.send <- encodedMessage:
        default:
                    delete(r.Clients, c)
        }
    }
}

func (r *Room) Run() {
    for {
        select {
        case client := <-r.register:
                r.registerClient(client)
        case client := <-r.unregister:
                if _, ok := r.Clients[client]; ok {
                    r.unregisterClient(client)
                }
        case message := <-r.Incoming:
    		r.incomingCallback(message)
        case message := <-r.Outgoing:
            r.sendOutgoing(message)
        }
    }
}

//Client moving functions

func (r *Room) registerClient(c *Client) {
    r.Clients[c] = true
    r.registerCallback(c)
}

func (r *Room) unregisterClient(c *Client) {
    delete(r.Clients, c)
    r.unregisterCallback(c)
}

func (r *Room) UnregisterAll() {
    for c, _ := range r.Clients {
        r.unregisterClient(c)
    }
}

func (r *Room) MoveClients(newR *Room) {
    for c, _ := range r.Clients {
	c.ChangeRoom(newR)
    }
}

func initGlobals() {
    lastID = 0
    randomSource = rand.New(rand.NewSource(99))
    upgrader = websocket.Upgrader{}
    upgrader.CheckOrigin = func(r *http.Request) bool {return true}
}

//**************************************************
//CLIENT FUNCTIONS
//**************************************************

func (c *Client) readPump() {
    defer func() {
        c.room.unregister <- c
        c.connection.Close()
    }()
    //add ping tests
    for {
        _, message, err := c.connection.ReadMessage()
        if err != nil {
            log.Println(err)
            break
        }
        var decodedMessage Message
        json.Unmarshal(message, &decodedMessage)
        decodedMessage.Sender = c.ID
        c.room.Incoming <- &decodedMessage
    }
}


func (c *Client) writePump() {
    defer func() {
        c.connection.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.connection.WriteMessage(websocket.CloseMessage, []byte{})
                return

            }
            w, err := c.connection.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)
            //add queue support
            if err := w.Close(); err != nil {
                return
            }
        }
    }
}

func (c *Client) ChangeRoom(r *Room) {
    c.room.unregister <- c
    c.room = r
    r.register <- c
}

func ServeWs(room *Room, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
	log.Println(err)
	return
    }
    client := &Client{room: room,
    		      connection: conn,
		      send: make(chan []byte, 10),
		      Name: "",
		      ID: GenerateID(),
	      }
    client.room.register <- client

    go client.writePump()
    go client.readPump()
}
