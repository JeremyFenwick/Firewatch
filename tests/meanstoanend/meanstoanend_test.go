package meanstoanend_test

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/JeremyFenwick/firewatch/internal/meanstoanend"
	// Replace with your actual module path
)

func TestMeansToAnEnd(t *testing.T) {
	// Start the server in a goroutine (assuming it's not already running)
	port := 5002
	go meanstoanend.Listen(port) // Replace 'primetime' with your actual package name

	// Give the server a little time to start
	time.Sleep(100 * time.Millisecond) // Adjust as needed

	t.Run("Test server response", func(t *testing.T) {
		testSession(t, port)
	})
}

func testSession(t *testing.T, port int) {
	// Connect to the server
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Helper function to send and log
	sendMessage := func(desc string, msg []byte) {
		t.Log(desc)
		_, err := conn.Write(msg)
		if err != nil {
			t.Fatalf("Failed to write %s: %v", desc, err)
		}
		t.Logf("Sent %d bytes: %v", len(msg), msg) // Added logging
	}

	sendMessage("Sending insert for time 12345", createInsertMessage(12345, 101))
	sendMessage("Sending insert for time 12346", createInsertMessage(12346, 102))
	sendMessage("Sending insert for time 12347", createInsertMessage(12347, 100))
	sendMessage("Sending query for range 12288-16843", createQueryMessage(12288, 16843))

	// Read the response with a timeout
	t.Log("Waiting for response...")
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buffer := make([]byte, 4)
	_, err = conn.Read(buffer)
	if err != nil {
		t.Fatalf("Could not read response: %v", err)
	}

	// Check for equality
	result := binary.BigEndian.Uint32(buffer[:])
	t.Logf("Received result: %d", result)
	if result != 101 {
		t.Fatalf("Did not get the expected query result of 101. Instead got: %d", result)
	}
}

func createInsertMessage(time, number int32) []byte {
	messageType := byte('I')
	message := make([]byte, 9)
	message[0] = messageType
	binary.BigEndian.PutUint32(message[1:5], uint32(time))
	binary.BigEndian.PutUint32(message[5:9], uint32(number))
	return message
}

func createQueryMessage(start, end int32) []byte {
	messageType := byte('Q')
	message := make([]byte, 9)
	message[0] = messageType
	binary.BigEndian.PutUint32(message[1:5], uint32(start))
	binary.BigEndian.PutUint32(message[5:9], uint32(end))
	return message
}
