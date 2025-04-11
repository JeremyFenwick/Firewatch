package smoketest_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JeremyFenwick/firewatch/internal/smoketest" // Replace with your actual module path
)

func TestEchoServer(t *testing.T) {
	// Use a random available port for testing
	port := getFreePort(t)

	// Start the server in a goroutine
	wg := &sync.WaitGroup{}
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer wg.Done()
		go func() {
			<-ctx.Done()
			// This is a trick to unblock the Listen function by connecting to it
			// after the test is done
			conn, _ := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
			if conn != nil {
				conn.Close()
			}
		}()
		smoketest.Listen(port)
	}()

	// Wait a short time for the server to start
	time.Sleep(100 * time.Millisecond)

	t.Run("SingleMessage", func(t *testing.T) {
		testSingleMessage(t, port)
	})

	t.Run("MultipleMessages", func(t *testing.T) {
		testMultipleMessages(t, port)
	})

	t.Run("LargeMessage", func(t *testing.T) {
		testLargeMessage(t, port)
	})

	t.Run("ConcurrentClients", func(t *testing.T) {
		testConcurrentClients(t, port)
	})

	t.Run("ConnectionClosed", func(t *testing.T) {
		testConnectionClosed(t, port)
	})
}

func testSingleMessage(t *testing.T, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	message := "Hello, Server!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	buffer := make([]byte, len(message))
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if n != len(message) || string(buffer) != message {
		t.Errorf("Expected response %q, got %q", message, string(buffer))
	}
}

func testMultipleMessages(t *testing.T, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	messages := []string{
		"First message",
		"Second message",
		"Third message",
	}

	for _, message := range messages {
		_, err = conn.Write([]byte(message))
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		buffer := make([]byte, len(message))
		n, err := conn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if n != len(message) || string(buffer) != message {
			t.Errorf("Expected response %q, got %q", message, string(buffer))
		}
	}
}

func testLargeMessage(t *testing.T, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create a message larger than the buffer size to test fragmentation
	messageSize := 4096 * 3 // This relies on the const being accessible
	message := generateRandomString(messageSize)

	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Failed to send large message: %v", err)
	}

	// Read the response in chunks
	var receivedData bytes.Buffer
	buffer := make([]byte, 1024)

	for receivedData.Len() < messageSize {
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			t.Fatalf("Failed to read response: %v", err)
		}
		if n > 0 {
			receivedData.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
	}

	if receivedData.String() != message {
		t.Errorf("Large message echo failed. Expected %d bytes, got %d bytes.",
			len(message), receivedData.Len())
	}
}

func testConcurrentClients(t *testing.T, port int) {
	const numClients = 10
	wg := &sync.WaitGroup{}
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", clientID, err)
				return
			}
			defer conn.Close()

			message := fmt.Sprintf("Hello from client %d!", clientID)
			_, err = conn.Write([]byte(message))
			if err != nil {
				t.Errorf("Client %d failed to send message: %v", clientID, err)
				return
			}

			buffer := make([]byte, len(message))
			n, err := conn.Read(buffer)
			if err != nil {
				t.Errorf("Client %d failed to read response: %v", clientID, err)
				return
			}

			if n != len(message) || string(buffer) != message {
				t.Errorf("Client %d expected response %q, got %q", clientID, message, string(buffer))
			}
		}(i)
	}

	wg.Wait()
}

func testConnectionClosed(t *testing.T, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	// Send a message and check response
	message := "About to close"
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	buffer := make([]byte, len(message))
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if n != len(message) || string(buffer) != message {
		t.Errorf("Expected response %q, got %q", message, string(buffer))
	}

	// Close the connection from client side
	conn.Close()

	// Wait a bit to ensure server has time to process the close
	time.Sleep(100 * time.Millisecond)

	// Try to connect again to verify server is still running
	newConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Server stopped accepting connections after client disconnect: %v", err)
	}
	defer newConn.Close()
}

// Helper function to get a free port
func getFreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to resolve TCP address: %v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

// Helper function to generate a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	sb := strings.Builder{}
	sb.Grow(length)

	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}
