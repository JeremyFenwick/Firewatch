package budgetchat_test

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/JeremyFenwick/firewatch/internal/budgetchat"
	// Replace with your actual module path
)

func TestChat(t *testing.T) {
	// Start the server in a goroutine (assuming it's not already running)
	port := 5003
	go budgetchat.Listen(port) // Replace 'primetime' with your actual package name

	// Give the server a little time to start
	time.Sleep(100 * time.Millisecond) // Adjust as needed

	t.Run("Test server response", func(t *testing.T) {
		testSession(t, port)
	})
}

func testSession(t *testing.T, port int) {
	// Connect to the server
	jeremyConn := createUser(t, port)
	aliceConn := createUser(t, port)
	defer jeremyConn.Close()
	// Get the welcome message and submit username
	enterChat("jeremy", t, jeremyConn)
	enterChat("alice", t, aliceConn)
	// Have alice send a message and jeremy recieve
	aliceConn.Write([]byte("Hi jeremy\n"))
	jeremyReader := bufio.NewReader(jeremyConn)
	aliceMessage, _, err := jeremyReader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to recieve message from alice: %v", err)
	}
	if string(aliceMessage) != "* alice has entered the room" {
		t.Fatalf("Failed to recieve expected message from alice: %s", err.Error())
	}
	aliceMessage, _, err = jeremyReader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to recieve message from alice: %v", err.Error())
	}
	if string(aliceMessage) != "[alice] Hi jeremy" {
		t.Fatalf("Failed to recieve expected message from alice: %s", err.Error())
	}
	// Have alice leave
	aliceConn.Close()
	// Have jeremy recieve the exit message
	aliceMessage, _, err = jeremyReader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to recieve message from alice: %v", err.Error())
	}
	if string(aliceMessage) != "* alice has left the room" {
		t.Fatalf("Failed to recieve expected message from alice: %s", err.Error())
	}
}

func createUser(t *testing.T, port int) net.Conn {
	// Connect to the server
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err.Error())
	}
	return conn
}

func enterChat(username string, t *testing.T, conn net.Conn) {
	reader := bufio.NewReader(conn)
	welcomeMessage, _, err := reader.ReadLine()
	if err != nil || string(welcomeMessage) != "Welcome to budgetchat! What shall I call you?" {
		t.Fatalf("Could not recieve welcome message from server")
	}
	conn.Write([]byte(username + "\n"))
	_, _, err = reader.ReadLine()
	if err != nil {
		t.Fatalf("Could not recieve room entrance message")
	}
}
