package status

import (
	"fmt"
	"os/exec"
)

// Checker probes the Factorio game port to determine if the server is reachable
type Checker struct {
	host string
	port string
}

func NewChecker(host, port string) *Checker {
	return &Checker{host: host, port: port}
}

// Check returns a human-readable status string
func (c *Checker) Check() string {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("nc -zv %s %s 2>&1", c.host, c.port))
	if err := cmd.Run(); err != nil {
		return "🔴 Недоступен"
	}
	return "🟢 Работает"
}
