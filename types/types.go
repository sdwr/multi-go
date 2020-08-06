package types

type Game struct {
    Size int
    Board [][]int
    Players map[int]*Player
    IncomingMessages chan *Message
    OutgoingMessages chan *Message
}

type Player struct {
    ID int
    Name string
    Color string
    Cooldown int
}

type Message struct {
    Sender int
    Reciever int
    Type string
    Payload Payload
}

type Payload struct {
    Player Player
    Move Move
    Remove []Position
}

type Move struct {
    Coords Position
    Player Player
}

type Position struct {
    X int
    Y int
}
