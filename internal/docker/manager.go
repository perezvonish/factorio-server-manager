package docker

import (
	"context"
	"fmt"
	"os/exec"
)

// Manager implements domain.ContainerManager using the Docker CLI
// Requires that the bot process has access to the Docker socket.
type Manager struct {
	containerName string
}

func NewManager(containerName string) *Manager {
	return &Manager{containerName: containerName}
}

// Start starts the Docker container running the Factorio server
func (m *Manager) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "start", m.containerName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker start %s: %w\n%s", m.containerName, err, string(out))
	}
	return nil
}

// Stop stops the Docker container running the Factorio server
func (m *Manager) Stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "stop", m.containerName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker stop %s: %w\n%s", m.containerName, err, string(out))
	}
	return nil
}
