package main

import (
	"time"

	"github.com/sdwr/multi-go/logger"
	"github.com/sdwr/multi-go/socket"
	. "github.com/sdwr/multi-go/types"
)

const COOLDOWN = 5000

var colors []string
var nextColor int

type Game struct {
	Size             int
	Board            [][]int
	LastUpdated      time.Time
	Players          map[int]*Player
	IncomingMessages chan *Message
	OutgoingMessages chan *Message
}

func NewGame(boardSize int, r *socket.Room) *Game {
	initColors()
	return &Game{
		Size:             boardSize,
		Board:            initBoard(boardSize),
		LastUpdated:      time.Now(),
		Players:          initPlayers(r.Clients),
		IncomingMessages: r.Incoming,
		OutgoingMessages: r.Outgoing,
	}
}

func NewPlayer(id int) *Player {
	return &Player{
		ID:        id,
		Name:      "",
		Color:     getNextColor(),
		Cooldown:  0,
		Territory: 0,
		Captures:  0,
	}
}

func initBoard(size int) [][]int {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return board
}

func getNextColor() string {
	color := colors[nextColor]
	nextColor = (nextColor + 1) % len(colors)
	return color
}

func playersToSlice(players map[int]*Player) []Player {
	out := []Player{}
	for _, p := range players {
		out = append(out, *p)
	}
	return out
}

func initPlayers(clients map[int]*socket.Client) map[int]*Player {
	logger.Log(4, "creating player list from client list", clients)
	players := make(map[int]*Player)
	for _, c := range clients {
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
	return pos.X*100 + pos.Y
}

func unhash(hash int) Position {
	return Position{hash / 100, hash % 100}
}

func add(pos1 Position, pos2 Position) Position {
	return Position{X: pos1.X + pos2.X, Y: pos1.Y + pos2.Y}
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
	return Position{X: -1, Y: -1}
}

func getSurrounding(pos Position) []Position {
	surround := []Position{Position{X: 1, Y: 0}, Position{X: -1, Y: 0}, Position{X: 0, Y: 1}, Position{X: 0, Y: -1}}
	for i, p := range surround {
		surround[i] = add(p, pos)
	}
	return surround
}

func addSurrounding(stack *[]Position, pos Position, groupBoard map[int]bool) {
	ar := *stack
	surround := getSurrounding(pos)
	for _, p := range surround {
		if groupBoard[hash(p)] == false {
			ar = append(ar, p)
		}
	}
	*stack = ar
}

func (game *Game) groupLives(pos Position) bool {
	playerID := game.getSpace(pos)
	if playerID == -1 || playerID == 0 {
		return true
	}
	empty := 0
	groupBoard := make(map[int]bool)
	stack := &[]Position{}
	push(stack, pos)
	groupBoard[hash(pos)] = true

	for len(*stack) > 0 {
		pos = pop(stack)
		switch square := game.getSpace(pos); square {
		case 0:
			empty++
		case playerID:
			groupBoard[hash(pos)] = true
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
	if playerID == -1 || playerID == 0 {
		return group
	}
	groupBoard := make(map[int]bool)
	stack := &[]Position{}
	push(stack, pos)
	groupBoard[hash(pos)] = true

	for len(*stack) > 0 {
		pos = pop(stack)
		switch square := game.getSpace(pos); square {
		case playerID:
			groupBoard[hash(pos)] = true
			addSurrounding(stack, pos, groupBoard)
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
		Type: "UPDATE",
		Payload: Payload{
			Move:   move,
			Remove: []Position{},
		},
	}
	if !game.spaceClear(move.Coords) {
		return
	}
	if game.Players[move.Player.ID].Cooldown > 0 {
		return
	}
	//add move
	game.Board[move.Coords.X][move.Coords.Y] = move.Player.ID
	//check surrounding groups
	//some specific logic here, dont change w/o making better
	alreadyRemoved := []Position{}
	for _, pos := range getSurrounding(move.Coords) {
		if !game.groupLives(pos) && !(game.Board[pos.X][pos.Y] == move.Player.ID) && !game.sameGroup(pos, alreadyRemoved) {
			alreadyRemoved = append(alreadyRemoved, pos)
			outMessage.Payload.Remove = append(outMessage.Payload.Remove, game.findGroup(pos)...)
			game.addCaptures(move.Player.ID, pos)
		}
	}

	if len(outMessage.Payload.Remove) > 0 || game.groupLives(move.Coords) {
		game.Players[move.Player.ID].Territory++
		game.Players[move.Player.ID].Cooldown = COOLDOWN
		game.removePieces(outMessage.Payload.Remove)
		outMessage.Payload.Players = playersToSlice(game.Players)
		game.sendMoveMessage(&outMessage)
	} else {
		game.Board[move.Coords.X][move.Coords.Y] = 0
	}
}

func (game *Game) sameGroup(p Position, s []Position) bool {
	group := game.findGroup(p)
	for _, otherP := range s {
		for _, groupP := range group {
			if otherP.X == groupP.X && otherP.Y == groupP.Y {
				return true
			}
		}
	}
	return false
}

func (game *Game) addCaptures(id int, pos Position) {
	amt := len(game.findGroup(pos))
	game.Players[id].Captures += amt
	game.Players[game.Board[pos.X][pos.Y]].Territory -= amt
}

func (game *Game) addTerritory(id int) {
	game.Players[id].Territory += 1
}

func (game *Game) updateTimers() {
	currTime := time.Now()
	elapsed := currTime.Sub(game.LastUpdated)
	elapsedMillis := elapsed.Milliseconds()
	for _, p := range game.Players {
		if p.Cooldown > 0 {
			p.Cooldown -= int(elapsedMillis)
		}
	}

	game.LastUpdated = currTime
}

//message without reciever goes to all clients
func (game *Game) sendMoveMessage(m *Message) {
	game.OutgoingMessages <- m
}

func (game *Game) sendInitMessage() {
	for _, p := range game.Players {
		game.OutgoingMessages <- game.createInitMessage(p)
	}
}

func (game *Game) createInitMessage(p *Player) *Message {
	return &Message{
		Reciever: p.ID,
		Type:     "GAMESTART",
		Payload:  Payload{Player: *p, Players: playersToSlice(game.Players)},
	}
}

func (game *Game) processMessage(m *Message) {
	game.updateTimers()
	move := m.Payload.Move
	player := &move.Player
	if game.Players[player.ID] == nil {
		game.Players[player.ID] = player
	}
	logger.Log(3, "adding piece")
	game.addPiece(move)
}

func initColors() {
	nextColor = 0
	colors = []string{"#657287", "#7ced2b", "#2032f7", "#ace6e0", "#610445", "#2dad21", "#fce428", "#000000", "#8c6c6c"}

}

func (game *Game) Run() {
	logger.Log(3, "starting game", *game)
	game.sendInitMessage()
	for {
		message, _ := <-game.IncomingMessages
		game.processMessage(message)
	}
}
