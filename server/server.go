package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	chat "test/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	chat.UnimplementedChatServer
	name        string
	port        string
	mutex       sync.Mutex
	lclock      uint64
	clients     map[string]chat.Chat_ConnectServer
	clientNames map[string]string
}

var name = flag.String("name", "localhost", "the server name")
var port = flag.String("port", "8080", "the server port")

func checkUsername(m map[string]string, name_ string) bool {
	for _, name := range m {
		if name == name_ {
			return true
		}
	}
	return false
}

func (s *Server) Connect(stream chat.Chat_ConnectServer) error {
	log.Printf("Participant connected.")

	name, err := stream.Recv()
	if err != nil {
		log.Printf("Unable to recieve the username %v", err)
	}
	if checkUsername(s.clientNames, name.Msg) {
		log.Printf("Username already taken: %v", name.Msg)
		return errors.New("username already taken")
	}

	peer_, _ := peer.FromContext(stream.Context())
	s.clients[peer_.Addr.String()] = stream
	log.Printf("User: %v, connected to the chat via %v\n", name.Msg, peer_.Addr.String())

	s.clientNames[peer_.Addr.String()] = name.Msg

	message := &chat.ServerMsg{
		Name:   "Server message",
		Msg:    fmt.Sprintf("%v just joined!", name.Msg),
		Lclock: s.lclock,
	}

	log.Printf("Sending new participant connection to all participants.")

	for _, client := range s.clients {
		client.Send(message)
	}

	for {
		log.Printf("Listening for messages")
		msg, err := stream.Recv()
		if err != nil {
			if status.Code(err).String() == "Cancelled" || status.Code(err).String() == "EOF" {
				log.Printf("%v disconnected from the chat via %v\n", name.Msg, peer_.Addr.String())
				break
			} else {
				log.Printf("Unable to recieve the message %v", err)
				break
			}
		}
		s.mutex.Lock()
		if msg.Lclock > s.lclock {
			s.lclock = msg.Lclock
		}
		s.lclock++
		s.mutex.Unlock()

		log.Printf("Lclock: %v | Recieved message %v, from %v", s.lclock, msg.Msg, msg.Name)

		s.mutex.Lock()
		s.lclock++
		s.mutex.Unlock()

		message := &chat.ServerMsg{
			Name:   s.clientNames[peer_.Addr.String()],
			Msg:    msg.Msg,
			Lclock: s.lclock,
		}

		log.Printf("Send message to all other participants with lclock: %v", s.lclock)
		for _, client := range s.clients {
			client.Send(message)
		}
	}
	leaveMessage := &chat.ServerMsg{
		Name:   "Server",
		Msg:    fmt.Sprintf("%v left the chat", s.clientNames[peer_.Addr.String()]),
		Lclock: s.lclock,
	}

	log.Printf("%v Send leave msg to all other participants\n", s.lclock)

	for _, client := range s.clients {
		client.Send(leaveMessage)
	}

	delete(s.clients, peer_.Addr.String())
	delete(s.clientNames, peer_.Addr.String())
	return nil
}

func start(s *Server) {
	listener, err := net.Listen("tcp", s.name+":"+s.port)
	if err != nil {
		log.Fatalf("Unable to create server %v", err)
	}

	grpcServer := grpc.NewServer()

	log.Printf("Creating server on port %v", s.port)

	chat.RegisterChatServer(grpcServer, s)
	serveErr := grpcServer.Serve(listener)
	if serveErr != nil {
		log.Fatalf("Unable to start the server %v", serveErr)
	}
}

func main() {
	f, err := os.OpenFile("log.server", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	flag.Parse()

	server := &Server{
		name:        *name,
		port:        *port,
		lclock:      0,
		clients:     make(map[string]chat.Chat_ConnectServer),
		clientNames: make(map[string]string),
	}

	go start(server)
	for {
		time.Sleep(500 * time.Millisecond)
	}
}
