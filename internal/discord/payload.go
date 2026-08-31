// Package discord provides low-level Discord Rich Presence IPC protocol implementation,
// including binary framing, handshakes, and activity state dispatch.
package discord

// Opcode represents the 32-bit little-endian operation code in Discord IPC frames.
type Opcode uint32

const (
	// OpcodeHandshake (0) is sent by the client to initialize the IPC session.
	OpcodeHandshake Opcode = 0

	// OpcodeFrame (1) is sent for data payloads (e.g. SET_ACTIVITY commands).
	OpcodeFrame Opcode = 1

	// OpcodeClose (2) is sent when terminating the session.
	OpcodeClose Opcode = 2

	// OpcodePing (3) is a heartbeat ping frame.
	OpcodePing Opcode = 3

	// OpcodePong (4) is a heartbeat pong response frame.
	OpcodePong Opcode = 4
)

// HandshakePayload represents the initial handshake JSON structure sent to Discord.
type HandshakePayload struct {
	V        int    `json:"v"`
	ClientID string `json:"client_id"`
}

// ActivityTimestamps contains Unix timestamps (in seconds or milliseconds) for start/end time.
type ActivityTimestamps struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

// ActivityAssets contains asset image keys and hover tooltips configured in Discord Developer Portal.
type ActivityAssets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

// ActivityButton represents an interactive button displayed on the user's Discord profile.
type ActivityButton struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Activity represents the complete Discord Rich Presence activity object.
type Activity struct {
	Details    string              `json:"details,omitempty"`
	State      string              `json:"state,omitempty"`
	Timestamps *ActivityTimestamps `json:"timestamps,omitempty"`
	Assets     *ActivityAssets     `json:"assets,omitempty"`
	Buttons    []*ActivityButton   `json:"buttons,omitempty"`
	Instance   bool                `json:"instance"`
}

// SetActivityArgs wraps the arguments for the SET_ACTIVITY command.
type SetActivityArgs struct {
	Pid      int       `json:"pid"`
	Activity *Activity `json:"activity,omitempty"`
}

// SetActivityPayload represents the full JSON envelope sent with Opcode 1 to update presence.
type SetActivityPayload struct {
	Cmd   string          `json:"cmd"`
	Args  SetActivityArgs `json:"args"`
	Nonce string          `json:"nonce"`
}

// IPCResponse represents a generic response payload received from Discord.
type IPCResponse struct {
	Cmd   string                 `json:"cmd"`
	Data  map[string]interface{} `json:"data"`
	Evt   string                 `json:"evt"`
	Nonce string                 `json:"nonce"`
}
