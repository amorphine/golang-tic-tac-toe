package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Client interface {
	AskForMove() (string, error)
	OnGameFinish()
	OnBoardStateUpdate(b *Board)
	SendMessage(s string) error
	SendBoardState(b *Board) error
	GetDisconnectChan() chan bool
}

type Symbol uint8

const (
	None Symbol = iota
	Cross
	Circle
)

func (s Symbol) String() string {
	switch s {
	case Cross:
		return "X"
	case Circle:
		return "O"
	}

	return " "
}

type Board struct {
	Size  int        `json:"size"`
	Cells [][]Symbol `json:"cells"`
}

func (b *Board) SetSymbol(s Symbol, x, y int) {
	b.Cells[y][x] = s
}

func (b *Board) GetSymbol(x, y int) Symbol {
	return b.Cells[y][x]
}

func (b *Board) CheckWinner(x, y int) bool {
	s := b.Cells[y][x]

	if s == None {
		return false
	}

	col, row, diag, rdiag := 0, 0, 0, 0

	n := len(b.Cells)

	x, y = y, x

	for i := 0; i < n; i++ {
		if b.Cells[x][i] == s {
			col += 1
		}
		if b.Cells[i][y] == s {
			row += 1
		}
		if b.Cells[i][i] == s {
			diag += 1
		}
		if b.Cells[x][n-i-1] == s {
			rdiag += 1
		}
	}

	if col == n || row == n || diag == n || rdiag == n {
		return true
	}

	return false
}

func (b *Board) CheckAllCellsBusy() bool {
	for y := range b.Cells {
		for x := range b.Cells {
			if b.Cells[y][x] == None {
				return false
			}
		}
	}

	return true
}

type Player struct {
	Client
	Symbol
}

type Game struct {
	Board
	Players            []*Player
	Winner             *Player
	DisconnectedPlayer chan *Player
}

func (g *Game) Broadcast(s string) error {
	for _, p := range g.Players {
		err := p.SendMessage(s)

		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) Finish() {
	if g.Winner != nil {

		_ = g.Broadcast(fmt.Sprintf("Player %s won. Thanks for the game!", g.Winner))

		return
	}

	if g.CheckAllCellsBusy() {
		_ = g.Broadcast("Draw!")
	}
}

func (g *Game) NextPlayer(p *Player) *Player {
	for _, player := range g.Players {
		if player != p {
			return player
		}
	}

	return nil
}

func (g *Game) Move(p *Player) error {
	for _, otherPlayer := range g.Players {
		if otherPlayer == p {
			continue
		}

		err := otherPlayer.SendMessage(fmt.Sprintf("Player %d is thinking", p.Symbol))

		if err != nil {
			return err
		}
	}

	x, y := g.AskForMove(p)

	// check winner after each move
	won := g.CheckWinner(x, y)

	if won {
		g.Winner = p
	}

	return nil
}

func (game *Game) SendBoardState() error {
	for _, p := range game.Players {
		err := p.SendBoardState(&game.Board)

		if err != nil {
			return err
		}
	}

	return nil
}

func (game *Game) Start() error {
	// check enough players
	if len(game.Players) < 2 {
		return errors.New("not enough players")
	}

	// find the first player
	var player *Player

	for _, p := range game.Players {
		if p.Symbol == Cross {
			player = p
		}
	}

	err := game.SendBoardState()

	if err != nil {
		log.Println(err)

		game.Finish()

		return err
	}

	if player == nil {
		player = game.Players[0]
	}

	// start cycle
	for {
		if game.Winner != nil || game.CheckAllCellsBusy() {
			break
		}

		err := game.Move(player)

		if err != nil {
			game.Finish()

			return err
		}
	
		for _, p := range game.Players {
			p.OnBoardStateUpdate(&game.Board)
		}

		player = game.NextPlayer(player)
	}

	// finish game
	game.Finish()

	return nil
}

func (g *Game) AskForMove(player *Player) (x, y int) {
	input, err := player.AskForMove()

	if err != nil {
		log.Println(err)

		g.Finish()

		return
	}

	input = strings.Trim(input, " ")

	arr := strings.Split(input, " ")

	if len(arr) != 2 {
		return g.AskForMove(player)
	}

	x, err = strconv.Atoi(arr[0])

	if err != nil {
		return g.AskForMove(player)
	}

	y, err = strconv.Atoi(arr[1])

	if err != nil {
		return g.AskForMove(player)
	}

	// check bounds
	if x > g.Size-1 || y > g.Size-1 {
		_ = player.SendMessage("Out of bounds")

		g.AskForMove(player)
	}

	if g.GetSymbol(x, y) != None {
		return g.AskForMove(player)
	}

	g.SetSymbol(player.Symbol, x, y)

	_ = player.SendMessage("You move has been accepted")

	return
}

func CreateBoard() Board {
	b := Board{
		Size:  3,
		Cells: make([][]Symbol, 3),
	}

	for y := range b.Cells {
		b.Cells[y] = make([]Symbol, b.Size)
	}

	return b
}

func CreateGame(a, b Client) *Game {
	p1 := &Player{
		Client: a,
		Symbol: Cross,
	}
	p2 := &Player{
		Client: b,
		Symbol: Circle,
	}

	g := Game{
		Board: CreateBoard(),
		Players: []*Player{
			p1,
			p2,
		},
	}

	for _, player := range g.Players {
		_ = player.SendMessage(fmt.Sprintf("You play for %s", player.Symbol))
	}

	return &g
}
