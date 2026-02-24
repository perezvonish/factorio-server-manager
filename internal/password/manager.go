package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789" +
	"!@#$%^&*()-_=+[]{}<>?"

// Manager generates and stores the RCON password in memory
type Manager struct {
	mu         sync.RWMutex
	password   string
	rconPwFile string
}

func NewManager(rconPwFile string) *Manager {
	return &Manager{rconPwFile: rconPwFile}
}

// Generate creates a new random password, writes it to the rconpw file
func (m *Manager) Generate(length int) error {
	pw, err := generatePassword(length)
	if err != nil {
		return fmt.Errorf("generating password: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(m.rconPwFile), 0755); err != nil {
		return fmt.Errorf("creating rcon pw dir: %w", err)
	}

	if err := os.WriteFile(m.rconPwFile, []byte(pw), 0600); err != nil {
		return fmt.Errorf("writing rcon pw file: %w", err)
	}

	m.mu.Lock()
	m.password = pw
	m.mu.Unlock()

	return nil
}

// Get returns the current in-memory RCON password
func (m *Manager) Get() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.password
}

func generatePassword(length int) (string, error) {
	password := make([]byte, length)

	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		password[i] = charset[randomIndex.Int64()]
	}

	return string(password), nil
}
