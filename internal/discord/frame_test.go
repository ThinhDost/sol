package discord

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeFrame(t *testing.T) {
	testPayload := []byte(`{"v":1,"client_id":"123456"}`)
	encoded := EncodeFrame(OpcodeHandshake, testPayload)

	expectedLen := HeaderSize + len(testPayload)
	if len(encoded) != expectedLen {
		t.Fatalf("expected encoded length %d, got %d", expectedLen, len(encoded))
	}

	buf := bytes.NewReader(encoded)
	op, decodedPayload, err := DecodeFrame(buf)
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}

	if op != OpcodeHandshake {
		t.Errorf("expected Opcode %d, got %d", OpcodeHandshake, op)
	}

	if !bytes.Equal(decodedPayload, testPayload) {
		t.Errorf("expected payload %s, got %s", string(testPayload), string(decodedPayload))
	}
}

func TestDecodeEmptyPayload(t *testing.T) {
	encoded := EncodeFrame(OpcodePing, []byte{})
	buf := bytes.NewReader(encoded)

	op, payload, err := DecodeFrame(buf)
	if err != nil {
		t.Fatalf("DecodeFrame with empty payload failed: %v", err)
	}

	if op != OpcodePing {
		t.Errorf("expected OpcodePing, got %d", op)
	}

	if len(payload) != 0 {
		t.Errorf("expected empty payload, got %d bytes", len(payload))
	}
}
