package main

import (
    "github.com/sdwr/multi-go/socket"
    . "github.com/sdwr/multi-go/types"
)

type Game struct {
    Size int
    Board [][]int
    Players map[int]*Player
    IncomingMessages chan *Message
    OutgoingMessages chan *Message
}

func NewGame(boardSize int, r *socket.Room) *Game {
        return &Game{
                Size:boardSize,
                Board:initBoard(boardSize),
		Players: initPlayers(r.Clients),
    		IncomingMessages: r.Incoming,
		OutgoingMessages: r.Outgoing,
	}
}

func NewPlayer(id int) *Player {
    return &Player{
	ID: id,
	Name: "",
	Color: randomColor(),
	Cooldown: 10000,
    }
}

func randomColor() string {
    return "#abcabc"
}

func initBoard(size int) [][]int {
        board := make([][]int, size)
    for i := range board {
        board[i] = make([]int, size)
    }
    return board
}

func initPlayers(clients map[int]*socket.Client) map[int]*Player {
    players := make(map[int]*Player)
    for _ , c := range clients {
	players[c.ID] = NewPlayer(c.ID)
    }
    return players
}

func (game *Game) removePieces(s []Position) {
    for _, pos := range s {
        game.Board[pos.X][pos.Y] = 0
    }
}

func (game *Game) getSpace(pos Position) int {
    if pos.X >= 0 && pos.X < game.Size && pos.Y >= 0 && pos.Y < game.Size {
        return game.Board[pos.X][pos.Y]
    }
    return -1
}

func (game *Game) spaceClear(pos Position) bool {
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

func push(stack *[]Position, pos Position) {
    *stack = append(*stack, pos)
}

func pop(stack *[]Position) Position {
    ar := *stack
    if len(ar) > 0 {
        pos := ar[len(ar)-1]
        *stack = ar[:len(ar)-1]
        return pos
    }
    return Position{X:-1,Y:-1}
}

func getSurrounding(pos Position) []Position {
    surround := []Position{Position{X:1,Y:0},Position{X:-1,Y:0},Position{X:0,Y:1},Position{X:0,Y:-1}}
    for i, p := range surround {
        surround[i] = add(p, pos)
    }
    return surround
}

func addSurrounding(stack *[]Position, pos Position, groupBoard map[int]bool) {
    ar := *stack
    surround := getSurrounding(pos)
    for _, p := range surround {
        if(groupBoard[hash(p)] == false) {
            ar = append(ar, p)
        }
    }
    *stack = ar
}

func (game *Game) groupLives(pos Position) bool {
    playerID := game.getSpace(pos)
    if(playerID == -1 || playerID == 0) {
        return true
    }
    empty := 0
    groupBoard := make(map[int]bool)
    stack := &[]Position{}
    push(stack, pos)
    groupBoard[hash(pos)]=true

    for ;len(*stack) > 0; {
        pos = pop(stack)
	switch square := game.getSpace(pos); square {
        case 0:
            empty++
        case playerID:
            groupBoard[hash(pos)]=true
            addSurrounding(stack, pos, groupBoard)
        default:
            break
        }
    }
    return empty > 0
}

func (game *Game) findGroup(pos Position) []Position {
	group := []Position{}
    playerID := game.getSpace(pos)
    if(playerID == -1 || playerID == 0) {
        return group
    }
    groupBoard := make(map[int]bool)
    stack := &[]Position{}
    push(stack, pos)
    groupBoard[hash(pos)]=true

    for ;len(*stack) > 0; {
        pos = pop(stack)
	switch square := game.getSpace(pos); square {
        case playerID:
            groupBoard[hash(pos)]=true
            addSurrounding(stack,pos,groupBoard)
        default:
            break
        }
    }
    for h, _ := range groupBoard {
        group = append(group, unhash(h))
    }
    return group
} 

func (game *Game) addPiece(move Move) {
	outMessage := Message{
		Type:"UPDATE",
		Payload:Payload{
			Move: move,
			Remove: []Position{},
		},
	}
    if !game.spaceClear(move.Coords) {
	    return
    }
    //add move
    game.Board[move.Coords.X][move.Coords.Y] = move.Player.ID
    //check surrounding groups
    for _, pos := range getSurrounding(move.Coords) {
	    if(!game.groupLives(pos)) {
		    outMessage.Payload.Remove = append(outMessage.Payload.Remove, game.findGroup(pos)...)
	    }
    }

    if len(outMessage.Payload.Remove) > 0 || game.groupLives(move.Coords) {
        game.removePieces(outMessage.Payload.Remove)
    	game.sendMoveMessage(&outMessage)
    } else {
	    game.Board[move.Coords.X][move.Coords.Y] = 0
    }

}
//message without reciever goes to all clients
func (game *Game) sendMoveMessage(m *Message) {
    game.OutgoingMessages <- m
}

func (game *Game) sendInitMessage() {
    for _, p := range game.Players {
        game.OutgoingMessages <- createInitMessage(p)
    }
}

func createInitMessage(p *Player) *Message {
        return &Message{
                Reciever:p.ID,
                Type:"GAMESTART",
                Payload:Payload{Player:*p},
        }
}

func (game *Game) processMessage(m *Message) {
    move := m.Payload.Move
    player := &move.Player
	if game.Players[player.ID] == nil {
	    game.Players[player.ID] = player
	}
	game.addPiece(move)
}

func (game *Game) Run() {
    game.sendInitMessage()
    for {
        message, _ := <-game.IncomingMessages
	game.processMessage(message)
	}
}

