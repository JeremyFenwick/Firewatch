package linereversal

import (
	"fmt"
	"log"
	"net"
	"os"
)

type udpMessage struct {
	message *LRMessage
	sender  net.Addr
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Determine if we are local or remote (fly io)
	bindingAddress := getBindingAddress(port)
	// Setup the udp listener
	udp, err := net.ListenPacket("udp", bindingAddress)
	if err != nil {
		log.Fatalf("can't listen on %d/udp: %s", port, err)
	}
	defer udp.Close()
	// Setup the session manager
	sessionManager := NewSessionManager()
	log.Printf("Line reversal listening on port %d", port)
	buffer := make([]byte, 999)
	// Start recieving messages
	for {
		n, senderAddress, err := udp.ReadFrom(buffer)
		if err != nil {
			log.Println("Could not recieve packet, continuing...")
			continue
		}
		message, err := DecodeLRMessage(buffer[:n])
		if err != nil {
			log.Printf("Could not decode message: %s", err)
			continue
		}
		request := &udpMessage{
			message: message,
			sender:  senderAddress,
		}
		handleRequest(request, udp, sessionManager)
	}
}

func handleRequest(request *udpMessage, udpConn net.PacketConn, sessionManager *SessionManager) {
	switch request.message.Type {
	case "connect":
		sessionManager.CreateSession(udpConn, request.sender, request.message.Session)
		sessionManager.SendMessage(request.message.Session, ConnectMessage(request.sender))
	case "data":
		sessionManager.SendMessage(request.message.Session, DataMessage(request.message.Position, request.message.Data, request.sender))
	case "ack":
		sessionManager.SendMessage(request.message.Session, AckMessage(request.message.Length, request.sender))
	case "close":
		sessionManager.SendMessage(request.message.Session, CloseMessage(request.sender))
	default:
		log.Printf("Unknown message type: %s", request.message.Type)
	}
}

func getBindingAddress(port int) string {
	_, exists := os.LookupEnv("FLY_APP_NAME")
	if exists {
		return fmt.Sprintf("fly-global-services:%d", port)
	} else {
		return fmt.Sprintf("0.0.0.0:%d", port)
	}
}
