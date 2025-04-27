package insecuresocketslayer

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
)

const BufferSize = 5001      // Spec says this is the maximum size of a completed message
const maxCipherSpecSize = 80 // Maximum size of the cipher spec

type Client struct {
	Conn             net.Conn
	IncomingPosition int
	OutboundPosition int
	Reader           *bufio.Reader
	Cipher           *Cipher
	Logger           *log.Logger
	Buffer           []byte
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
		OutboundPosition: 0,
		Buffer:           make([]byte, 0, BufferSize),
		Logger: log.New(log.Writer(),
			fmt.Sprintf("[%s] ", conn.RemoteAddr().String()),
			log.Flags()|log.Lmsgprefix|log.Lshortfile),
	}
	log.Printf("Handling connection from %s", conn.RemoteAddr())
	// Get the cipher
	specBytes, err := readCipherSpecBytes(client.Reader)
	if err != nil {
		client.Logger.Printf("Error reading cipher spec bytes: %v", err)
		return
	}
	cipher, err := NewCipher(specBytes)
	if err != nil {
		client.Logger.Printf("Error creating cipher: %v", err)
		return
	}
	if !cipher.Valid {
		client.Logger.Printf("Invalid cipher")
		return
	}
	client.Cipher = cipher
	// Begin handling the connection
	recieveLoop(client)
}

func readCipherSpecBytes(reader *bufio.Reader) ([]byte, error) {
	var specBytes []byte
	expectingOperand := false

	for len(specBytes) < maxCipherSpecSize {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		specBytes = append(specBytes, b)

		if expectingOperand {
			expectingOperand = false
			continue
		}

		switch b {
		case 0x00: // Terminator
			return specBytes, nil
		case 0x01, 0x03, 0x05: // Opcodes without operands
			expectingOperand = false
		case 0x02, 0x04: // Opcodes needing one operand
			expectingOperand = true
		default: // Invalid opcode
			return specBytes, fmt.Errorf("encountered invalid opcode when recieving byte %x", b)
		}
	}

	return nil, fmt.Errorf("cipher spec too long, max size is %d bytes", maxCipherSpecSize)
}

// Loop for reading data from the client
func recieveLoop(client *Client) {
	temp := make([]byte, BufferSize)
	// Being the loop to read data from the client
	for {
		readBytes, err := client.Reader.Read(temp)
		if err != nil {
			return
		}
		if readBytes == 0 {
			client.Logger.Printf("Empty data received from client")
			return
		}
		decoded := client.Cipher.DecodeData(client.IncomingPosition, temp[:readBytes])
		client.Logger.Printf("Decoded data of length: %d", len(decoded))
		client.IncomingPosition += readBytes
		client.Buffer = append(client.Buffer, decoded...)
		// Loop through the buffer to find new lines
		for {
			newLineIndex := bytes.Index(client.Buffer, []byte{'\n'})
			if newLineIndex == -1 {
				break // No new line found, exit the loop
			}
			err = respondToClient(client, client.Buffer[:newLineIndex+1])
			if err != nil {
				client.Logger.Printf("Error responding to client: %v", err)
				return
			}
			// Remove the processed data from the buffer
			client.Buffer = client.Buffer[newLineIndex+1:]
		}
	}
}

// Respond to the client with the most common toy
func respondToClient(client *Client, decoded []byte) error {
	// Handle the incoming data
	plainToy, err := MostCommonToy(decoded)
	if err != nil {
		return err
	}
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
