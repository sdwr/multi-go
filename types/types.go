package types

type State struct {
    Size int
    Board [][]int
    Players map[int]Player
}

type Move struct {
    Coords Position
    Player Player
}

type Player struct {
    ID int
    Name string
    Color string
}

type Position struct {
    X int
    Y int
}

type Payload struct {
    Move Move
}

