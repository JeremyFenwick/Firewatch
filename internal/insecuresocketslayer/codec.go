package insecuresocketslayer

import "fmt"

const DummyMessage = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

type Cipher struct {
	Operations []*Instruction
	Valid      bool
}

// USER INTERFACE

// Returns the cipher, the number of bytes used, and an error if any
func NewCipher(data []byte) (*Cipher, int, error) {
	// Generate the cipher
	cipher, bytesUsed, err := getCipher(data)
	if err != nil {
		return nil, 0, err
	}
	// Check cipher validity
	cipher.Valid = cipher.validCipher()
	return cipher, bytesUsed, nil
}

func (c *Cipher) EncodeData(data []byte) []byte {
	encoded := make([]byte, len(data))
	for i, element := range data {
		for _, inst := range c.Operations {
			element = inst.encode(i, element)
		}
		encoded[i] = element
	}
	return encoded
}

func (c *Cipher) DecodeData(data []byte) []byte {
	decoded := make([]byte, len(data))
	for i, element := range data {
		// Walk through the operations in reverse order
		for j := len(c.Operations) - 1; j >= 0; j-- {
			inst := c.Operations[j]
			element = inst.decode(i, element)
		}
		decoded[i] = element
	}
	return decoded
}

// Validate the cipher. It cannot reproduce the same data after encoding
func (c *Cipher) validCipher() bool {
	testData := []byte(DummyMessage)
	encoded := c.EncodeData(testData)
	for i, element := range encoded {
		if element != testData[i] {
			return true
		}
	}
	return false
}

// CIPHER EXTRACTOR

func getCipher(data []byte) (*Cipher, int, error) {
	var (
		index      int
		operations []*Instruction
	)

	for index < len(data) {
		switch b := data[index]; b {
		case 0x00: // End of cipher spec
			index++
			return &Cipher{Operations: operations}, index, nil

		case 0x01: // Reverse bits
			operations = append(operations, &Instruction{Operation: "reversebits"})
			index++

		case 0x02: // Xor N
			if index+1 >= len(data) {
				return nil, 0, fmt.Errorf("unexpected end of data after xor instruction")
			}
			operations = append(operations, &Instruction{
				Operation: "xor",
				Operand:   int(data[index+1]),
			})
			index += 2

		case 0x03: // Xor pos
			operations = append(operations, &Instruction{Operation: "xorpos"})
			index++

		case 0x04: // Add N
			if index+1 >= len(data) {
				return nil, 0, fmt.Errorf("unexpected end of data after add instruction")
			}
			operations = append(operations, &Instruction{
				Operation: "add",
				Operand:   int(data[index+1]),
			})
			index += 2

		case 0x05: // Add pos
			operations = append(operations, &Instruction{Operation: "addpos"})
			index++

		default:
			return nil, 0, fmt.Errorf("encountered unknown cipher operation: %d", b)
		}
	}

	return nil, 0, fmt.Errorf("could not find end of cipher")
}

// ENCODER AND DECORED FOR INSTRUCTIONS

// Reversebits
// Xor (n)
// Xorpos
// Add (n)
// Addpos

type Instruction struct {
	Operation string
	Operand   int
}

func (inst *Instruction) encode(position int, element byte) byte {
	switch inst.Operation {
	case "reversebits":
		return reverseBits(element)
	case "xor":
		return element ^ byte(inst.Operand)
	case "xorpos":
		return element ^ byte(position)
	case "add":
		return element + byte(inst.Operand) // Overflow wraps in go as per the spec
	case "addpos":
		return element + byte(position) // Overflow wraps in go as per the spec
	default:
		panic(fmt.Sprintf("unknown operation: %s", inst.Operation))
	}
}

func (inst *Instruction) decode(position int, element byte) byte {
	switch inst.Operation {
	// Reverse bits is its own inverse
	case "reversebits":
		return reverseBits(element)
	// Xor is its own inverse
	case "xor":
		return element ^ byte(inst.Operand)
	// Xorpos is its own inverse
	case "xorpos":
		return element ^ byte(position)
	// Minus to reverse add
	case "add":
		return element - byte(inst.Operand) // Overflow wraps in go as per the spec
	// Minus to reverse addpos
	case "addpos":
		return element - byte(position) // Overflow wraps in go as per the spec
	default:
		panic(fmt.Sprintf("unknown operation: %s", inst.Operation))
	}
}

func reverseBits(element byte) byte {
	var result byte
	for range 8 {
		result <<= 1
		result |= element & 1
		element >>= 1
	}
	return result
}
