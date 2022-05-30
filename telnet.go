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
