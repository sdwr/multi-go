package main

import (
    . "github.com/sdwr/multi-go/types"
)

var game Game
func NewGame(boardSize int, r *Room) *Game {
        return &Game{
                Size:boardSize,
                Board:initBoard(boardSize),
		Players: initPlayers(r.Clients),
    		IncomingMessages: r.Incoming,
		OutoingMessages: r.Outgoing,
	}
}

func NewPlayer(id int) *Player {
    return &Player{
	ID: id,
	Name: "",
	Color: randomColor(),
	Cooldown: 10000
    }
}

func randomColor() {
    return "#abcabc"
}

func initBoard(size) Board {
        board := make([][]int, size)
    for i := range board {
        board[i] = make([]int, size)
    }
    return board
}

func initPlayers(clients map[*Client]bool) map[int]*Player {
    players := make(map[int]*Player)
    for c, _ := range clients {
	players[c.ID] = NewPlayer(c.ID)
    }
    return players
}

func removePieces(s []Position) {
    for _, pos := s {
        game.Board[s.X][s.Y] = 0
    }
}

func getSpace(pos Position) int {
    if pos.X >= 0 && pos.X < game.Size && pos.Y >= 0 && pos.Y < game.Size {
        return game.Board[pos.X][pos.Y]
    }
    return -1
}

func spaceClear(pos Position) bool {
    return game.Board[pos.X][pos.Y] == 0
}

func hash(pos Position) int {
    return pos.X * 100 + pos.Y
}

func unhash(hash int) Position {
    return Position{hash/100,hash%100}
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

func findGroup(pos Position) []Position {
	group := []Position{}
    playerID := getSpace(pos)
    if(playerID == -1 || playerID == 0) {
        return group
    }
    groupBoard := make(map[int]bool)
    stack := &[]Position{}
    stack.push(pos)
    groupBoard[hash(pos)]=true

    while len(stack) > 0 {
        pos = stack.pop()
        square = getSpace(pos)
        select {
        case square == playerID:
            groupBoard[hash(pos)]=true
            stack.addSurrounding(pos)
        default:
            break
        }
    }
    for h, _ := range groupBoard {
        group = append(group, unhash(h))
    }
    return group
} 

func addPiece(move Move) {
	outMessage := Message{
		Type:"UPDATE",
		Payload:Payload{
			Move: move,
			Remove: []Position{}
		}
    if !spaceClear(move.Coords) {
	    return
    }
    //add move
    game.Board[move.Coords.X][move.Coords.Y] == move.Player.ID
    //check surrounding groups
    for _, pos := range getSurrounding(move.Coords) {
	    if(!groupLives(pos)) {
		    outMessage.Payload.Remove = append(outMessage.Payload.Remove, findGroup(pos))
	    }
    }

    if len(outMessage.Remove) > 0 || groupLives(move.Coords) {
        removePieces(outMessage.Remove)
    	outgoingMessages <-move
    } else {
	    game.Board[move.Coords.X][move.Coords.Y] == 0
    }

}

func sendInitMessage() {
    for _, p := game.Players {
        game.Outgoing <- createInitMessage(p)
    }
}

func createInitMessage(p *Player) *Message {
        return &Message{
                Reciever:p.ID,
                Type:"GAMESTART"
                Payload:Payload{Player:*p}
        }
}

func processMessage(m *Message) {
    move := m.Payload.Move
    player := &move.Player
	if game.Players[player.ID] == nil {
	    game.Players[player.ID] = player
	}
	addPiece(move)
}

func Run() {
    sendInitMessage()
    for {
	    message, _ := <-incomingMessages:
	processMessage(message)
	}
}

