package chat

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type messageUnit struct {
	ClientName  string
	MessageBody string
	ClientId    int
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}

type Chat struct {
}

func (is *Chat) ChatService(csi Services_ChatServiceServer) error {
	ClientId := rand.Intn(1e6)
	errch := make(chan error)

	go receiveFromStream(csi, ClientId, errch)
	go sendToStream(csi, ClientId, errch)

	return <-errch
}

func receiveFromStream(csi_ Services_ChatServiceServer, ClientId_ int, errch_ chan error) {
	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else {
			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:  mssg.Name,
				MessageBody: mssg.Body,
				ClientId:    ClientId_,
			})
			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()
		}
	}
}

func sendToStream(csi_ Services_ChatServiceServer, ClientId_ int, errch_ chan error) {
	for {
		for {
			time.Sleep(250 * time.Millisecond) //sleep to prevent spam

			messageHandleObject.mu.Lock()
			if len(messageHandleObject.MQue) == 0 { //check for messages in queue
				messageHandleObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject.MQue[0].ClientId
			senderName4Client := messageHandleObject.MQue[0].ClientName
			message4Client := messageHandleObject.MQue[0].MessageBody

			messageHandleObject.mu.Unlock()

			//send message to all clients except the sender
			if senderUniqueCode != ClientId_ {
				err := csi_.Send(&FromServer{Name: senderName4Client, Body: message4Client})
				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete message after sending
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}
				messageHandleObject.mu.Unlock()
			}
		}
		time.Sleep(50 * time.Millisecond) //sleep to prevent spam
	}
}
