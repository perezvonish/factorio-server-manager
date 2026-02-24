package domain

// RconExecutor executes commands on the Factorio RCON interface.
type RconExecutor interface {
	Execute(command string) (string, error)
}
