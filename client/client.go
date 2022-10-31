package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	chat "grpcChatServer/chat"

	"google.golang.org/grpc"
)

func main() {
	//set ip/port for server
	fmt.Print("Enter Server IP:Port, press enter for default IP:Port")
	reader := bufio.NewReader(os.Stdin)
	serverID, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Unable to read from console: %v", err)
	}
	serverID = strings.Trim(serverID, "\r\n")

	//if ip:port is not set then default value.
	if serverID == "" {
		serverID = "localhost:8080"
	}
	log.Println("Connecting to the server: " + serverID)

	//connect to grpc server
	conn, err := grpc.Dial(serverID, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Unable to connect to gRPC server :: %v", err)
	}
	defer conn.Close()

	//create stream via chat
	client := chat.NewServicesClient(conn)

	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Unable to call ChatService: %v", err)
	}

	// reg communication with server
	ch := clienthandle{stream: stream}
	ch.clientConfig()
	go ch.sendMessage()
	go ch.receiveMessage()

	bl := make(chan bool)
	<-bl

}

type clienthandle struct {
	stream     chat.Services_ChatServiceClient
	clientName string
}

// method to set custom usernames.
func (ch *clienthandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your username: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Unable to read username from console: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
}

func (ch *clienthandle) sendMessage() {
	for {
		fmt.Printf("->")
		reader := bufio.NewReader(os.Stdin) //setup reader to console
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Unable to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		//register message from console input
		clientMessageBox := &chat.FromClient{
			Name: ch.clientName,
			Body: clientMessage,
		}

		err = ch.stream.Send(clientMessageBox)
		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}
	}
}

func (ch *clienthandle) receiveMessage() {
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
		}
		fmt.Printf("\nUser %s : %s \n->", mssg.Name, mssg.Body)
	}
}
