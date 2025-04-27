package insecuresocketslayer

import (
	"bytes" // For validation check comparison
	"fmt"
	"runtime"
	"sync"
)

const DummyMessage = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

// --- Operation Codes ---
const (
	opEnd         byte = 0x00
	opReverseBits byte = 0x01
	opXor         byte = 0x02
	opXorPos      byte = 0x03
	opAdd         byte = 0x04
	opAddPos      byte = 0x05
	// We can add inverse ops codes if needed, but applying the inverse logic
	// directly might be clearer. Let's stick to the original codes for now.
)

// Instruction now uses OpCode instead of string
type Instruction struct {
	OpCode  byte
	Operand byte // Operand is always a byte
}

type Cipher struct {
	encodeOps     []*Instruction // Operations for encoding (original order)
	decodeOps     []*Instruction // Precomputed inverse operations for decoding (reversed order)
	rawCipherSpec []byte         // Store the original spec for validation check
	Valid         bool
}

// --- Bit Reverse Table (no changes needed) ---
var bitReverseTable [256]byte

func init() {
	fmt.Println("Precomputing bit reverse table...")
	for i := range 256 {
		var result byte
		element := byte(i)
		for range 8 {
			result <<= 1
			result |= element & 1
			element >>= 1
		}
		bitReverseTable[i] = result
	}
	fmt.Println("Bit reverse table computed.")
}

// --- User Interface ---

// NewCipher now precomputes decode operations
func NewCipher(data []byte) (*Cipher, error) {
	encodeOps, specLen, err := parseCipherSpec(data)
	if err != nil {
		return nil, err
	}

	// Store the raw spec slice (only the relevant part)
	rawSpec := make([]byte, specLen)
	copy(rawSpec, data[:specLen])

	cipher := &Cipher{
		encodeOps:     encodeOps,
		rawCipherSpec: rawSpec,
		// decodeOps will be populated below
	}

	// Precompute the inverse operations in reverse order
	cipher.decodeOps = make([]*Instruction, len(encodeOps))
	for i, op := range encodeOps {
		// Target index in decodeOps is reversed
		decodeIndex := len(encodeOps) - 1 - i

		// Create the inverse instruction
		invInst := &Instruction{
			OpCode:  op.OpCode,  // OpCode stays the same, logic handles inversion
			Operand: op.Operand, // Operand stays the same
		}
		// Note: The *logic* in applyDecode will handle the inversion (e.g., add becomes subtract)
		cipher.decodeOps[decodeIndex] = invInst
	}

	// Check cipher validity using the original encodeOps
	cipher.Valid = cipher.isValidCipher()

	return cipher, nil
}

// EncodeData uses encodeOps and iterates forward
func (c *Cipher) EncodeData(startPos int, data []byte) []byte {
	encoded := make([]byte, len(data))
	// Optimization: Copy data to avoid modifying the original slice if it's needed elsewhere
	copy(encoded, data)

	for i := range encoded {
		element := encoded[i]
		currentPos := startPos + i
		for _, inst := range c.encodeOps { // Iterate forward through encodeOps
			element = applyEncode(inst, currentPos, element)
		}
		encoded[i] = element
	}
	return encoded
}

// --- Decoding Logic ---

// Threshold for switching to parallel execution (tune based on benchmarking)
const parallelDecodeThreshold = 2048 // e.g., 2KB

// DecodeData acts as a dispatcher
func (c *Cipher) DecodeData(startPos int, data []byte) []byte {
	if len(data) < parallelDecodeThreshold || runtime.NumCPU() <= 1 {
		// Use sequential version for small data or single-core machines
		return c.decodeSequential(startPos, data)
	} else {
		// Use parallel version for larger data on multi-core machines
		return c.decodeParallel(startPos, data)
	}
}

// decodeSequential - The optimized sequential implementation
func (c *Cipher) decodeSequential(startPos int, data []byte) []byte {
	decoded := make([]byte, len(data))
	copy(decoded, data) // Work on a copy

	for i := range decoded {
		element := decoded[i]
		currentPos := startPos + i
		for _, inst := range c.decodeOps { // Use precomputed inverse ops
			element = applyDecode(inst, currentPos, element)
		}
		decoded[i] = element
	}
	return decoded
}

// decodeParallel - The parallel implementation
func (c *Cipher) decodeParallel(startPos int, data []byte) []byte {
	n := len(data)
	decoded := make([]byte, n)
	copy(decoded, data) // Start with encoded data, decode in place

	numWorkers := runtime.NumCPU()
	// Calculate chunk size, ensuring it distributes work somewhat evenly
	chunkSize := (n + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		// Calculate the start and end index for this worker's chunk
		chunkStart := i * chunkSize
		chunkEnd := chunkStart + chunkSize
		if chunkEnd > n {
			chunkEnd = n // Clamp to the end of the data
		}

		// Skip if the chunk is empty (can happen if n is small or chunkSize is large)
		if chunkStart >= chunkEnd {
			continue
		}

		wg.Add(1)
		// Launch goroutine for the chunk
		go func(startOffset, endOffset int) {
			defer wg.Done()
			// Process each byte within the assigned chunk [startOffset, endOffset)
			for j := startOffset; j < endOffset; j++ {
				element := decoded[j]              // Read the byte to be decoded
				currentPos := startPos + j         // Calculate the absolute stream position
				for _, inst := range c.decodeOps { // Apply precomputed inverse operations
					element = applyDecode(inst, currentPos, element)
				}
				decoded[j] = element // Write the decoded byte back to the slice
			}
		}(chunkStart, chunkEnd) // Pass chunk boundaries to the goroutine
	}

	wg.Wait() // Wait for all goroutines to finish
	return decoded
}

// --- Cipher Validation ---

// isValidCipher checks if the cipher is a no-op
func (c *Cipher) isValidCipher() bool {
	// Quick check for empty cipher spec
	if len(c.rawCipherSpec) == 1 && c.rawCipherSpec[0] == opEnd {
		return false
	}

	// Test with a known pattern
	testData := []byte(DummyMessage)
	encoded := c.EncodeData(0, testData) // Use EncodeData for the test

	// Compare original and encoded data
	return !bytes.Equal(testData, encoded)
}

// --- Cipher Spec Parsing ---

func parseCipherSpec(data []byte) ([]*Instruction, int, error) {
	var (
		index      int
		operations []*Instruction
	)

	for index < len(data) {
		opCode := data[index]
		inst := &Instruction{OpCode: opCode}

		switch opCode {
		case opEnd: // End of cipher spec
			index++
			return operations, index, nil // Return total length consumed

		case opReverseBits, opXorPos, opAddPos: // Single-byte instructions
			operations = append(operations, inst)
			index++

		case opXor, opAdd: // Two-byte instructions
			if index+1 >= len(data) {
				return nil, 0, fmt.Errorf("unexpected end of data after op code %02x", opCode)
			}
			inst.Operand = data[index+1]
			operations = append(operations, inst)
			index += 2

		default:
			return nil, 0, fmt.Errorf("encountered unknown cipher operation: %02x", opCode)
		}
	}

	return nil, 0, fmt.Errorf("could not find end of cipher spec (0x00)")
}

// --- Core Encoding/Decoding Logic (using OpCodes) ---

// applyEncode applies a single encoding operation
func applyEncode(inst *Instruction, position int, element byte) byte {
	switch inst.OpCode {
	case opReverseBits:
		return bitReverseTable[element]
	case opXor:
		return element ^ inst.Operand
	case opXorPos:
		return element ^ byte(position) // Modulo 256 handled by byte cast
	case opAdd:
		return element + inst.Operand // Overflow wraps correctly
	case opAddPos:
		return element + byte(position) // Overflow wraps correctly
	default:
		// Should not happen if parseCipherSpec is correct
		panic(fmt.Sprintf("unknown operation code during encode: %02x", inst.OpCode))
	}
}

// applyDecode applies a single inverse operation
// Note: It receives the *original* instruction but applies the *inverse* logic.
func applyDecode(inst *Instruction, position int, element byte) byte {
	switch inst.OpCode {
	// Self-inverting operations
	case opReverseBits:
		return bitReverseTable[element] // Inverse is the same
	case opXor:
		return element ^ inst.Operand // Inverse is the same
	case opXorPos:
		return element ^ byte(position) // Inverse is the same

	// Invert Add operations with Subtract
	case opAdd:
		return element - inst.Operand // Subtraction is inverse of addition (wraps correctly)
	case opAddPos:
		return element - byte(position) // Subtraction is inverse of addition (wraps correctly)

	default:
		// Should not happen if parseCipherSpec/precomputation is correct
		panic(fmt.Sprintf("unknown operation code during decode: %02x", inst.OpCode))
	}
}
