package domain

import "context"

// ContainerManager controls the lifecycle of the Factorio Docker container
type ContainerManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
