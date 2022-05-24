package main

import (
	"fmt"
	"strconv"
	"strings"
)

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

func (b *Board) Print()string {
	r := ""

	cells := b.Cells

	r += " "

	for k := range cells {
		r += fmt.Sprintf("|%d", k)
	}

	r += "|\r\n"

	r += "--------\r\n"

	for i := range cells {
		r += fmt.Sprintf("%d|", i)
		for k, j := range cells[i] {
			r += j.String()

			if k < len(cells[i]) {
				r += "|"
			}
		}

		r += "\r\n"

		r += "--------\r\n"
	}

	return r	
}

type Player struct {
	*Client
	Symbol
}

type Game struct {
	Board
	Players []*Player
	Winner *Player
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
	g.Broadcast(fmt.Sprintf("Player %s won. Thanks for the game!", g.Winner))

	for _, p := range g.Players {
		p.Connection.Close()
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

func (g *Game) Move(p *Player) {
	for _, otherPlayer := range g.Players {
		if otherPlayer == p {
			continue
		}

		otherPlayer.Send(fmt.Sprintf("Player %d is thinking", p.Symbol))
	}

	x, y := g.AskForMove(p)
	
	// check winner after each move
	won := g.CheckWinner(x, y)

	if won {
		g.Winner = p

		g.Finish()

		return
	}
		
	g.Move(g.NextPlayer(p))
}

func (g *Game) AskForMove(player *Player) (x, y int) {
	player.Send(g.Print())
	
	prompt, _ := player.Prompt("Your turn")

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
		player.Send("Out of bounds")
		g.AskForMove(player)
	}

	if g.GetSymbol(x, y) != None {
		return g.AskForMove(player)
	}

	g.SetSymbol(player.Symbol, x, y)

	player.Send("You move has been accepted")

	player.Send(g.Print())

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

func CreateGame(a, b *Client) *Game {
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
		player.Send(fmt.Sprintf("You play for %s", player.Symbol))
	}

	return &g
}
