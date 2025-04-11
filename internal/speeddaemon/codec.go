package speeddaemon

import "bytes"

// Message type constants
const (
	ErrorMsgType      U8 = 0x10
	PlateMsgType      U8 = 0x20
	TicketMsgType     U8 = 0x21
	WantHeartbeatType U8 = 0x40
	HeartbeatType     U8 = 0x41
	IAmCameraType     U8 = 0x80
	IAmDispatcherType U8 = 0x81
	CameraSignalType  U8 = 0x00
)

type SdBuffer struct {
	Reader     *bytes.Reader
	ValidBytes int
}

func NewSdBuffer(originalBuffer []byte) *SdBuffer {
	buffer := make([]byte, len(originalBuffer))
	copy(buffer, originalBuffer)
	reader := bytes.NewReader(buffer)
	return &SdBuffer{
		Reader:     reader,
		ValidBytes: 0,
	}
}

func Decode(buffer *SdBuffer) (ClientMessage, error) {
	if buffer.Reader.Len() < 1 {
		return nil, ErrIncompleteMessage
	}

	msgType, err := readU8(buffer)
	if err != nil {
		return nil, err
	}

	var msg ClientMessage

	switch msgType {
	case ErrorMsgType:
		msg, err = decodeError(buffer)
	case PlateMsgType:
		msg, err = decodePlate(buffer)
	case TicketMsgType:
		msg, err = decodeTicket(buffer)
	case WantHeartbeatType:
		msg, err = decodeWantHeartbeat(buffer)
	case HeartbeatType:
		msg, err = decodeHeartbeat(buffer)
	case IAmCameraType:
		msg, err = decodeIAmCamera(buffer)
	case IAmDispatcherType:
		msg, err = decodeIAmDispatcher(buffer)
	default:
		return nil, ErrInvalidMessage
	}

	if err != nil {
		return nil, ErrIncompleteMessage
	}

	return msg, err
}

func ExtractFromSbBuffer(buffer *SdBuffer) ([]ClientMessage, int) {
	var messages []ClientMessage
	extractedBytes := 0

	for buffer.Reader.Len() > 0 {
		msg, err := Decode(buffer)
		if err == ErrInvalidMessage {
			messages = append(messages, &ErrorMessage{Content: "Invalid message"})
			break
		}
		if err != nil {
			break
		}
		messages = append(messages, msg)
		extractedBytes += buffer.ValidBytes
		buffer.ValidBytes = 0
	}

	return messages, extractedBytes
}
