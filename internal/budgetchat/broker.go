package budgetchat

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type broker struct {
	mutex   sync.Mutex
	users   map[string]chan<- string
	channel chan *brokerMessage
}

type messageType int

const (
	register messageType = iota
	send
	logoff
)

type brokerMessage struct {
	op            messageType
	sender        string
	payload       string
	senderChannel chan<- string
}

func newBrokerMessage(op messageType, sender, payload string, senderChannel chan<- string) *brokerMessage {
	return &brokerMessage{
		op:            op,
		sender:        sender,
		payload:       payload,
		senderChannel: senderChannel,
	}
}

func (b *broker) initateBroker() {
	for {
		message := <-b.channel
		switch message.op {
		case register:
			registerUser(b, message.sender, message.senderChannel)
		case send:
			userSend(b, message.sender, message.payload)
		case logoff:
			userLogoff(b, message.sender)
		default:
			log.Printf("Unknown message type received: %d", message.op)
		}
	}
}

func registerUser(b *broker, name string, userChannel chan<- string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.users[name]; exists {
		// User has already registered
		return fmt.Errorf("user %s is already registered", name)
	}
	// Build the current active users list. We are protected by the mutex lock
	i := 0
	activeUsers := make([]string, len(b.users))
	for user, channel := range b.users {
		activeUsers[i] = user
		channel <- fmt.Sprintf("* %s has entered the room", name)
		i++
	}
	activeUsersList := strings.Join(activeUsers, ", ")
	// Add the users list message to the new members queue
	if len(activeUsers) > 0 {
		userChannel <- fmt.Sprintf("* The room contains: %s", activeUsersList)
	} else {
		userChannel <- "* The room is empty"
	}
	b.users[name] = userChannel
	return nil
}

// Sends the provided message to all except the sender
func userSend(b *broker, sender string, message string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	outputMessage := fmt.Sprintf("[%s] %s", sender, message)
	for user, userChannel := range b.users {
		if user == sender {
			// Don't try and send a message to yourself
			continue
		}
		userChannel <- outputMessage
	}
}

// Logoff removes name from the Users map
func userLogoff(b *broker, existedUser string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	delete(b.users, existedUser)
	outputMessage := fmt.Sprintf("* %s has left the room", existedUser)
	// Now tell all users that
	for _, userChannel := range b.users {
		userChannel <- outputMessage
	}
}
