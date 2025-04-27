package insecuresocketlayer_test

import (
	"testing"

	"github.com/JeremyFenwick/firewatch/internal/insecuresocketslayer"
	"github.com/stretchr/testify/assert"
)

func SetupCipher(t *testing.T, cipherData []byte) *insecuresocketslayer.Cipher {
	cipher, usedBytes, err := insecuresocketslayer.NewCipher(cipherData)
	assert.Equal(t, len(cipherData), usedBytes)
	assert.NoError(t, err)
	assert.Equal(t, true, cipher.Valid)
	return cipher
}

func TestDecode(t *testing.T) {
	cipherData := []byte{0x02, 0x01, 0x01, 0x00}
	cipher := SetupCipher(t, cipherData)
	data := []byte("hello")
	encoded := cipher.EncodeData(data)
	assert.Equal(t, encoded, []byte{0x96, 0x26, 0xb6, 0xb6, 0x76})
	decoded := cipher.DecodeData(encoded)
	assert.Equal(t, data, decoded)
}

func TestOtherDecode(t *testing.T) {
	cipherData := []byte{0x05, 0x05, 0x00}
	cipher := SetupCipher(t, cipherData)
	data := []byte("hello")
	encoded := cipher.EncodeData(data)
	assert.Equal(t, encoded, []byte{0x68, 0x67, 0x70, 0x72, 0x77})
	decoded := cipher.DecodeData(encoded)
	assert.Equal(t, data, decoded)
}

func TestBasicEncodeDecode(t *testing.T) {
	// Setup the cipher
	codecData := []byte{0x01, 0x03, 0x00}
	cipher := SetupCipher(t, codecData)
	data := []byte("Hello, World!\n")
	encoded := cipher.EncodeData(data)
	decoded := cipher.DecodeData(encoded)
	assert.Equal(t, data, decoded)
}
