package linereversal

import (
	"fmt"
	"net"
	"time"
)

// Interval for monitoring sessions for closure
// This prevents data leaks
const MonitorInterval = 3 * time.Minute

type SessionManager struct {
	sessions   map[int]*Session
	processors map[int]*LineProcessor
}

// Instructions for the sessions behavior
type SessionMessage struct {
	Type    string
	Data    []byte // Payload
	Number  int    // Either a position or length
	Address net.Addr
}

// Control messages
func CloseMessage(address net.Addr) SessionMessage {
	return SessionMessage{
		Type:    "close_client",
		Address: address,
	}
}

func AckMessage(length int, address net.Addr) SessionMessage {
	return SessionMessage{
		Type:    "recieved_ack",
		Number:  length,
		Address: address,
	}
}

func DataMessage(position int, data []byte, address net.Addr) SessionMessage {
	return SessionMessage{
		Type:    "recieved_data",
		Number:  position,
		Data:    data,
		Address: address,
	}
}

func WriteMessage(data []byte) SessionMessage {
	return SessionMessage{
		Type: "write_data",
		Data: data,
	}
}

func ConnectMessage(address net.Addr) SessionMessage {
	return SessionMessage{
		Type:    "connect_client",
		Address: address,
	}
}

// Create a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions:   make(map[int]*Session),
		processors: make(map[int]*LineProcessor),
	}
	go sm.MonitorSessions()
	return sm
}

// Sessions close themselves, so we monitor them every 3 minutes
func (sm *SessionManager) MonitorSessions() {
	for {
		for _, session := range sm.sessions {
			if session.IsClosed {
				close(session.Channel)
				delete(sm.sessions, session.ID)
				delete(sm.processors, session.ID)
			}
		}
		time.Sleep(MonitorInterval)
	}
}

// Create a session. Does nothing if it already exists
func (sm *SessionManager) CreateSession(conn net.PacketConn, address net.Addr, id int) {
	if session, ok := sm.sessions[id]; ok {
		// Session already exists
		if session.IsClosed {
			// Session is closed, so we can create a new one. Delete the existing first
			delete(sm.sessions, id)
			delete(sm.processors, id)
		} else {
			// Session is still open, so we do nothing
			return
		}
	}
	messageChannel := make(chan SessionMessage)
	dataChannel := make(chan []byte)
	sm.sessions[id] = NewSession(conn, address, id, messageChannel, dataChannel)
	sm.processors[id] = NewLineProcessor(id, messageChannel, dataChannel)
}

// Send a session a message
func (sm *SessionManager) SendMessage(id int, msg SessionMessage) error {
	session, ok := sm.sessions[id]
	if !ok {
		return fmt.Errorf("could not send message, session %d not found", id)
	}
	if session.IsClosed {
		return fmt.Errorf("could not send message, session %d is closed", id)
	}
	session.Channel <- msg
	return nil
}
