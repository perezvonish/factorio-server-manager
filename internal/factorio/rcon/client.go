package rcon

import (
	"fmt"

	gorcon "github.com/gorcon/rcon"
)

// PasswordProvider returns the current RCON password
type PasswordProvider interface {
	Get() string
}

// Client implements domain.RconExecutor using the gorcon library
type Client struct {
	host      string
	port      string
	passwords PasswordProvider
}

func NewClient(host, port string, passwords PasswordProvider) *Client {
	return &Client{
		host:      host,
		port:      port,
		passwords: passwords,
	}
}

// Execute dials the Factorio RCON server, runs the command, and returns the response
func (c *Client) Execute(command string) (string, error) {
	addr := fmt.Sprintf("%s:%s", c.host, c.port)

	conn, err := gorcon.Dial(addr, c.passwords.Get())
	if err != nil {
		return "", fmt.Errorf("RCON connect: %w", err)
	}
	defer conn.Close()

	resp, err := conn.Execute(command)
	if err != nil {
		return "", fmt.Errorf("RCON exec: %w", err)
	}

	return resp, nil
}
