package linereversal

import (
	"fmt"
	"log"
	"net"
	"os"
)

type udpMessage struct {
	data   []byte
	sender net.Addr
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
	// Setup the incoming buffer
	incoming := make(chan *udpMessage, 1000)
	defer close(incoming)
	// Setup the session manager
	sessionManager := NewSessionManager()
	// Start the incoming buffer
	go incomingBuffer(incoming, udp, sessionManager)
	log.Printf("Line reversal listening on port %d", port)
	// Start recieving messages
	for {
		buffer := make([]byte, 999)
		n, senderAddress, err := udp.ReadFrom(buffer)
		if err != nil {
			log.Println("Could not recieve packet, continuing...")
			continue
		}
		udpMessage := &udpMessage{
			data:   buffer[:n],
			sender: senderAddress,
		}
		incoming <- udpMessage
	}
}

func incomingBuffer(incoming chan *udpMessage, udpConn net.PacketConn, sessionManager *SessionManager) {
	outputBuffer := make([]byte, 0, 999)
	for {
		// Read the incomingMessage from the channel
		incomingMessage, ok := <-incoming
		if !ok {
			return
		}
		// Process the data
		decodedMessage, err := DecodeLRMessage(incomingMessage.data)
		if err != nil {
			log.Printf("Could not decode message: %s", err)
			continue
		}
		handleRequest(decodedMessage, incomingMessage.sender, udpConn, sessionManager, outputBuffer)
	}
}

func handleRequest(message *LRMessage, sender net.Addr, udpConn net.PacketConn, sessionManager *SessionManager, outputBuffer []byte) {
	log.Printf("Recieved message: %s", message.String())
	switch message.Type {
	case "connect":
		sessionManager.CreateSession(udpConn, sender, message.Session)
		sessionManager.SendMessage(message.Session, ConnectMessage(sender))
	case "data":
		if !sessionManager.SessionExists(message.Session) {
			sendCloseResponse(message.Session, udpConn, sender, outputBuffer)
		} else {
			sessionManager.SendMessage(message.Session, DataMessage(message.Position, message.Data, sender))
		}
	case "ack":
		if !sessionManager.SessionExists(message.Session) {
			sendCloseResponse(message.Session, udpConn, sender, outputBuffer)
		} else {
			sessionManager.SendMessage(message.Session, AckMessage(message.Length, sender))
		}
	case "close":
		if !sessionManager.SessionExists(message.Session) {
			sendCloseResponse(message.Session, udpConn, sender, outputBuffer)
		} else {
			sessionManager.SendMessage(message.Session, CloseMessage(sender))
		}
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

func sendCloseResponse(sessionId int, udpConn net.PacketConn, sender net.Addr, outputBuffer []byte) {
	message := &LRMessage{
		Type:    "close",
		Session: sessionId,
	}
	// Reset the output buffer
	outputBuffer = outputBuffer[:0]
	encodedBytes, err := message.Encode(outputBuffer)
	if err != nil {
		log.Printf("Error encoding message: %s", err)
		return
	}
	// Send the message to the sender
	_, err = udpConn.WriteTo(outputBuffer[:encodedBytes], sender)
	if err != nil {
		log.Printf("Error sending message: %s", err)
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
