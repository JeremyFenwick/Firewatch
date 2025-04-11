package speeddaemon_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/JeremyFenwick/firewatch/internal/speeddaemon" // Replace with your actual module path
	"github.com/stretchr/testify/assert"
)

func TestErrorMessage(t *testing.T) {
	errorMessage := &speeddaemon.ErrorMessage{
		Content: "Test error message",
	}
	encodedMessage, err := errorMessage.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err, "Decoding should not produce an error")

	msg := decodedMessage.(*speeddaemon.ErrorMessage)
	assert.Equal(t, speeddaemon.ErrorMsgType, decodedMessage.GetType(), "Message type should match")
	assert.Equal(t, errorMessage, msg)
}

func TestPlateMessage(t *testing.T) {
	plateMessage := &speeddaemon.PlateMessage{
		Plate:     "ABC123",
		Timestamp: 1234567890,
	}
	encodedMessage, err := plateMessage.Encode()
	assert.NoError(t, err)

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err)

	msg := decodedMessage.(*speeddaemon.PlateMessage)
	assert.Equal(t, speeddaemon.PlateMsgType, decodedMessage.GetType())
	assert.Equal(t, plateMessage, msg)

}

func TestTicketMessage(t *testing.T) {
	ticketMessage := &speeddaemon.TicketMessage{
		Plate:        "XYZ789",
		Road:         16,
		MileOne:      100,
		TimeStampOne: 8000,
		MileTwo:      200,
		TimeStampTwo: 8400,
		Speed:        120,
	}
	encodedMessage, err := ticketMessage.Encode()
	assert.NoError(t, err)

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err)

	msg := decodedMessage.(*speeddaemon.TicketMessage)
	assert.Equal(t, speeddaemon.TicketMsgType, decodedMessage.GetType())
	assert.Equal(t, ticketMessage, msg)
}

func TestWantHeartbeatMessage(t *testing.T) {
	wantHeartbeatMessage := &speeddaemon.WantHeartbeatMessage{
		Interval: 25,
	}
	encodedMessage, err := wantHeartbeatMessage.Encode()
	assert.NoError(t, err)

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err)

	msg := decodedMessage.(*speeddaemon.WantHeartbeatMessage)
	assert.Equal(t, speeddaemon.WantHeartbeatType, decodedMessage.GetType())
	assert.Equal(t, wantHeartbeatMessage, msg)
}

func TestHeartbeatMessage(t *testing.T) {
	heartbeatMessage := &speeddaemon.HeartbeatMessage{}
	encodedMessage, err := heartbeatMessage.Encode()
	assert.NoError(t, err)

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err)

	assert.Equal(t, speeddaemon.HeartbeatType, decodedMessage.GetType())
}

func TestIAmCameraMessage(t *testing.T) {
	cameraMessage := &speeddaemon.IAmCameraMessage{
		Road:  120,
		Mile:  100,
		Limit: 80,
	}
	encodedMessage, err := cameraMessage.Encode()
	assert.NoError(t, err)

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err)

	msg := decodedMessage.(*speeddaemon.IAmCameraMessage)
	assert.Equal(t, cameraMessage, msg)
}

func TestIAmDispatcherMessage(t *testing.T) {
	dispatcherMessage := &speeddaemon.IAmDispatcherMessage{
		Numroads: 5,
		Roads:    []speeddaemon.U16{1, 2, 3, 4, 5},
	}
	encodedMessage, err := dispatcherMessage.Encode()
	assert.NoError(t, err)

	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	decodedMessage, err := speeddaemon.Decode(buffer)
	assert.NoError(t, err)

	msg := decodedMessage.(*speeddaemon.IAmDispatcherMessage)
	assert.Equal(t, dispatcherMessage, msg)
}

func TestExtractor(t *testing.T) {
	errorMessage := &speeddaemon.ErrorMessage{
		Content: "Test error message",
	}
	encodedMessage, err := errorMessage.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	buffer := speeddaemon.NewSdBuffer(encodedMessage)
	messages, extractedBytes := speeddaemon.ExtractFromSbBuffer(buffer)
	assert.Equal(t, 1, len(messages), "Should extract one message")
	assert.Equal(t, len(encodedMessage), extractedBytes, "Extracted bytes should match encoded message length")
	assert.Equal(t, errorMessage, messages[0], "Extracted message should match original")
}

func TestIncompleteMessage(t *testing.T) {
	// Create a buffer with an incomplete message
	incompleteBuffer := speeddaemon.NewSdBuffer([]byte{0x01, 0x02}) // Incomplete message
	messages, extractedBytes := speeddaemon.ExtractFromSbBuffer(incompleteBuffer)
	assert.Equal(t, 1, len(messages), "There should be one message")
	assert.Equal(t, 0, extractedBytes, "No bytes should be extracted from an incomplete message")
}

func TestPartialMessage(t *testing.T) {
	errorMessage := &speeddaemon.ErrorMessage{
		Content: "Test error message",
	}
	encodedMessage, err := errorMessage.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	incompleteBuffer := append(encodedMessage, 0x80, 0x00, 0x00) // Incomplete message
	buffer := speeddaemon.NewSdBuffer(incompleteBuffer)
	messages, extractedBytes := speeddaemon.ExtractFromSbBuffer(buffer)
	assert.Equal(t, 1, len(messages), "Should extract one message")
	assert.Equal(t, len(encodedMessage), extractedBytes, "Extracted bytes should match encoded message length")
	assert.Equal(t, errorMessage, messages[0], "Extracted message should match original")
	remainingBytes := incompleteBuffer[extractedBytes:]
	assert.Equal(t, 3, len(remainingBytes), "Remaining bytes should be the rest of the buffer")
}

func TestCalculateSpeed(t *testing.T) {
	r1 := speeddaemon.Record{
		Mile: 8,
		Time: 0,
	}
	r2 := speeddaemon.Record{
		Mile: 9,
		Time: 45,
	}
	speed := speeddaemon.CalculateSpeed(r1, r2)
	assert.Equal(t, speed, speeddaemon.U16(8000), "Speed should be 80")
	// Now check the reverse
	speed = speeddaemon.CalculateSpeed(r2, r1)
	assert.Equal(t, speed, speeddaemon.U16(8000), "Speed should be 80")
}

func TestCentralDispatcher(t *testing.T) {
	dispatcher := speeddaemon.NewCentralDispatcher()
	go dispatcher.Start()
	dispatcherChannel := make(chan speeddaemon.ClientMessage, 5)
	dispatcher.MessageQueue <- &speeddaemon.RegisterDispatcher{
		Roads:   []speeddaemon.U16{1},
		Channel: dispatcherChannel,
	}
	dispatcher.MessageQueue <- &speeddaemon.RegisterCamera{
		Road:  1,
		Limit: 60,
	}
	dispatcher.MessageQueue <- &speeddaemon.Observation{
		Road:      1,
		License:   "ABC123",
		Mile:      8,
		Timestamp: 0,
	}
	dispatcher.MessageQueue <- &speeddaemon.Observation{
		Road:      1,
		License:   "ABC123",
		Mile:      9,
		Timestamp: 45,
	}
	// Recieves speeding ticket
	ticket := <-dispatcherChannel
	assert.Equal(t, ticket.GetType(), speeddaemon.TicketMsgType, "Ticket message type should match")
	// Doesn't recieve speeding ticket on same day
	dispatcher.MessageQueue <- &speeddaemon.Observation{
		Road:      1,
		License:   "ABC123",
		Mile:      10,
		Timestamp: 100,
	}
	select {
	case <-dispatcherChannel:
		t.Error(t, "Should not receive a ticket message")
	default:
		// No ticket message received, as expected
	}
}

func TestSession(t *testing.T) {
	// Start the server
	port := 5006
	go speeddaemon.Listen(port)
	time.Sleep(100 * time.Millisecond)
	dispatcher, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
	assert.NoError(t, err, "Failed to connect to server")
	dMessage := &speeddaemon.IAmDispatcherMessage{
		Numroads: 2,
		Roads:    []speeddaemon.U16{1, 2},
	}
	encodedMessage, err := dMessage.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	_, err = dispatcher.Write(encodedMessage)
	assert.NoError(t, err, "Failed to write to server")
	// Register camera 1
	camera1, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
	assert.NoError(t, err, "Failed to connect to server")
	cMessage1 := &speeddaemon.IAmCameraMessage{
		Road:  1,
		Mile:  8,
		Limit: 60,
	}
	encodedMessage, err = cMessage1.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	_, err = camera1.Write(encodedMessage)
	assert.NoError(t, err, "Failed to write to server")
	// Register camera 2
	camera2, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
	assert.NoError(t, err, "Failed to connect to server")
	cMessage2 := &speeddaemon.IAmCameraMessage{
		Road:  1,
		Mile:  9,
		Limit: 60,
	}
	encodedMessage, err = cMessage2.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	_, err = camera2.Write(encodedMessage)
	assert.NoError(t, err, "Failed to write to server")
	// Send the first plate message
	plate1 := &speeddaemon.PlateMessage{
		Plate:     "ABC123",
		Timestamp: 0,
	}
	encodedMessage, err = plate1.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	_, err = camera1.Write(encodedMessage)
	assert.NoError(t, err, "Failed to write to server")
	// Send the second plate message
	plate2 := &speeddaemon.PlateMessage{
		Plate:     "ABC123",
		Timestamp: 45,
	}
	encodedMessage, err = plate2.Encode()
	assert.NoError(t, err, "Encoding should not produce an error")
	_, err = camera2.Write(encodedMessage)
	assert.NoError(t, err, "Failed to write to server")
	// Read out the ticket message
	buffer := make([]byte, 1024)
	n, err := dispatcher.Read(buffer)
	assert.NoError(t, err, "Failed to read from server")
	ticketMessage, err := speeddaemon.Decode(speeddaemon.NewSdBuffer(buffer[:n]))
	assert.NoError(t, err, "Failed to decode message")
	assert.Equal(t, ticketMessage.GetType(), speeddaemon.TicketMsgType, "Ticket message type should match")
	ticket := ticketMessage.(*speeddaemon.TicketMessage)
	assert.Equal(t, ticket.Plate, speeddaemon.Str("ABC123"), "Ticket plate should match")
	assert.Equal(t, ticket.Road, speeddaemon.U16(1), "Ticket road should match")
	assert.Equal(t, ticket.MileOne, speeddaemon.U16(8), "Ticket mile one should match")
	assert.Equal(t, ticket.TimeStampOne, speeddaemon.U32(0), "Ticket timestamp one should match")
	assert.Equal(t, ticket.MileTwo, speeddaemon.U16(9), "Ticket mile two should match")
	assert.Equal(t, ticket.TimeStampTwo, speeddaemon.U32(45), "Ticket timestamp two should match")
	assert.Equal(t, ticket.Speed, speeddaemon.U16(8000), "Ticket speed should match")
	// Clean up
	dispatcher.Close()
	camera1.Close()
	camera2.Close()
}
