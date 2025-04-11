package budgetchat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"unicode"
)

type user struct {
	name      string
	ctx       context.Context
	cancelCtx context.CancelFunc
	logger    *log.Logger
	conn      net.Conn
	broker    *broker
	scanner   *bufio.Scanner
	channel   chan string
}

func Listen(port int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	broker := &broker{
		users:   make(map[string]chan<- string, 0),
		channel: make(chan *brokerMessage, 50),
	}
	go broker.initateBroker()

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Could not start listener. REASON: " + err.Error())
	}
	log.Printf("Means to an end listening on port %d\n", port)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Encountered error accepting connection. REASON: " + err.Error())
			continue
		}

		go handleConnection(conn, broker)
	}
}

func handleConnection(conn net.Conn, broker *broker) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	logger := log.New(log.Writer(),
		fmt.Sprintf("[%s] ", conn.RemoteAddr().String()),
		log.Flags()|log.Lmsgprefix|log.Lshortfile)
	// Get the username
	userName, err := getUserName(conn, scanner)
	if err != nil {
		logger.Println(err)
		return
	}
	// Register the user
	userChannel := make(chan string, 10)
	broker.channel <- newBrokerMessage(register, userName, "", userChannel)
	// Join the chat and begin reading and writing
	ctx, cancelCtx := context.WithCancel(context.Background())
	user := user{
		name:      userName,
		ctx:       ctx,
		cancelCtx: cancelCtx,
		logger:    logger,
		conn:      conn,
		broker:    broker,
		scanner:   scanner,
		channel:   userChannel,
	}
	logger.Println("User joined the chat")
	go userReader(&user)
	userWriter(&user)
}

func getUserName(conn net.Conn, scanner *bufio.Scanner) (string, error) {
	// Send the entrance message
	_, err := conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	if err != nil {
		return "", fmt.Errorf("failed to ask client for name: %s", err)
	}
	// Get a line for user name
	gotSomething := scanner.Scan()
	if !gotSomething {
		return "", fmt.Errorf("couldn't scan name from client")
	}
	// Validate the username
	userName, err := validate(scanner.Bytes())
	if err != nil {
		message := fmt.Sprintf("Invalid name: %s\n", err)
		conn.Write([]byte(message))
		return "", errors.New(message)
	}
	return userName, nil
}

func userReader(user *user) {
	for {
		select {
		case <-user.ctx.Done():
			user.logger.Println("Reader done")
			return
		default:
			gotSomething := user.scanner.Scan()
			if !gotSomething {
				user.logger.Println("Unexpected error reading scan")
				user.cancelCtx()
				user.broker.channel <- newBrokerMessage(logoff, user.name, "", nil)
				return
			}
			message := user.scanner.Text()
			user.logger.Println(message)
			user.broker.channel <- newBrokerMessage(send, user.name, message, nil)
		}
	}
}

func userWriter(user *user) {
	for {
		select {
		case <-user.ctx.Done():
			user.logger.Println("Writer done")
			return
		case message := <-user.channel:
			_, err := user.conn.Write([]byte(message + "\n"))
			if err != nil {
				user.logger.Printf("%s: Error writing to client: %s", user.name, err)
				user.cancelCtx()
				user.broker.channel <- newBrokerMessage(logoff, user.name, "", nil)
				return
			}
		}
	}
}

func validate(rawInput []byte) (string, error) {
	name := strings.TrimSpace(string(rawInput))
	// Check if the name is long enough
	if len(name) < 1 {
		return "", fmt.Errorf("the name was too short")
	}
	// No longer than 16 character
	if len(name) > 16 {
		return "", fmt.Errorf("name length must be at most 16 characters, recieved %d", len(name))
	}
	// Check the name only contains alphanumerics
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return "", fmt.Errorf("name contained an illegal character %c", char)
		}
	}
	return name, nil
}
