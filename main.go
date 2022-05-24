package main

import (
	"log"
	"net"
	"strings"
)

type Client struct {
	Connection net.Conn
}

func (p *Client) Send(s string) error {
	con := p.Connection

	_, err := con.Write([]byte(s + "\n\r"))

	if err != nil {
		log.Fatal("Client lost connection")
	}

	return err
}

func (p *Client) Prompt(msg string) (string, error) {
	err := p.Send(msg)

	if err != nil {
		return "", err
	}

	input := make([]byte, 1)

	var sb strings.Builder

	for {
		_, err = p.Connection.Read(input)

		if err != nil {
			return "", err
		}

		stringInput := string(input)

		if stringInput == "\r" || stringInput == "\n" {
			return sb.String(), nil
		}

		sb.WriteString(stringInput)
	}
}

type Queue struct {
	Items []*Client
}

var queue = Queue{
	Items: make([]*Client, 0),
}

func (q *Queue) Add(p *Client) {
	q.Items = append(q.Items, p)
}

func (q *Queue) Pop() (p *Client) {
	if len(q.Items) > 0 {
		p = q.Items[0]

		q.Items = q.Items[1:]
	}

	return
}

func main() {
	c := make(chan *Client)

	go StartTelnetServer(c)

	for {
		<-c

		MakeMatch()
	}
}

func MakeMatch() {
	if len(queue.Items) == 1 {
		queue.Items[0].Send("Please wait for other players")

		return
	}

	if len(queue.Items) >= 2 {
		go StartGame(queue.Pop(), queue.Pop())
	}
}

func StartGame(a, b *Client) {
	log.Println("Start game called")

	g := CreateGame(a, b)

	log.Println("Game created")

	g.Move(g.Players[0])
}

func StartTelnetServer(c chan *Client) {
	log.Print("Making listener")

	listener, err := net.Listen("tcp", ":5555")

	defer func() {
		listener.Close()
	}()

	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			continue
		}

		log.Print("New player joined")

		go func() {
			player := Client{
				Connection: conn,
			}

			player.Send("Hello there\n")

			queue.Add(&player)

			c <- &player
		}()
	}
}
