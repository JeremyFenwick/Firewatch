package speeddaemon

import (
	"encoding/binary"
	"errors"
	"io"
)

type (
	U8  uint8
	U16 uint16
	U32 uint32
	Str string
)

var (
	ErrIncompleteMessage = errors.New("incomplete message")
	ErrInvalidMessage    = errors.New("invalid message")
)

func readU8(r *SdBuffer) (U8, error) {
	var buffer [1]byte
	n, err := io.ReadFull(r.Reader, buffer[:])
	if err != nil {
		return 0, err
	}
	r.ValidBytes += n
	return U8(buffer[0]), nil
}

func readU16(r *SdBuffer) (U16, error) {
	var buffer [2]byte
	n, err := io.ReadFull(r.Reader, buffer[:])
	if err != nil {
		return 0, err
	}
	r.ValidBytes += n
	return U16(binary.BigEndian.Uint16(buffer[:])), nil
}

func readU32(r *SdBuffer) (U32, error) {
	var buffer [4]byte
	n, err := io.ReadFull(r.Reader, buffer[:])
	if err != nil {
		return 0, err
	}
	r.ValidBytes += n
	return U32(binary.BigEndian.Uint32(buffer[:])), nil
}

func readStr(r *SdBuffer) (Str, error) {
	length, err := readU8(r)
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	buffer := make([]byte, length)
	n, err := io.ReadFull(r.Reader, buffer)
	if err != nil {
		return "", err
	}
	r.ValidBytes += n
	return Str(buffer), nil
}

func writeU8(w io.Writer, value U8) error {
	_, err := w.Write([]byte{byte(value)})
	return err
}

func writeU16(w io.Writer, value U16) error {
	var buffer [2]byte
	binary.BigEndian.PutUint16(buffer[:], uint16(value))
	_, err := w.Write(buffer[:])
	return err
}

func writeU32(w io.Writer, value U32) error {
	var buffer [4]byte
	binary.BigEndian.PutUint32(buffer[:], uint32(value))
	_, err := w.Write(buffer[:])
	return err
}

func writeStr(w io.Writer, value Str) error {
	length := len(value)
	if length > 255 {
		return errors.New("string too long")
	}
	err := writeU8(w, U8(length))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(value))
	return err
}
