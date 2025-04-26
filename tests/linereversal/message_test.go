package linereversal_test

import (
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/linereversal"
	"github.com/stretchr/testify/assert"
)

func TestCreateLRConnectMessage(t *testing.T) {
	connect := &linereversal.LRMessage{
		Type:    "connect",
		Session: 12345,
	}
	buffer := make([]byte, 1000)
	connectEncode, err := connect.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "/connect/12345/", string(buffer[:connectEncode]))
}

func TestCreateLRAckMessage(t *testing.T) {
	ack := &linereversal.LRMessage{
		Type:    "ack",
		Session: 12345,
		Length:  0,
	}
	buffer := make([]byte, 1000)
	ackEncode, err := ack.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "/ack/12345/0/", string(buffer[:ackEncode]))
}

func TestCreateLRDataMessage(t *testing.T) {
	data := &linereversal.LRMessage{
		Type:     "data",
		Session:  12345,
		Position: 5,
		Data:     []byte("hello"),
	}
	buffer := make([]byte, 1000)
	dataEncode, err := data.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "/data/12345/5/hello/", string(buffer[:dataEncode]))
}

func TestCreateLRCloseMessage(t *testing.T) {
	close := &linereversal.LRMessage{
		Type:    "close",
		Session: 12345,
	}
	buffer := make([]byte, 1000)
	closeEncode, err := close.Encode(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "/close/12345/", string(buffer[:closeEncode]))
}

func TestValidateConnectMessage(t *testing.T) {
	connect := &linereversal.LRMessage{
		Type:     "connect",
		Session:  12345,
		Position: 4,
	}
	assert.False(t, connect.Validate())
}

func TestValidateAckMessage(t *testing.T) {
	ack := &linereversal.LRMessage{
		Type:    "ack",
		Session: -12345,
		Length:  4,
	}
	assert.False(t, ack.Validate())
}

func TestPackDataMessage(t *testing.T) {
	message := &linereversal.LRMessage{
		Type:     "data",
		Session:  12345,
		Position: 0,
	}
	rawData := []byte("hello/")
	added := linereversal.PackDataMessage(message, rawData, 0)
	assert.Equal(t, []byte("hello\\/"), message.Data)
	assert.Equal(t, 6, added)
}

func TestDecodeMessage(t *testing.T) {
	message := "/connect/12345/"
	buffer := []byte(message)
	decodedMessage, err := linereversal.DecodeLRMessage(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "connect", decodedMessage.Type)
	assert.Equal(t, 12345, decodedMessage.Session)
}

func TestSlashMessage(t *testing.T) {
	message := "/data/123/0/foo\\/\\/bar\\/\\/baz/"
	buffer := []byte(message)
	decodedMessage, err := linereversal.DecodeLRMessage(buffer)
	assert.NoError(t, err)
	assert.Equal(t, "data", decodedMessage.Type)
	assert.Equal(t, 123, decodedMessage.Session)
	assert.Equal(t, 0, decodedMessage.Position)
	assert.Equal(t, []byte("foo//bar//baz"), decodedMessage.Data)
}

func TestUnescapeData(t *testing.T) {
	data := []byte("a\\/b")
	unescaped, err := linereversal.UnescapeData(data)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a/b"), unescaped)
}
