package linereversal

import (
	"fmt"
	"strconv"
	"strings"
)

const maxInteger = 2147483647 // 2**31 - 1
const maxIntegerLength = 10   // The max number of digits in the max integer as a string
const maxMessageSize = 999

type LRMessage struct {
	Type     string
	Session  int
	Position int
	Data     []byte
	Length   int
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
	// Check we start and end with a slash
	if buffer[0] != '/' || buffer[len(buffer)-1] != '/' {
		return nil, fmt.Errorf("invalid message format: %s", buffer)
	}
	// Extract the message
	fields := []string{}
	var sb strings.Builder
	for i, char := range buffer {
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
			i++
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

// Packs a byte array into a data message. Returns the amount of bytes written
// As there is a limit of 1000 bytes per message
func PackDataMessage(message *LRMessage, data []byte, startPosition int) int {
	// The max int length for position, 5 slashes and the length of the session id
	// /data/SESSION/POS/DATA/
	maxDataSize := maxMessageSize - (maxIntegerLength + 5 + len(strconv.Itoa(message.Session)))
	escapedData := make([]byte, 0, len(data))
	dataLength := 0
	escapeCharacters := 0
	// We need to escape the data if it contains slashes
	for _, b := range data {
		if dataLength+escapeCharacters >= maxDataSize {
			break
		}
		if b == '/' || b == '\\' {
			// If we don't have any more room to pack two bytes, we need to break
			if dataLength+escapeCharacters+2 >= maxDataSize {
				break
			}
			// If we have room, we need to escape the character
			escapedData = append(escapedData, '\\')
			escapeCharacters += 1
		}
		escapedData = append(escapedData, b)
		dataLength += 1
	}
	message.Data = escapedData
	message.Position = startPosition
	return dataLength
}
