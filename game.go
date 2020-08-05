package main

import (
    . "github.com/sdwr/multi-go/types"
)

var state State
var incomingMessages chan *socket.Message
func InitGame(boardSize int) {
        state := State{
                Size:boardSize,
                Board:initBoard(boardSize),
                Players: make(map[int]Player),
        }
}

func initBoard(size) Board {
        board := make([][]int, size)
    for i := range board {
        board[i] = make([]int, size)
    }
    return board
}

func Run() {
    
}
