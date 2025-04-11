package speeddaemon

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	camera = iota
	ticketDispatcher
	unknown
)

type ConnKind int

type Connection struct {
	Conn       net.Conn
	ConnKind   ConnKind
	Limit      U16 // For a camera only
	Road       U16 // For a camera only
	Mile       U16 // For a camera only
	HBInterval float64
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		// Use Fatalf for formatted error messages
		log.Fatalf("Could not start listener. REASON: %v", err)
	}
	log.Printf("Speed daemon now listening on port %d\n", port)
	defer listener.Close()

	dispatcher := NewCentralDispatcher()
	go dispatcher.Start()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Use Printf for formatted error messages
			log.Printf("Encountered error accepting connection. REASON: %v", err)
			continue
		}

		connection := &Connection{
			Conn:       conn,
			ConnKind:   unknown,
			HBInterval: 0,
		}

		go handleConnection(connection, dispatcher)
	}
}

func handleConnection(connection *Connection, dispatcher *CentralDispatcher) {
	defer connection.Conn.Close()
	buffer := make([]byte, 0, 1024) // Adjust buffer size as needed

	for {
		data := make([]byte, 1024)
		n, err := connection.Conn.Read(data)
		if err != nil {
			log.Printf("Error reading from %s: %v", connection.Conn.RemoteAddr(), err)
			return
		}
		buffer = append(buffer, data[:n]...)
		sfBuffer := NewSdBuffer(buffer)
		messages, extractedBytes := ExtractFromSbBuffer(sfBuffer)
		buffer = buffer[extractedBytes:]
		processMessages(messages, connection, dispatcher)
	}
}

func processMessages(messages []ClientMessage, connection *Connection, dispatcher *CentralDispatcher) {
	for _, message := range messages {
		switch message.GetType() {
		case IAmCameraType:
			log.Printf("Received IAmCameraMessage from %s", connection.Conn.RemoteAddr())
			registerCamera(message.(*IAmCameraMessage), connection, dispatcher)
		case IAmDispatcherType:
			log.Printf("Received IAmDispatcherMessage from %s", connection.Conn.RemoteAddr())
			channel := registerDispatcher(message.(*IAmDispatcherMessage), connection, dispatcher)
			if channel == nil {
				log.Printf("Failed to register dispatcher for %s", connection.Conn.RemoteAddr())
				return
			}
			go dispatcherListener(channel, connection)
		case PlateMsgType:
			log.Printf("Received PlateMessage from %s", connection.Conn.RemoteAddr())
			handlePlateMessage(message.(*PlateMessage), connection, dispatcher)
		case WantHeartbeatType:
			log.Printf("Received WantHeartbeatMessage from %s", connection.Conn.RemoteAddr())
			handleHeartbeatRequest(message.(*WantHeartbeatMessage), connection)
		default:
			log.Printf("Received unknown message type from %s: %d", connection.Conn.RemoteAddr(), message.GetType())
			sendError("Unknown message. Closing connection", connection)
			return
		}
	}
}

func handleHeartbeatRequest(wantHeartbeatMessage *WantHeartbeatMessage, connection *Connection) {
	if connection.HBInterval != 0 {
		sendError("Already registered a heartbeat to this connection", connection)
		connection.Conn.Close()
		return
	}
	connection.HBInterval = float64(wantHeartbeatMessage.Interval) * 0.1
	log.Printf("Received WantHeartbeatMessage from %s with interval %f", connection.Conn.RemoteAddr(), connection.HBInterval)
	go heartbeat(connection)
}

func heartbeat(connection *Connection) {
	heartbeat := &HeartbeatMessage{}
	encoded, err := heartbeat.Encode()
	if err != nil {
		log.Printf("Error encoding heartbeat message: %v", err)
		return
	}
	for range time.Tick(time.Duration(connection.HBInterval * float64(time.Second))) {
		_, err := connection.Conn.Write(encoded)
		if err != nil {
			log.Printf("Error sending heartbeat message: %v", err)
			return
		}
		// log.Printf("Sent heartbeat to %s", connection.Conn.RemoteAddr())
	}
}

func handlePlateMessage(plateMessage *PlateMessage, connection *Connection, dispatcher *CentralDispatcher) {
	if connection.ConnKind != camera {
		sendError("Only cameras can send plate messages", connection)
		connection.Conn.Close()
		return
	}
	dispatcher.MessageQueue <- &Observation{
		Road:      connection.Road,
		License:   plateMessage.Plate,
		Mile:      connection.Mile,
		Timestamp: plateMessage.Timestamp,
	}
}

func registerCamera(message *IAmCameraMessage, connection *Connection, dispatcher *CentralDispatcher) {
	if connection.ConnKind != unknown {
		sendError("Already registered this connection", connection)
		connection.Conn.Close()
		return
	}
	connection.ConnKind = camera
	connection.Limit = message.Limit
	connection.Road = message.Road
	connection.Mile = message.Mile
	dispatcher.MessageQueue <- &RegisterCamera{
		Road:  message.Road,
		Limit: message.Limit,
	}
	log.Printf("Registering camera on road %d at mile %d with limit %d", message.Road, message.Mile, message.Limit)
}

func registerDispatcher(message *IAmDispatcherMessage, connection *Connection, dispatcher *CentralDispatcher) chan ClientMessage {
	if connection.ConnKind != unknown {
		sendError("Already registered this connection", connection)
		connection.Conn.Close()
		return nil
	}
	connection.ConnKind = ticketDispatcher
	dispatchChannel := make(chan ClientMessage, 10)
	dispatcher.MessageQueue <- &RegisterDispatcher{
		Roads:   message.Roads,
		Channel: dispatchChannel,
	}
	log.Printf("Registering dispatcher with roads %v", message.Roads)
	return dispatchChannel
}

func dispatcherListener(channel chan ClientMessage, connection *Connection) {
	for message := range channel {
		switch message.GetType() {
		case TicketMsgType:
			log.Printf("Received ticket message from central dispatcher %s", connection.Conn.RemoteAddr())
			ticketMessage := message.(*TicketMessage)
			encoded, err := ticketMessage.Encode()
			if err != nil {
				log.Printf("Error encoding ticket message: %v", err)
				return
			}
			connection.Conn.Write(encoded)
		default:
			log.Printf("Received unknown message type from dispatcher %s: %d", connection.Conn.RemoteAddr(), message.GetType())
		}
	}
}

func sendError(errorMessage string, connection *Connection) {
	errorMsg := &ErrorMessage{
		Content: Str(errorMessage),
	}
	encodedMessage, err := errorMsg.Encode()
	if err != nil {
		log.Printf("Error encoding error message: %v", err)
		return
	}
	_, err = connection.Conn.Write(encodedMessage)
	if err != nil {
		log.Printf("Error sending error message: %v", err)
		return
	}
}
