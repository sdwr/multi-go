package socket

import (
    "math/rand"
    "time"
    "log"
    "net/http"
    "encoding/json"

    "github.com/gorilla/websocket"

    "github.com/sdwr/multi-go/logger"
    . "github.com/sdwr/multi-go/types"
)

//DATA STRUCTURES:
// Room.Clients is modified on register/unregister
// all room channels are read in Room.Run()

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Room struct {
    ID int
    Name string
    Settings RoomSettings
    Clients map[int]*Client

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
                        Clients:make(map[int]*Client),
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
    logger.Log(4, "sending outgoing")
    logger.Log(4, m)
    encodedMessage, err := json.Marshal(m)
    if(err != nil) {
        log.Println(err)
    return
    }
    if(m.Reciever != 0) {
        r.FindClient(m.Reciever).send <- encodedMessage
    } else {
        r.sendAll(encodedMessage)
    }
}

func (r *Room) sendAll(encodedMessage []byte) {
    clients := r.Clients
    for _, c := range clients {
        select {
        case c.send <- encodedMessage:
        default:
                    delete(r.Clients, c.ID)
        }
    }
}

func (r *Room) Run() {
    for {
        select {
        case client := <-r.register:
                r.registerClient(client)
        case client := <-r.unregister:
                if _, ok := r.Clients[client.ID]; ok {
                    r.unregisterClient(client)
                }
        case message := <-r.Incoming:
    		r.incomingCallback(message)
        case message := <-r.Outgoing:
            r.sendOutgoing(message)
        }
    }
}

//Client helper functions
func (r *Room) FindClient(id int) *Client {
    return r.Clients[id]
}

//Client moving functions
func (r *Room) registerClient(c *Client) {
    logger.Log(4, "registerClient ", c, " in room ", r)
    c.room = r
    r.Clients[c.ID] = c
    sendInitMessage(r, c)
    r.registerCallback(c)
}

func (r *Room) unregisterClient(c *Client) {
    logger.Log(4, "unregisterClient ", c, " in room ", r)
    delete(r.Clients, c.ID)
    r.unregisterCallback(c)
}

func (r *Room) UnregisterAll() {
    for _, c := range r.Clients {
        r.unregisterClient(c)
    }
}

func sendInitMessage(r *Room, c *Client) {
    m := createInitMessage(r, c)
    r.sendOutgoing(m)
}

func createInitMessage(r *Room, c *Client) *Message {
    m := Message{}
    m.Type = "INIT"
    m.Sender = c.ID
    return &m
}


//NOT THREAD SAFE
func (r *Room) MoveClients(newR *Room) {
	clients := r.Clients
    for _, c := range clients {
	c.ChangeRoom(newR)
    }
}

func initGlobals() {
    if(lastID == 0){
        lastID = 1
        randomSource = rand.New(rand.NewSource(99))
        upgrader = websocket.Upgrader{}
        upgrader.CheckOrigin = func(r *http.Request) bool {return true}
    }
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
    c.connection.SetReadDeadline(time.Now().Add(pongWait))
    c.connection.SetPongHandler(func(string) error { c.connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })
    for {
        _, message, err := c.connection.ReadMessage()
        if err != nil {
            log.Println(err)
            break
        }
        var decodedMessage Message
        json.Unmarshal(message, &decodedMessage)
        decodedMessage.Sender = c.ID
	logger.Log(3, "recieving message from client", c)
	logger.Log(3, decodedMessage)
        c.room.Incoming <- &decodedMessage
    }
}


func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
	ticker.Stop()
        c.connection.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
 		c.connection.SetWriteDeadline(time.Now().Add(writeWait))
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
	case <- ticker.C:
		c.connection.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
        }
    }
}

func (c *Client) ChangeRoom(r *Room) *Client {
    client := c
    logger.Log(4, "changing client", client, " to room ", r)
    c.room.unregister <- c
    r.register <- client
    return c
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
    logger.Log(3, "registered client", *client)

    go client.writePump()
    go client.readPump()
}
