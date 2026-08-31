package discord

import (
	"encoding/binary"
	"fmt"
	"io"
)

// HeaderSize is the fixed 8-byte size of a Discord IPC packet header (4 bytes Opcode + 4 bytes Length).
const HeaderSize = 8

// MaxPayloadSize protects against allocating excessive memory on malformed frames (1MB max).
const MaxPayloadSize = 1024 * 1024

// EncodeFrame serializes an opcode and payload bytes into the Discord binary packet format:
// [4 bytes Opcode LE] + [4 bytes Length LE] + [Payload bytes]
func EncodeFrame(op Opcode, payload []byte) []byte {
	length := uint32(len(payload))
	buf := make([]byte, HeaderSize+len(payload))

	// Write 32-bit Little Endian Opcode
	binary.LittleEndian.PutUint32(buf[0:4], uint32(op))

	// Write 32-bit Little Endian Payload Length
	binary.LittleEndian.PutUint32(buf[4:8], length)

	// Copy JSON payload bytes
	copy(buf[8:], payload)

	return buf
}

// DecodeFrame reads an 8-byte header from the reader, parses the Opcode and Length,
// and then reads the exact payload bytes into memory.
func DecodeFrame(r io.Reader) (Opcode, []byte, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, fmt.Errorf("failed to read frame header: %w", err)
	}

	op := Opcode(binary.LittleEndian.Uint32(header[0:4]))
	length := binary.LittleEndian.Uint32(header[4:8])

	if length > MaxPayloadSize {
		return 0, nil, fmt.Errorf("payload size %d exceeds maximum limit %d", length, MaxPayloadSize)
	}

	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return 0, nil, fmt.Errorf("failed to read frame payload: %w", err)
		}
	}

	return op, payload, nil
}
