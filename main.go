package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type TelnetClient struct {
	Connection net.Conn
}

func (p *TelnetClient) Send(s string) error {
	con := p.Connection

	_, err := con.Write([]byte(s + "\n\r"))

	if err != nil {
		log.Fatal("TelnetClient lost connection")
	}

	log.Println(fmt.Sprintf("Message sent: %s", s))

	return err
}

func (p *TelnetClient) Prompt(msg string) (string, error) {
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

func (p *TelnetClient) OnGameFinish() {
	_ = p.Connection.Close()
}

type Queue struct {
	Items []Client
}

var queue = Queue{
	Items: make([]Client, 0),
}

func (q *Queue) Add(p *TelnetClient) {
	q.Items = append(q.Items, p)
}

func (q *Queue) Pop() (p Client) {
	if len(q.Items) > 0 {
		p = q.Items[0]

		q.Items = q.Items[1:]
	}

	return
}

func main() {
	c := make(chan *TelnetClient)

	go StartTelnetServer(c)

	for {
		<-c

		MakeMatch()
	}
}

func MakeMatch() {
	if len(queue.Items) == 1 {
		_ = queue.Items[0].Send("Please wait for other players")

		return
	}

	if len(queue.Items) < 2 {
		return
	}

	go func() {
		c1, c2 := queue.Pop(), queue.Pop()

		StartGame(c1, c2)

		c1.OnGameFinish()

		c2.OnGameFinish()
	}()
}

func StartGame(a, b Client) {
	log.Println("Start game called")

	g := CreateGame(a, b)

	log.Println("Game created")

	err := g.Start()

	if err != nil {
		log.Println(err)

		return
	}

	log.Println("Game finished")
}

func StartTelnetServer(c chan *TelnetClient) {
	log.Print("Making listener")

	listener, err := net.Listen("tcp", ":5555")

	defer func() {
		_ = listener.Close()
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
			player := TelnetClient{
				Connection: conn,
			}

			queue.Add(&player)

			c <- &player
		}()
	}
}
