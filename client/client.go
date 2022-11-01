package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	chat "test/proto"

	"google.golang.org/grpc"
)

type Client struct {
	name   string
	port   string
	lclock uint64
}

func parse() (string, error) {
	var input string
	sc := bufio.NewScanner(os.Stdin)
	if sc.Scan() {
		input = sc.Text()
	}
	return input, nil
}

func listen(clientConnection chat.Chat_ConnectClient, c *Client) {
	for {
		msg, err := clientConnection.Recv()
		if err != nil {
			log.Fatalf("Unable to recieve the message %v", err)
		}
		if msg.Lclock > c.lclock {
			c.lclock = msg.Lclock
		}
		c.lclock++

		fmt.Printf("%v: %v\n", msg.Name, msg.Msg)
	}
}

func start(c *Client) {
	fmt.Printf("Enter you username: ")

	var clientConnection chat.Chat_ConnectClient

	for {
		connection, err := grpc.Dial(c.name+":"+c.port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Unable to connect to the server %v", err)
		}

		client := chat.NewChatClient(connection)

		clientConnection, err = client.Connect(context.Background())
		if err != nil {
			log.Fatalf("Unable to connect to the server %v", err)
		}

		username, err := parse()
		if err != nil {
			log.Fatalf("Unable to retrieve username %v", err)
		}

		clientConnection.Send(&chat.ClientMsg{
			Name:   c.name,
			Msg:    username,
			Lclock: c.lclock,
		})

		message, err := clientConnection.Recv()
		if err == nil {
			fmt.Printf("%v: %v\n", message.Name, message.Msg)
			break
		}
		fmt.Println("Username already in use, please try again.")
	}

	go listen(clientConnection, c)

	for {
		input, err := parse()
		if err != nil {
			log.Fatalf("Unable to retrieve the user input")
		}

		if input == "exit" {
			break
		}

		c.lclock++

		clientConnection.Send(&chat.ClientMsg{
			Name:   c.name,
			Msg:    input,
			Lclock: c.lclock,
		})
	}
	clientConnection.CloseSend()
}

func main() {
	//set ip/port for server
	fmt.Print("Enter Server IP:Port, press enter for default IP:Port")
	reader := bufio.NewReader(os.Stdin)
	server, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Unable to read from console: %v", err)
	}
	server = strings.Trim(server, "\r\n")
	var serverName string
	var serverPort string
	//if ip:port is not set then default value.
	if server == "" {
		serverName = "localhost"
		serverPort = "8080"
	} else {
		serverInfo := strings.Split(server, ":")
		serverName = serverInfo[0]
		serverPort = serverInfo[1]
	}

	log.Printf("Connecting to the server: %v:%v", serverName, serverPort)

	client := &Client{
		name:   serverName,
		port:   serverPort,
		lclock: 0,
	}
	start(client)
}
