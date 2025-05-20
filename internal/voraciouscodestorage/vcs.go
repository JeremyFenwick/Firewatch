package voraciouscodestorage

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	// Create the file system
	dataDir := getDataDir()
	fs, err := NewFileSystem(dataDir)
	if err != nil {
		log.Fatalf("Error creating file system: %v", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn, fs)
	}
}

func handleConnection(conn net.Conn, fs *FileSystem) {
	defer conn.Close()
	// Handle the connection
	reader := bufio.NewReader(conn)
	for {
		// Send the ready message
		_, err := conn.Write([]byte("READY\n"))
		if err != nil {
			log.Printf("Error writing to connection: %v", err)
			return
		}
		// Read the command from the connection
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
		command = command[:len(command)-1] // Remove the newline character
		commandList := strings.Split(command, " ")
		// Match the command
		switch strings.ToUpper(commandList[0]) {
		case "HELP":
			err := handleHelp(conn)
			if err != nil {
				log.Printf("Error handling HELP command: %v", err)
				return
			}
		case "LIST":
			if len(commandList) < 2 {
				_, err := conn.Write([]byte("ERR usage: LIST dir\n"))
				if err != nil {
					log.Printf("Error writing to connection: %v", err)
				}
				continue
			}
			err := handleList(conn, fs, commandList)
			if err != nil {
				log.Printf("Error handling list request: %v", err)
				return
			}
		case "GET":
			if len(commandList) < 2 {
				_, err := conn.Write([]byte("ERR usage: GET file [revision]\n"))
				if err != nil {
					log.Printf("Error writing to connection: %v", err)
				}
				continue
			}
			err := handleGet(conn, fs, commandList)
			if err != nil {
				log.Printf("Error handling get request: %v", err)
				return
			}
		case "PUT":
			if len(commandList) < 3 {
				_, err := conn.Write([]byte("ERR usage: PUT file length newline data\n"))
				if err != nil {
					log.Printf("Error writing to connection: %v", err)
					return
				}
			}
			err = handlePut(conn, fs, commandList)
			if err != nil {
				log.Printf("Error handling put request: %v", err)
				return
			}
		case "CLEAR":
			fs.Clear()
			_, err := conn.Write([]byte("OK cleared fs contents\n"))
			if err != nil {
				log.Printf("Error writing to connection: %v", err)
				return
			}
		default:
			_, err := conn.Write([]byte("ERR illegal method: " + commandList[0] + "\n"))
			if err != nil {
				log.Printf("Error writing illegal method  to connection: %v", err)
				return
			}
			return
		}
	}
}

func handlePut(conn net.Conn, fs *FileSystem, commandList []string) error {
	// Check if the file name is valid
	if !isValidPath(commandList[1]) {
		_, err := conn.Write([]byte("ERR illegal file name\n"))
		if err != nil {
			return fmt.Errorf("error writing to connection: %v", err)
		}
	}
	fileName := commandList[1]
	// Get the read limit
	readLimit, err := strconv.Atoi(commandList[2])
	if err != nil {
		readLimit = 0
	}
	// Create the file
	limitReader := io.LimitReader(conn, int64(readLimit))
	file, err := fs.AddFile(limitReader, fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	// Write the version number back to the connection
	_, err = conn.Write([]byte(fmt.Sprintf("OK r%d\n", file.LatestVersion)))
	return nil
}

func handleGet(conn net.Conn, fs *FileSystem, commandList []string) error {
	// Check if the file name is valid
	if !isValidPath(commandList[1]) {
		_, err := conn.Write([]byte("ERR illegal file name\n"))
		if err != nil {
			return fmt.Errorf("error writing to connection: %v", err)

		}
	}
	fullPath := commandList[1]
	folderName, fileName := splitDirFile(fullPath)
	// Get the target folder
	targetFolder, err := fs.GetFolder(folderName)
	if err != nil {
		_, err := conn.Write([]byte("ERR no such file\n"))
		if err != nil {
			return fmt.Errorf("error writing to connection: %v", err)
		}
		return nil
	}
	// Check if the file exists
	if !targetFolder.HasChildFile(fileName) {
		_, err := conn.Write([]byte("ERR no such file\n"))
		if err != nil {
			return fmt.Errorf("error writing to connection: %v", err)
		}
		return nil
	}
	// Set the revision if provided
	revision := targetFolder.Files[fileName].LatestVersion
	if len(commandList) == 3 {
		// Parse the revision number
		parsedRevision, err := strconv.Atoi(commandList[2])
		if err != nil || parsedRevision < 0 || parsedRevision > targetFolder.Files[fileName].LatestVersion {
			_, err = conn.Write([]byte("err no such revision\n"))
			if err != nil {
				return fmt.Errorf("error writing to connection: %v", err)
			}
			return nil
		}
		revision = parsedRevision
	}
	// Read the file
	targetFolder.ReadFile(fileName, revision, conn)
	return nil
}

func handleHelp(conn net.Conn) error {
	helpMessage := "OK Usage: HELP|GET|PUT|LIST"
	_, err := conn.Write([]byte(helpMessage + "\n"))
	if err != nil {
		return err
	}
	return nil
}

func handleList(conn net.Conn, fs *FileSystem, commandList []string) error {
	// Check if the directory is valid
	if !isValidPath(commandList[1]) {
		_, err := conn.Write([]byte("ERR illegal dir name\n"))
		if err != nil {
			return fmt.Errorf("error writing to connection: %v", err)
		}
	}
	targetDir := commandList[1]
	// Get the target folder
	targetFolder, err := fs.GetFolder(targetDir)
	if err != nil || targetFolder.IsEmpty() {
		_, err := conn.Write([]byte("OK 0\n"))
		if err != nil {
			return err
		}
		return nil
	}
	// Send all files in the folder
	for _, file := range targetFolder.GetChildAllFiles() {
		_, err := conn.Write([]byte(file.Name + "\n"))
		if err != nil {
			return err
		}
	}
	// Send all folders in the folder
	for _, folder := range targetFolder.GetChildAllFolders() {
		_, err := conn.Write([]byte(folder.Name + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
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

func isValidPath(p string) bool {
	// Must be absolute
	if !filepath.IsAbs(p) {
		return false
	}
	// Must not contain double slashes (except the leading one, which is part of Unix absolute paths)
	if strings.Contains(p[1:], "//") {
		return false
	}
	// Must not be empty
	if p == "" {
		return false
	}
	return true
}

func splitDirFile(fullPath string) (string, string) {
	lastSlash := strings.LastIndex(fullPath, string(os.PathSeparator))
	return fullPath[:lastSlash+1], fullPath[lastSlash+1:]
}
