package types

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
