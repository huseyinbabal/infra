package executor

import "context"

type Executor interface {
	Name() string
	Run(ctx context.Context) error
}
