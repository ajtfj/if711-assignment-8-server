package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	MAX_CLIENT = 5
)

type SushiBarRPC struct {
	Capacity   int
	clients    map[int]*Client
	lastTicket int
	wg         sync.WaitGroup
	mu         sync.Mutex
}

func NewSushiBarRPC(capacity int) *SushiBarRPC {
	return &SushiBarRPC{
		Capacity:   capacity,
		clients:    make(map[int]*Client),
		lastTicket: 0,
		wg:         sync.WaitGroup{},
		mu:         sync.Mutex{},
	}
}

func (sb *SushiBarRPC) Enter(args *SushiBarEnterArgs, reply *SushiBarEnterReply) error {
	requestTime := time.Now()
	fmt.Printf("Client %s is waiting to enter the sushi bar\n", args.Client.Name)

	sb.mu.Lock()
	if len(sb.clients) == sb.Capacity {
		sb.wg.Wait()
	}
	sb.wg.Add(1)

	timeWaiting := time.Since(requestTime)

	ticket := sb.getNextTicket()
	sb.clients[ticket] = args.Client

	fmt.Printf("Client %s have waited %dms to enter the sushi bar, and got ticket %d\n", args.Client.Name, timeWaiting.Milliseconds(), ticket)

	reply.Ticket = ticket

	sb.mu.Unlock()

	return nil
}

func (sb *SushiBarRPC) Leave(args *SushiBarLeaveArgs, reply *SushiBarLeaveReply) error {
	client, ok := sb.clients[args.Ticket]
	if !ok {
		return fmt.Errorf("Client not found")
	}
	delete(sb.clients, args.Ticket)

	fmt.Printf("Client %s left the restaurant\n", client.Name)

	reply.Farewell = fmt.Sprintf("Thank you %s! Hope to see you again", client.Name)

	sb.wg.Done()

	return nil
}

func (sb *SushiBarRPC) getNextTicket() int {
	sb.lastTicket++
	return sb.lastTicket
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("undefined PORT")
	}

	sushiBarRPC := NewSushiBarRPC(MAX_CLIENT)

	server := rpc.NewServer()
	if err := server.RegisterName("SushiBar", sushiBarRPC); err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("localhost:%s", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server.Accept(ln)
}

type Client struct {
	Name string
}

type SushiBarEnterArgs struct {
	Client *Client
}

type SushiBarEnterReply struct {
	Ticket int
}

type SushiBarLeaveArgs struct {
	Ticket int
}

type SushiBarLeaveReply struct {
	Farewell string
}
