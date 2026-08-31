package discord

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sol-rpc/sol/internal/config"
)

// Client manages the persistent IPC connection to the local Discord desktop application.
type Client struct {
	clientID  string
	conn      net.Conn
	mu        sync.Mutex
	connected bool
	pid       int
}

// NewClient initializes a new Discord Rich Presence client.
func NewClient(clientID string) *Client {
	if clientID == "" {
		clientID = config.DefaultDiscordClientID
	}
	return &Client{
		clientID: clientID,
		pid:      os.Getpid(),
	}
}

// Connect establishes the socket connection and performs the initial handshake.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected && c.conn != nil {
		return nil
	}

	conn, err := dialDiscordSocket()
	if err != nil {
		c.connected = false
		return fmt.Errorf("discord socket dial failed: %w", err)
	}
	c.conn = conn

	// Send Handshake
	handshake := HandshakePayload{
		V:        1,
		ClientID: c.clientID,
	}
	data, err := json.Marshal(handshake)
	if err != nil {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		return fmt.Errorf("failed to marshal handshake: %w", err)
	}

	frame := EncodeFrame(OpcodeHandshake, data)
	if _, err := c.conn.Write(frame); err != nil {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		return fmt.Errorf("failed to write handshake frame: %w", err)
	}

	// Read Handshake response
	op, respPayload, err := DecodeFrame(c.conn)
	if err != nil {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		return fmt.Errorf("failed to read handshake response: %w", err)
	}

	if op != OpcodeFrame {
		c.conn.Close()
		c.conn = nil
		c.connected = false
		return fmt.Errorf("unexpected handshake response opcode: %d (%s)", op, string(respPayload))
	}

	c.connected = true
	return nil
}

// SetActivity updates the user's Rich Presence status in Discord.
func (c *Client) SetActivity(activity *Activity) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.conn == nil {
		return fmt.Errorf("discord client is not connected")
	}

	payload := SetActivityPayload{
		Cmd: "SET_ACTIVITY",
		Args: SetActivityArgs{
			Pid:      c.pid,
			Activity: activity,
		},
		Nonce: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SET_ACTIVITY payload: %w", err)
	}

	frame := EncodeFrame(OpcodeFrame, data)
	if _, err := c.conn.Write(frame); err != nil {
		c.connected = false
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("failed to write SET_ACTIVITY frame: %w", err)
	}

	// Read response asynchronously / drain response buffer
	go func(conn net.Conn) {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, _ = DecodeFrame(conn)
	}(c.conn)

	return nil
}

// ClearActivity clears the current Rich Presence status.
func (c *Client) ClearActivity() error {
	return c.SetActivity(nil)
}

// Close gracefully terminates the connection to Discord.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.conn == nil {
		return nil
	}

	// Attempt to send OpcodeClose
	frame := EncodeFrame(OpcodeClose, []byte(`{}`))
	_, _ = c.conn.Write(frame)

	err := c.conn.Close()
	c.conn = nil
	c.connected = false
	return err
}

// IsConnected returns whether the client is currently connected to Discord.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}
