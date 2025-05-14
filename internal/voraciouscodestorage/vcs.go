package voraciouscodestorage

import (
	"fmt"
	"log"
	"net"
	"os"
)

const (
	dataDirEnvVar       = "DATA_DIR" // Environment variable name
	localDefaultDataDir = "./data"   // Default relative path if env var missing
	dirPerms            = 0755       // Permissions if creating local dir
)

func Listen(port int) {
	// Listen for incoming connections on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()
	log.Printf("Server listening on port %d", port)
	if err != nil {
		log.Fatalf("Error creating file system: %v", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
}

func getDataDir() string {
	// Check if the environment variable is set
	dataDir := os.Getenv(dataDirEnvVar)
	if dataDir == "" {
		// If not set, use the default relative path
		dataDir = localDefaultDataDir
	}

	// Create the directory if it doesn't exist
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.MkdirAll(dataDir, dirPerms)
		if err != nil {
			log.Fatalf("Error creating data directory: %v", err)
		}
	}

	return dataDir
}
