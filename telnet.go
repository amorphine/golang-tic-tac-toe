package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type TelnetClient struct {
	Connection net.Conn
	disconnectChan chan bool
}

func (c *TelnetClient) GetDisconnectChan() chan bool {
	return c.disconnectChan
}

func (c *TelnetClient) Send(s string) error {
	con := c.Connection

	_, err := con.Write([]byte(s + "\n\r"))

	if err != nil {
		log.Fatal("TelnetClient lost connection")
	}

	log.Println(fmt.Sprintf("Message sent: %s", s))

	return err
}

func PrintBoard(b *Board) string {
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

func (c *TelnetClient) SendBoardState(b *Board) error {
	err := c.Send(PrintBoard(b))

	if err != nil {
		log.Fatal("TelnetClient lost connection")
	}

	return err
}

func (c *TelnetClient) AskForMove() (string, error) {
	err := c.Send("Your turn")

	if err != nil {
		return "", err
	}

	input := make([]byte, 1)

	var sb strings.Builder

	for {
		_, err = c.Connection.Read(input)

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

func (c *TelnetClient) OnGameFinish() {
	_ = c.Connection.Close()
}

func StartTelnetServer(c chan Client) {
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
				disconnectChan: make(chan bool),
			}

			queue.Add(&player)

			c <- &player
		}()
	}
}
