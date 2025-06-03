package voraciouscodestorage

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
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
	log.Println("Data directory:", dataDir)
	fs, err := NewFileManager(dataDir)
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

func handleConnection(conn net.Conn, fm *FileManager) {
	defer conn.Close()
	// Send the ready message
	err := sendMessage(conn, "READY", false)
	if err != nil {
		return
	}
	// Handle the connection
	reader := bufio.NewReader(conn)
	for {
		// Read the command from the connection
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
		log.Printf("Received command: %s", command)
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
				err := sendMessage(conn, "ERR usage: LIST dir", true)
				if err != nil {
					return
				}
				continue
			}
			err := handleList(conn, fm, commandList)
			if err != nil {
				log.Printf("Error handling list request: %v", err)
				return
			}
		case "GET":
			err := handleGet(conn, fm, commandList)
			if err != nil {
				log.Printf("Error handling get request: %v", err)
				return
			}
		case "PUT":
			err = handlePut(reader, conn, fm, commandList)
			if err != nil {
				log.Printf("Error handling put request: %v", err)
				return
			}
		case "CLEAR":
			fm.Clear()
			err := sendMessage(conn, "OK cleared fs contents", true)
			if err != nil {
				return
			}
		default:
			sendMessage(conn, fmt.Sprintf("ERR illegal method %s", commandList[0]), false)
			return
		}
	}
}

func handlePut(r io.Reader, conn net.Conn, fm *FileManager, commandList []string) error {
	// Validate the command
	if len(commandList) != 3 {
		return sendMessage(conn, "ERR usage: PUT file length newline data", true)
	}
	// Check if the file name is valid
	if !isValidPath(commandList[1]) {
		err := sendMessage(conn, "ERR illegal file name", false)
		if err != nil {
			return err
		}
		return nil
	}
	fileName := commandList[1]
	// Get the read limit. Default to 0 is fine
	readLimit, _ := strconv.Atoi(commandList[2])
	// Create the file
	limitReader := io.LimitReader(r, int64(readLimit))
	file, err := fm.AddFile(fileName, limitReader, readLimit)
	if err == ErrNonTextData {
		return sendMessage(conn, "ERR text files only", true)
	}
	if err != nil {
		return fmt.Errorf("error adding file: %v", err)
	}
	// Write the version number back to the connection
	return sendMessage(conn, fmt.Sprintf("OK r%d", file.LatestVersion), true)
}

func handleGet(conn net.Conn, fm *FileManager, commandList []string) error {
	sendNoSuchFile := func() error {
		return sendMessage(conn, "ERR no such file", true)
	}
	// Validate the command
	if len(commandList) < 2 || len(commandList) > 3 {
		return sendMessage(conn, "ERR usage: GET file [revision]", true)
	}
	// Check if the file name is valid
	if !isValidPath(commandList[1]) {
		return sendMessage(conn, "ERR illegal file name", false)
	}
	fullPath := commandList[1]
	dir, fileName := splitDirFile(fullPath)
	// Get the target folder
	targetFolder, err := fm.GetFolder(dir)
	if err != nil {
		return sendNoSuchFile()
	}
	// Check if there was a prior error or the file doesn't exist
	_, exists := targetFolder.Files[fileName]
	if !exists {
		return sendNoSuchFile()
	}
	// Set the revision if provided
	revision := targetFolder.Files[fileName].LatestVersion
	if len(commandList) == 3 {
		revisionInput := commandList[2]
		// Check if the revision number starts with 'r'
		if strings.HasPrefix(commandList[2], "r") {
			revisionInput = revisionInput[1:] // Remove the 'r' prefix
		}
		// Parse the revision number
		parsedRevision, err := strconv.Atoi(revisionInput)
		if err != nil || parsedRevision < 1 || parsedRevision > targetFolder.Files[fileName].LatestVersion {
			return sendMessage(conn, "ERR no such revision", true)
		}
		revision = parsedRevision
	}
	// Read the file
	targetFile := targetFolder.Files[fileName].Files[revision-1]
	err = sendMessage(conn, fmt.Sprintf("OK %d", targetFile.Bytes), false)
	if err != nil {
		return err
	}
	err = targetFolder.Files[fileName].Files[revision-1].ReadFile(conn)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	return sendMessage(conn, "READY", false)
}

func handleHelp(conn net.Conn) error {
	return sendMessage(conn, "OK Usage: HELP|GET|PUT|LIST", true)
}

func handleList(conn net.Conn, fm *FileManager, commandList []string) error {
	// Check if the directory is valid
	if !isValidPath(commandList[1]) {
		return sendMessage(conn, "ERR illegal dir name", false)
	}
	targetDir := commandList[1]
	// Get the target folder
	targetFolder, err := fm.GetFolder(targetDir)
	if err != nil {
		return sendMessage(conn, "OK 0", true)
	}
	itemCount := len(targetFolder.Files) + len(targetFolder.SubFolders)
	// Send the number of items in the folder
	err = sendMessage(conn, fmt.Sprintf("OK %d", itemCount), false)
	if err != nil {
		return err
	}
	// Get the folders and sort them
	folders := targetFolder.GetSubFolders()
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})
	// Send all folders in the folder
	for _, folder := range folders {
		err := sendMessage(conn, folder.Name+"/ DIR", false)
		if err != nil {
			return err
		}
	}
	// Get the files and sort them
	files := targetFolder.GetFiles()
	sort.Slice(files, func(i, j int) bool {
		return files[i].FileName < files[j].FileName
	})
	// Send all files in the folder
	for _, file := range files {
		err := sendMessage(conn, fmt.Sprintf("%s r%d", file.FileName, file.LatestVersion), false)
		if err != nil {
			return err
		}
	}
	return sendMessage(conn, "READY", false)
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
	isLegal := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' || r == '/' || r == '\\'
	}
	// Must be absolute
	if !filepath.IsAbs(p) {
		return false
	}
	// Must not contain double slashes (except the leading one, which is part of Unix absolute paths)
	if strings.Contains(p, "//") {
		return false
	}
	// Must not contain illegal characters
	for _, r := range p {
		if !isLegal(r) {
			return false
		}
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

func sendMessage(conn net.Conn, message string, readyFollowUp bool) error {
	// Send a message to the connection
	log.Printf("Sending message: %s", message)
	_, err := conn.Write([]byte(message + "\n"))
	if err != nil {
		log.Printf("Error writing to connection: %v", err)
		return fmt.Errorf("error writing to connection: %v", err)
	}
	if readyFollowUp {
		return sendMessage(conn, "READY", false)
	}
	return nil
}
