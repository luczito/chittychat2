package main

import (
	"log"
	"net"
	"os"

	chat "grpcChatServer/chat"

	"google.golang.org/grpc"
)

func main() {
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "8080" //default Port 8080 if not set in commandline
	}

	//reg listener on port
	listen, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Unable to listen on port: %v :: %v", Port, err)
	}
	log.Println("Listening on port: " + Port)

	//reg grpc
	grpcserver := grpc.NewServer()

	cs := chat.Chat{}
	chat.RegisterServicesServer(grpcserver, &cs)

	err = grpcserver.Serve(listen)
	if err != nil {
		log.Fatalf("Unable to start gRPC Server :: %v", err)
	}

}
