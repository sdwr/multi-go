package types

type State struct {
    Size int
    Board [][]int
    Players map[int]Player
}

type Player struct {
    ID int
    Name string
    Color string
    Cooldown int
}

type Payload struct {
    Move Move
}

type Move struct {
    Coords Position
    Player Player
}

type Position struct {
    X int
    Y int
}
