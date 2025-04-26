package linereversal

import (
	"fmt"
	"strconv"
	"strings"
)

const maxInteger = 2147483647 // 2**31 - 1
const maxMessageSize = 999

type LRMessage struct {
	Type     string
	Session  int
	Position int
	Data     []byte
	Length   int
}

func (m *LRMessage) String() string {
	if m.Type == "data" {
		containsNewline := strings.Contains(string(m.Data), "\n")
		return fmt.Sprintf("Data message. Session: %d, Position: %d, Contains Newline: %t ", m.Session, m.Position, containsNewline)
	} else {
		return fmt.Sprintf("Message type: %s, Session: %d, Position: %d, Length: %d", m.Type, m.Session, m.Position, m.Length)
	}
}

func (m *LRMessage) Validate() bool {
	if m.Session < 0 || m.Session > maxInteger {
		return false
	}
	if m.Type == "connect" || m.Type == "close" {
		if m.Position != 0 || m.Length != 0 || len(m.Data) != 0 {
			return false
		}
	}
	if m.Type == "ack" {
		if m.Length < 0 || m.Length > maxInteger {
			return false
		}
		if m.Position != 0 {
			return false
		}
	}
	if m.Type == "data" {
		if m.Position > maxInteger || m.Position < 0 {
			return false
		}
		if m.Length != 0 {
			return false
		}
		totalDataLength := m.Position + len(m.Data)
		if totalDataLength > maxInteger {
			return false
		}
	}
	return true
}

func (m *LRMessage) Encode(buffer []byte) (int, error) {
	var data []byte
	switch m.Type {
	case "connect":
		data = []byte(fmt.Sprintf("/connect/%d/", m.Session))
	case "ack":
		data = []byte(fmt.Sprintf("/ack/%d/%d/", m.Session, m.Length))
	case "data":
		data = []byte(fmt.Sprintf("/data/%d/%d/%s/", m.Session, m.Position, m.Data))
	case "close":
		data = []byte(fmt.Sprintf("/close/%d/", m.Session))
	default:
		return 0, fmt.Errorf("invalid message type: %s", m.Type)
	}
	copied := copy(buffer, data)
	return copied, nil
}

func DecodeLRMessage(buffer []byte) (*LRMessage, error) {
	// If the buffer is empty, return an error
	if len(buffer) == 0 {
		return nil, fmt.Errorf("empty message")
	}
	// Check we start and end with a slash
	if buffer[0] != '/' || buffer[len(buffer)-1] != '/' {
		return nil, fmt.Errorf("invalid message format: %s", buffer)
	}
	// Extract the message
	fields := []string{}
	var sb strings.Builder
	for i := 0; i < len(buffer); i++ {
		char := buffer[i]
		// If we encounter an unescaped slash, we should split the string
		if char == '/' && (i == 0 || buffer[i-1] != '\\') {
			if sb.Len() == 0 {
				continue
			}
			fields = append(fields, sb.String())
			sb.Reset()
			continue
		}
		// If we encounter an escaped slash, we should just add it to the string
		if char == '\\' && i < len(buffer)-1 && (buffer[i+1] == '/' || buffer[i+1] == '\\') {
			sb.WriteByte(buffer[i+1])
			i += 1
			continue
		}
		// Otherwise, we just add the character to the string
		sb.WriteByte(char)
	}
	message, err := constructLRMessage(fields)
	if err != nil {
		return nil, err
	}
	if !message.Validate() {
		return nil, fmt.Errorf("invalid message: %v", message)
	}
	return message, nil
}

func constructLRMessage(fields []string) (*LRMessage, error) {
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid message format")
	}
	messageType := fields[0]
	session, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %s", fields[1])
	}
	message := &LRMessage{
		Type:    messageType,
		Session: session,
	}
	if messageType == "ack" {
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid ack message format")
		}
		length, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("invalid length: %s", fields[2])
		}
		message.Length = length
	} else if messageType == "data" {
		if len(fields) != 4 {
			return nil, fmt.Errorf("invalid data message format")
		}
		position, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("invalid position: %s", fields[2])
		}
		data := []byte(fields[3])
		message.Position = position
		message.Data = data
	}
	return message, nil
}

// Packs a byte array into a data message. Returns the amount of *unescaped* bytes written.
func PackDataMessage(message *LRMessage, data []byte, startPosition int) int {
	sessionStr := strconv.Itoa(message.Session)
	positionStr := strconv.Itoa(startPosition)

	// Calculate precise overhead: /data/ (6) + / (1) + / (1) + / (1) + len(sessionStr) + len(positionStr)
	overhead := 9 + len(sessionStr) + len(positionStr)
	maxDataSize := maxMessageSize - overhead  // Max length of the *escaped* data field
	escapedData := make([]byte, 0, len(data)) // Initial capacity guess
	currentEscapedLen := 0
	dataLength := 0 // Unescaped bytes consumed

	for _, b := range data {
		bytesToAdd := 1
		needsEscape := false
		if b == '/' || b == '\\' {
			bytesToAdd = 2
			needsEscape = true
		}

		// Check if adding the next byte (escaped or not) exceeds the limit
		if currentEscapedLen+bytesToAdd >= maxDataSize {
			break // Cannot add this byte, stop processing
		}

		// Add the byte(s)
		if needsEscape {
			escapedData = append(escapedData, '\\')
		}
		escapedData = append(escapedData, b)
		currentEscapedLen += bytesToAdd
		dataLength += 1 // Increment count of original bytes processed
	}

	message.Data = escapedData
	message.Position = startPosition
	return dataLength
}
