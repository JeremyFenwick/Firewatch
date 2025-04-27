package insecuresocketslayer

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Client struct {
	Conn             net.Conn
	IncomingPosition int
	OutboundPosition int
	Reader           *bufio.Reader
	Cipher           *Cipher
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		// Use Fatalf for formatted error messages
		log.Fatalf("Could not start listener. REASON: %v", err)
	}
	log.Printf("Insecure socket layer now listening on port %d\n", port)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Use Printf for formatted error messages
			log.Printf("Encountered error accepting connection. REASON: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	client := &Client{
		Conn:             conn,
		IncomingPosition: 0,
		Reader:           bufio.NewReader(conn),
	}
	log.Printf("Handling connection from %s", conn.RemoteAddr())
	err := getClientCipher(client)
	if err != nil {
		log.Printf("Error getting client cipher: %v", err)
		return
	}
	// Begin handling the connection
	recieveLoop(client)
}

func getClientCipher(client *Client) error {
	// The cipher ends in 0x00
	cipherData, err := client.Reader.ReadBytes(0x00)
	if err != nil {
		log.Printf("Error reading cipher data: %v", err)
		return err
	}
	cipher, err := NewCipher(cipherData)
	if err != nil {
		log.Printf("Error creating cipher: %v", err)
		return err
	}
	client.Cipher = cipher
	return nil
}

// Loop for reading data from the client
func recieveLoop(client *Client) {
	buffer := make([]byte, 5001)
	// Being the loop to read data from the client
	for {
		readBytes, err := client.Reader.Read(buffer)
		if err != nil {
			log.Printf("Error reading from client: %v", err)
			return
		}
		if readBytes == 0 {
			log.Printf("Empty data received from client")
			return
		}
		err = respondToClient(client, buffer[:readBytes])
		if err != nil {
			log.Printf("Error responding to client: %v", err)
			return
		}
	}
}

// Respond to the client with the most common toy
func respondToClient(client *Client, data []byte) error {
	// Handle the incoming data
	decoded := client.Cipher.DecodeData(client.IncomingPosition, data)
	if decoded[len(decoded)-1] != '\n' {
		return fmt.Errorf("data from client is not terminated with a newline")
	}
	plainToy, err := MostCommonToy(string(decoded))
	if err != nil {
		return err
	}
	client.IncomingPosition += len(data)
	// Handle the outgoing data
	encodedToy := client.Cipher.EncodeData(client.OutboundPosition, append([]byte(plainToy), '\n'))
	client.OutboundPosition += len(encodedToy)
	// Now send the response
	_, err = client.Conn.Write(encodedToy)
	if err != nil {
		return err
	}
	return nil
}
