package primetime_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/JeremyFenwick/firewatch/internal/primetime" // Replace with your actual module path
)

func TestIsPrime(t *testing.T) {
	t.Run("First prime", func(t *testing.T) {
		testIsPrime(t, 11, true)
	})

	t.Run("Second prime", func(t *testing.T) {
		testIsPrime(t, 7919, true)
	})

	t.Run("First non-prime", func(t *testing.T) {
		testIsPrime(t, 12, false)
	})

	t.Run("Second non-prime", func(t *testing.T) {
		testIsPrime(t, 7921, false)
	})

	// Start the server in a goroutine (assuming it's not already running)
	port := 5001
	go primetime.Listen(port) // Replace 'primetime' with your actual package name

	// Give the server a little time to start
	time.Sleep(100 * time.Millisecond) // Adjust as needed

	t.Run("Test server response", func(t *testing.T) {
		testPrimeNumberCheck(t, port, 7, true)
	})

	t.Run("Test server response", func(t *testing.T) {
		testPrimeNumberCheck(t, port, 4, false)
	})
}

func testIsPrime(t *testing.T, number int, isPrime bool) {
	result := primetime.IsPrime(number)
	if result != isPrime {
		t.Errorf("Number %d. Expected result to be %t, but got %t instead.", number, isPrime, result)
	}

}

func testPrimeNumberCheck(t *testing.T, port, number int, isPrime bool) {
	// Connect to the server
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// JSON request to send
	requestJSON := fmt.Sprintf(`{"method":"isPrime","number":%d}`, number)

	// Send the request
	_, err = fmt.Fprint(conn, requestJSON+"\n")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	// Read the response
	reader := bufio.NewReader(conn)
	responseJSON, err := reader.ReadString('\n') // Assuming your server sends a newline
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	responseJSON = strings.TrimSpace(responseJSON)

	// Expected response
	expectedResponse := fmt.Sprintf(`{"method":"isPrime","prime":%t}`, isPrime)

	// Verify the response
	if responseJSON != expectedResponse {
		t.Errorf("Unexpected response:\nGot:  %q\nWant: %q", responseJSON, expectedResponse)
	}
}
