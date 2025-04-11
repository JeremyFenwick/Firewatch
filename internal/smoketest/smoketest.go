package smoketest

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	maxTotalBytes = 1024 * 1024 // 1MB max before disconnecting
	batchSize     = 4 * 1024    // 4KB per batch
)

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		// Use Fatalf for formatted error messages
		log.Fatalf("Could not start listener. REASON: %v", err)
	}
	log.Printf("Smoke test now listening on port %d\n", port)
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
	log.Printf("Handling connection from %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	totalBytes := 0
	buffer := make([]byte, batchSize)

	for {
		// Read a batch
		n, err := reader.Read(buffer)
		if n > 0 {
			totalBytes += n

			// Echo the batch back
			_, writeErr := conn.Write(buffer[:n])
			if writeErr != nil {
				log.Printf("Error writing to %s: %v", conn.RemoteAddr(), writeErr)
				return
			}

			log.Printf("Echoed %d bytes to %s: %s", n, conn.RemoteAddr(), string(buffer[:n]))
		}

		// Stop if we reached max allowed bytes
		if totalBytes >= maxTotalBytes {
			log.Printf("Max byte limit reached for %s. Closing connection.", conn.RemoteAddr())
			return
		}

		// Handle errors or EOF
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection %s closed by client.", conn.RemoteAddr())
			} else {
				log.Printf("Error reading from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}
	}
}
