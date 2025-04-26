package linereversal

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

const Retransmission = 100 * time.Millisecond
const SessionAcknowledgementTimeout = 60 * time.Second
const BufferCapacity = 64 * 1024 // 64KB

// We use an actor model for the session
type Session struct {
	// Session state
	IsClosed         bool
	Address          net.Addr
	ID               int
	Conn             net.PacketConn
	Channel          chan SessionMessage
	Logger           *log.Logger
	LastAck          int
	MaxAck           int
	RecievedPosition int

	// Used for handling outgoing data
	RetransmitTicker *time.Timer
	PendingData      *PendingData
	OutgoingBuffer   []byte
	WritePosition    int

	// Used for storing recieved data
	DataStore []byte

	// Buffer used for sending messages
	// Note there is a 1000 byte limit on the UDP message size
	SendBuffer []byte
}

// Used to track outgoing data in transit
type PendingData struct {
	Length  int
	SentAt  time.Time
	Payload *LRMessage
}

func NewSession(conn net.PacketConn, address net.Addr, id int, messageChannel chan SessionMessage) *Session {
	session := &Session{
		Conn:           conn,
		Address:        address,
		ID:             id,
		IsClosed:       false,
		Channel:        messageChannel,
		Logger:         log.New(log.Writer(), fmt.Sprintf("Session %d: ", id), log.LstdFlags),
		DataStore:      make([]byte, 0, BufferCapacity),
		OutgoingBuffer: make([]byte, 0, BufferCapacity),
		SendBuffer:     make([]byte, 0, 1000),
	}

	// Start the recieve loop
	go session.RecieveMessage()
	return session
}

func (s *Session) RecieveMessage() {
	for {
		// The transmitter may not be there
		var retransmit <-chan time.Time
		if s.RetransmitTicker != nil {
			retransmit = s.RetransmitTicker.C
		}
		select {
		case <-retransmit:
			if s.PendingData == nil {
				continue
			}
			// If the data is expired, we need to close the session
			if time.Since(s.PendingData.SentAt) > SessionAcknowledgementTimeout {
				s.Logger.Printf("did not recieve acknowledgement from client after %s, closing session.", SessionAcknowledgementTimeout)
				s.Close()
				return
			} else {
				// If the data is not expired, we need to retransmit it
				s.Logger.Println("Retransmitting data to client")
				s.SendDataMessage(s.PendingData.Payload)
				s.RetransmitTicker.Reset(Retransmission)
			}
		// Else if we receive a message from the channel
		case msg, ok := <-s.Channel:
			if !ok {
				s.Logger.Println("Channel closed, exiting receive loop.")
				return
			}
			// Update the address if it is different
			if msg.Address != nil && msg.Address != s.Address {
				s.Address = msg.Address
			}
			switch msg.Type {
			case "connect_client":
				s.SendAckMessage(0)
			// Instruction to close the session
			case "close_client":
				s.Close()
				return
			// Instruction to add to the buffer
			case "recieved_data":
				s.HandleRecieveData(msg.Data, msg.Number)
			// If we recieve an ack we need to handle it
			case "recieved_ack":
				s.HandleAck(msg.Number)
			}
		}
	}
}

func (s *Session) Close() {
	s.SendCloseMessage()
	s.IsClosed = true
}

func (s *Session) HandleAck(length int) {
	s.Logger.Printf("Recieved ack message: %d", length)
	// If this is smaller than our last ack, we ignore it as it is a delayed message
	if length < s.LastAck {
		s.Logger.Printf("Received delayed ack message: u%d < %d", length, s.LastAck)
		return
	}
	// If this is larger than our max ack, we need to close the session
	if length > s.MaxAck {
		s.Logger.Printf("Received ack message: %d > %d. Closing the session", length, s.MaxAck)
		s.Close()
		return
	}
	// If we have no pending data, we can exit
	if s.PendingData == nil {
		s.Logger.Printf("Received ack message: %d. No pending data", length)
		return
	}
	// If the ack matches or is less that what we send, we can remove the acknowledged data
	expectedLength := s.WritePosition + s.PendingData.Length
	if length <= expectedLength {
		s.Logger.Printf("Received ack for pending data: %d", length)
		s.RetransmitTicker.Stop()
		// Remove the acknowledged data from the outgoing buffer
		s.OutgoingBuffer = s.OutgoingBuffer[length-s.WritePosition:]
		// Set the write position to the new position
		s.WritePosition += length - s.WritePosition
		// Wipe the pending data
		s.PendingData = nil
		// Update the last ack
		s.LastAck = length
		// We may have more data to send, so we need to check the outgoing buffer
		s.HandleOutgoingBuffer()
	}
}

func (s *Session) HandleOutgoingBuffer() {
	// If the buffer is empty, or we already have an outgoing message we can return
	if len(s.OutgoingBuffer) == 0 || s.PendingData != nil {
		return
	}
	// Create a new buffer for the message
	buffer := make([]byte, 0, 1000)
	// If we have pending data, we need to need to pack it and send it
	newDataMessage := &LRMessage{
		Type:     "data",
		Session:  s.ID,
		Position: s.WritePosition,
		Data:     buffer,
	}
	bytesUsed := PackDataMessage(newDataMessage, s.OutgoingBuffer, s.WritePosition)
	// The max ack is the write position + the bytes used. Anything we recieve beyond this number we recieve from the client is invalid
	s.MaxAck = s.WritePosition + bytesUsed
	// Set the pending data
	s.PendingData = &PendingData{
		Length:  bytesUsed,
		SentAt:  time.Now(),
		Payload: newDataMessage,
	}
	// Send the data
	s.SendDataMessage(newDataMessage)
	// Now set the timer to recieve the ack from the client
	s.RetransmitTicker = time.NewTimer(Retransmission)
}

func (s *Session) HandleRecieveData(data []byte, position int) {
	// According to the protocol, we recieve data in order. So if the data
	// is out of order we send an ack message with our current position
	if position != s.RecievedPosition {
		s.SendAckMessage(s.RecievedPosition)
		return
	}
	// If the data is in order, we add it to the buffer
	unescaped, err := UnescapeData(data)
	if err != nil {
		s.Logger.Println("Error unescaping data:", err)
		return
	}
	s.RecievedPosition += len(unescaped)
	// Send an ack in response
	s.SendAckMessage(s.RecievedPosition)
	// Transmit the data to the read channel
	s.Logger.Printf("Recieved data of length %d \n", len(unescaped))
	s.DataStore = append(s.DataStore, unescaped...)
	// We need to check to see if we recieved a newline
	s.CheckDataStore()
}

func (s *Session) CheckDataStore() {
	for {
		// If the data store is empty, we can return
		if len(s.DataStore) == 0 {
			break
		}
		// Search for the next new line
		newLineIndex := bytes.IndexByte(s.DataStore, '\n')
		// No new line found, so we can break
		if newLineIndex == -1 {
			break
		}
		// We have a new line, so we need to reverse the data up to the new line
		reversed := make([]byte, newLineIndex+1)
		for i, b := range s.DataStore[0:newLineIndex] {
			reversed[newLineIndex-1-i] = b
		}
		// Add the new line character
		reversed[newLineIndex] = '\n'
		// Now load the data into the output buffer
		s.OutgoingBuffer = append(s.OutgoingBuffer, reversed...)
		// Remove the processed data from the buffer
		s.DataStore = s.DataStore[newLineIndex+1:]
	}
	// Now we need to handle the outgoing buffer
	s.HandleOutgoingBuffer()
}

func (s *Session) SendCloseMessage() {
	closeMessage := &LRMessage{
		Type:    "close",
		Session: s.ID,
	}
	// Reset the send buffer
	s.SendBuffer = s.SendBuffer[:0]
	messageLength, err := closeMessage.Encode(s.SendBuffer)
	if err != nil {
		s.Logger.Println("Error encoding close message:", err)
		return
	}
	_, err = s.Conn.WriteTo(s.SendBuffer[:messageLength], s.Address)
	if err != nil {
		s.Logger.Println("Error sending close message:", err)
		return
	}
}

func (s *Session) SendDataMessage(message *LRMessage) {
	// Reset the send buffer
	s.SendBuffer = s.SendBuffer[:0]
	messageLength, err := message.Encode(s.SendBuffer)
	if err != nil {
		s.Logger.Println("Error encoding data message:", err)
		return
	}
	s.Logger.Printf("Sending data message of length: %d. Expecting ack back of: %d \n", len(message.Data), s.MaxAck)
	_, err = s.Conn.WriteTo(s.SendBuffer[:messageLength], s.Address)
	if err != nil {
		s.Logger.Println("Error sending data message:", err)
		return
	}
}

func (s *Session) SendAckMessage(length int) {
	ackMessage := &LRMessage{
		Type:    "ack",
		Session: s.ID,
		Length:  length,
	}
	// Reset the send buffer
	s.SendBuffer = s.SendBuffer[:0]
	messageLength, err := ackMessage.Encode(s.SendBuffer)
	if err != nil {
		s.Logger.Println("Error encoding ack message:", err)
		return
	}
	_, err = s.Conn.WriteTo(s.SendBuffer[:messageLength], s.Address)
	if err != nil {
		s.Logger.Println("Error sending ack message:", err)
		return
	}
}

func UnescapeData(data []byte) ([]byte, error) {
	output := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		// We end with an escaping slash which is invalid
		if i == len(data)-1 && data[i] == '\\' {
			return nil, fmt.Errorf("invalid data: %s", data)
		}
		// If we encounter an escaped slash, add the next character to the output
		if data[i] == '\\' && i < len(data)-1 && (data[i+1] == '/' || data[i+1] == '\\') {
			output = append(output, data[i+1])
			i++
			continue
		}
		// Otherwise, just add the character to the output
		output = append(output, data[i])
	}
	return output, nil
}
