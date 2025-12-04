package ports

import (
	"context"
)

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}
