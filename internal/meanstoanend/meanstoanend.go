package meanstoanend

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

type messageType int

const (
	Insert messageType = iota
	Query
	Unknown
)

type message struct {
	messageType
	first  int32
	second int32
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Could not start listener. REASON: " + err.Error())
	}
	log.Printf("Means to an end listening on port %d\n", port)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Encountered error accepting connection. REASON: " + err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	history := map[int32]int32{}

	for {
		buffer := make([]byte, 9)
		_, err := io.ReadFull(reader, buffer)
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Failed to receive message from connection:", err)
			return
		}

		message := extractMessage(buffer)
		handleMessage(message, history, conn)
	}
}

func extractMessage(data []byte) *message {
	var messageT messageType
	// Get the message type
	if data[0] == byte('I') {
		messageT = Insert
	} else if data[0] == byte('Q') {
		messageT = Query
	} else {
		messageT = Unknown
	}
	// Get the first number (signed int32)
	firstNumberBytes := data[1:5]
	firstNumber := int32(binary.BigEndian.Uint32(firstNumberBytes))

	// Get the second number (signed int32)
	secondNumberBytes := data[5:9]
	secondNumber := int32(binary.BigEndian.Uint32(secondNumberBytes))

	return &message{
		messageType: messageT,
		first:       firstNumber,
		second:      secondNumber,
	}
}

func handleMessage(message *message, history map[int32]int32, conn net.Conn) {
	if message.messageType == Insert {
		history[message.first] = message.second
	} else if message.messageType == Query {
		handleQuery(message.first, message.second, history, conn)
	} else {
		return
	}
}

func handleQuery(start, end int32, history map[int32]int32, conn net.Conn) {
	var count int64
	var sum int64
	var average int32

	log.Printf("Processing query for range %d-%d with history: %+v\n", start, end, history) // Added detailed log

	for time, price := range history {
		if time >= start && time <= end {
			count++
			sum += int64(price)
		}
	}

	if count > 0 {
		average64 := sum / count
		average = int32(average64)
	}

	response := make([]byte, 4)
	binary.BigEndian.PutUint32(response, uint32(average))
	log.Printf("Responding with average: %d", average)

	bytes, err := conn.Write(response)
	if err != nil {
		log.Println("Could not write to client:", err)
		return
	}
	log.Printf("Successfully sent %d bytes", bytes)
}
