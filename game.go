package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Client interface {
	AskForMove() (string, error)
	OnGameFinish()
	Send(s string) error
	SendBoardState(b *Board) error
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
	Size  int
	Cells [][]Symbol
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
	Players []*Player
	Winner  *Player
}

func (g *Game) Broadcast(s string) error {
	for _, p := range g.Players {
		err := p.Send(s)

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

		err := otherPlayer.Send(fmt.Sprintf("Player %d is thinking", p.Symbol))

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

func (g *Game) Start() error {
	// check enough players
	if len(g.Players) < 2 {
		return errors.New("not enough players")
	}

	// find the first player
	var player *Player

	for _, p := range g.Players {
		if p.Symbol == Cross {
			player = p
		}
	}

	if player == nil {
		player = g.Players[0]
	}

	// start cycle
	for {
		if g.Winner != nil || g.CheckAllCellsBusy() {
			break
		}

		err := g.Move(player)

		if err != nil {
			return err
		}

		player = g.NextPlayer(player)
	}

	// finish game
	g.Finish()

	return nil
}

func (g *Game) AskForMove(player *Player) (x, y int) {
	prompt, _ := player.AskForMove()

	prompt = strings.Trim(prompt, " ")

	arr := strings.Split(prompt, " ")

	if len(arr) != 2 {
		return g.AskForMove(player)
	}

	x, err := strconv.Atoi(arr[0])

	if err != nil {
		return g.AskForMove(player)
	}

	y, err = strconv.Atoi(arr[1])

	if err != nil {
		return g.AskForMove(player)
	}

	// check bounds
	if x > g.Size-1 || y > g.Size-1 {
		_ = player.Send("Out of bounds")

		g.AskForMove(player)
	}

	if g.GetSymbol(x, y) != None {
		return g.AskForMove(player)
	}

	g.SetSymbol(player.Symbol, x, y)

	_ = player.Send("You move has been accepted")

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
	g := Game{
		Board: CreateBoard(),
		Players: []*Player{
			{
				Client: a,
				Symbol: Cross,
			}, {
				Client: b,
				Symbol: Circle,
			},
		},
	}

	for _, player := range g.Players {
		_ = player.Send(fmt.Sprintf("You play for %s", player.Symbol))
	}

	return &g
}
