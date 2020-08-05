package main

import (
    . "github.com/sdwr/multi-go/types"
)

var state State
var incomingMessages chan *Message
var outgoingMessages chan *Message
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

func getSpace(pos Position) int {
    if pos.X >= 0 && pos.X < state.Size && pos.Y >= 0 && pos.Y < state.Size {
        return board[pos.X][pos.Y]
    }
    return -1
}

func spaceClear(pos Position) bool {
    return board[pos.X][pos.Y] == 0
}

func hash(pos Position) int {
    return pos.X * 100 + pos.Y
}

func add(pos1 Position, pos2 Position) Position {
    return Position{X:pos1.X+pos2.X, Y:pos1.Y+pos2.Y}
}

func (stack *[]Position) push(pos Position) {
    *stack = append(*stack, pos)
}

func (stack *[]Position) pop() Position {
    ar := *stack
    if len(ar) > 0 {
        pos := ar[len(stack)-1]
        *stack = ar[:len(stack)-1]
        return pos
    }
    return nil
}

func getSurrounding(pos Position) []Position {
    surround := [Position{X:1,Y:0},Position{X:-1,Y:0},Position{X:0,Y:1},Position{X:0,Y:-1}
    for i, p := range surround {
        surround[i] = add(p, pos)
    }
    return surround
}

func (stack *[]Position) addSurrounding(pos Position, groupBoard map[int]bool) {
    ar := *stack
    surround := getSurrounding(pos)
    for _, p := range surround {
        if(groupBoard[hash(p)] == false) {
            ar = ar.append(p)
        }
    }
    *stack = ar
}

func groupLives(pos Position) bool {
    playerID = getSpace(pos)
    if(playerID == -1 || playerID == 0) {
        return true
    }
    empty := 0
    groupBoard := make(map[int]bool)
    stack := &[]Position{}
    stack.push(pos)
    groupBoard[hash(pos)]=true

    while len(stack) > 0 {
        pos = stack.pop()
        square = getSpace(pos)
        select {
        case square == 0:
            empty++
        case square == playerID:
            groupBoard[hash(pos)]=true
            stack.addSurrounding(pos)
        default:
            break
        }
    }
    return empty > 0
}

func addPiece(move Move) {
    if !spaceClear(move.Coords) {
    	return
    }
    board[pos.X][pos.Y] == move.Player.ID
    
    outgoingMessages <-move
    
}

func processMessage(m *Message) {
        move := m.Payload.Move
        player := move.Player
	if state.Players[player.ID] == nil {
	    state.Players[player.ID] = player
	    addPiece(move)
	}
}

func sendState() {
    
}

func Run() {
   for {
	message, _ := <-incomingMessages:
	processMessage(message)
	}
   }
}
