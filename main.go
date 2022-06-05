package main

import (
	"log"
)

type Queue struct {
	Items []Client
}

var queue = Queue{
	Items: make([]Client, 0),
}

func (q *Queue) Add(p Client) {
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
	c := make(chan Client)

	go StartTelnetServer(c)

	go StartWebsocketServer(c)

	for {
		<-c

		MakeMatch()
	}
}

func MakeMatch() {
	if len(queue.Items) == 1 {
		err := queue.Items[0].Send("Please wait for other players")

		if err != nil {
			log.Println(err)
		}

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
