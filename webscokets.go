package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	*websocket.Conn
	moves          chan string
	disconnectChan chan bool
}

func NewWebsocketClient(connection *websocket.Conn) *WebsocketClient {
	c := WebsocketClient{
		connection,
		make(chan string),
		make(chan bool),
	}

	go func() {
		defer func(Conn *websocket.Conn) {
			err := Conn.Close()

			if err != nil {
				log.Println(err)
			}
		}(c.Conn)

	loop:
		for {
			msgType, buf, err := c.Conn.ReadMessage()

			if err != nil {
				log.Println(err)
				c.disconnectChan <- true
				break
			}

			switch msgType {
			case websocket.CloseMessage:
				c.disconnectChan <- true
				break loop
			case websocket.TextMessage:
				var move MoveJson

				err = json.Unmarshal(buf, &move)

				if err != nil {
					log.Println(err)
					continue
				}

				c.moves <- fmt.Sprintf("%s %s", move.X, move.Y)
			}
		}
	}()

	return &c
}

type MoveJson struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type MessageJson struct {
	Message string `json:"message"`
}

func (w *WebsocketClient) AskForMove() (string, error) {
	select {
	case <- w.disconnectChan:
		return "", errors.New("Client disconnected")
	case move := <-w.moves:
		return move, nil
	}
}

func (w *WebsocketClient) OnGameFinish() {
	close(w.disconnectChan)
	close(w.moves)
	w.Conn.Close()
}

func (w *WebsocketClient) Send(s string) error {
	msg := MessageJson{
		Message: s,
	}

	return w.Conn.WriteJSON(msg)
}

func (w *WebsocketClient) SendBoardState(b *Board) error {
	return w.WriteJSON(b)
}

func (w *WebsocketClient) GetDisconnectChan() chan bool {
	return w.disconnectChan
}

func StartWebsocketServer(c chan Client) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		con, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println(err)
			return
		}

		websocketClient := NewWebsocketClient(con)

		queue.Add(websocketClient)

		c <- websocketClient
	})

	err := http.ListenAndServe("localhost:8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
