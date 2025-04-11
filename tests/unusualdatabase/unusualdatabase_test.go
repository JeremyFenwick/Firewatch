package unusualdatabase_test

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/JeremyFenwick/firewatch/internal/unusualdatabase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPort = 4321
	timeout  = 500 * time.Millisecond
)

// TestMain manages the test suite setup and teardown
func TestMain(m *testing.M) {
	// Start the server once for all tests
	go func() {
		unusualdatabase.Listen(testPort)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Run all tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}

func createUDPClient(t *testing.T) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", testPort))
	require.NoError(t, err)

	conn, err := net.DialUDP("udp", nil, addr)
	require.NoError(t, err)

	err = conn.SetDeadline(time.Now().Add(timeout))
	require.NoError(t, err)

	return conn
}

func TestBasicInsertAndRetrieve(t *testing.T) {
	client := createUDPClient(t)
	defer client.Close()

	// Insert a key-value pair
	key, value := "testKey1", "testValue1"
	_, err := client.Write([]byte(fmt.Sprintf("%s=%s", key, value)))
	require.NoError(t, err)

	// Small delay to ensure server processes the request
	time.Sleep(50 * time.Millisecond)

	// Request the value for the key
	_, err = client.Write([]byte(key))
	require.NoError(t, err)

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(t, err)

	// Verify response
	expected := fmt.Sprintf("%s=%s", key, value)
	assert.Equal(t, expected, string(buffer[:n]))
}

func TestMultipleInserts(t *testing.T) {
	client := createUDPClient(t)
	defer client.Close()

	// Insert multiple key-value pairs
	testData := map[string]string{
		"multi_key1": "value1",
		"multi_key2": "value2",
		"multi_key3": "value3",
	}

	for k, v := range testData {
		_, err := client.Write([]byte(fmt.Sprintf("%s=%s", k, v)))
		require.NoError(t, err)
		time.Sleep(20 * time.Millisecond)
	}

	// Verify all inserts
	for k, v := range testData {
		_, err := client.Write([]byte(k))
		require.NoError(t, err)

		buffer := make([]byte, 1024)
		n, err := client.Read(buffer)
		require.NoError(t, err)

		expected := fmt.Sprintf("%s=%s", k, v)
		assert.Equal(t, expected, string(buffer[:n]))
	}
}

func TestNonExistentKey(t *testing.T) {
	client := createUDPClient(t)
	defer client.Close()

	// Request a non-existent key
	key := "nonExistentKey"
	_, err := client.Write([]byte(key))
	require.NoError(t, err)

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(t, err)

	// Verify response for non-existent key
	expected := fmt.Sprintf("%s=", key)
	assert.Equal(t, expected, string(buffer[:n]))
}

func TestProtectedVersionKey(t *testing.T) {
	client := createUDPClient(t)
	defer client.Close()

	// Try to modify the protected version key
	_, err := client.Write([]byte("version=hackedVersion"))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// Request the version
	_, err = client.Write([]byte("version"))
	require.NoError(t, err)

	// Read the response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(t, err)

	// Verify that version remains unchanged
	assert.Equal(t, "version=madvillains vault of villainy", string(buffer[:n]))
}

func TestConcurrentRequests(t *testing.T) {
	// Create multiple clients
	numClients := 5
	clients := make([]*net.UDPConn, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = createUDPClient(t)
		defer clients[i].Close()
	}

	// Each client inserts unique data
	for i, client := range clients {
		key := fmt.Sprintf("concurrent_key_%d", i)
		value := fmt.Sprintf("concurrent_value_%d", i)

		_, err := client.Write([]byte(fmt.Sprintf("%s=%s", key, value)))
		require.NoError(t, err)
	}

	time.Sleep(100 * time.Millisecond)

	// Each client reads back its data
	for i, client := range clients {
		key := fmt.Sprintf("concurrent_key_%d", i)
		value := fmt.Sprintf("concurrent_value_%d", i)

		_, err := client.Write([]byte(key))
		require.NoError(t, err)

		buffer := make([]byte, 1024)
		n, err := client.Read(buffer)
		require.NoError(t, err)

		expected := fmt.Sprintf("%s=%s", key, value)
		assert.Equal(t, expected, string(buffer[:n]))
	}
}
