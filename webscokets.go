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
	movesChan          chan string
	disconnectChan chan bool
}

func NewWebsocketClient(connection *websocket.Conn) *WebsocketClient {
	client := WebsocketClient{
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
		}(client.Conn)

	loop:
		for {
			msgType, buf, err := client.Conn.ReadMessage()

			if err != nil {
				log.Println(err)
				client.disconnectChan <- true
				break
			}

			switch msgType {
			case websocket.CloseMessage:
				client.disconnectChan <- true
				break loop
			case websocket.TextMessage:
				var move MoveJson

				err = json.Unmarshal(buf, &move)

				if err != nil {
					log.Println(err)
					continue
				}

				client.movesChan <- fmt.Sprintf("%s %s", move.X, move.Y)
			}
		}
	}()

	return &client
}

type MoveJson struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type MessageJson struct {
	Message string `json:"message"`
}

func (client *WebsocketClient) AskForMove() (string, error) {
	err := client.SendMessage("Your turn")

	if err != nil {
		return "", err
	}

	select {
	case <- client.disconnectChan:
		return "", errors.New("Client disconnected")
	case move := <-client.movesChan:
		return move, nil
	}
}

func (c *WebsocketClient) OnBoardStateUpdate(board *Board) {
	err := c.SendBoardState(board)

	if err != nil {
		c.disconnectChan <- true
	}
}

func (w *WebsocketClient) OnGameFinish() {
	close(w.disconnectChan)
	close(w.movesChan)
	w.Conn.Close()
}

func (w *WebsocketClient) SendMessage(s string) error {
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
		log.Println("incoming websocket connection")

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
