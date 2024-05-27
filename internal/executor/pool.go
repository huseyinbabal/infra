package executor

import (
	"context"
	"github.com/kyokomi/emoji/v2"
	"sync"
)

type Pool struct {
	Executors []Executor
	Results   map[string]error
	l         sync.Mutex
}

func NewPool(executors []Executor) *Pool {
	return &Pool{Executors: executors, Results: make(map[string]error)}
}

func (p *Pool) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, executor := range p.Executors {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			emoji.Println(":hourglass_flowing_sand:" + executor.Name() + " started")
			err := executor.Run(ctx)
			if err != nil {
				emoji.Println(":x:" + executor.Name() + " failed\n\t:ladybug:" + err.Error() + "\n")
			} else {
				emoji.Println(":white_check_mark:" + executor.Name() + " succeeded")

			}
			p.l.Lock()
			p.Results[executor.Name()] = err
			p.l.Unlock()
		}(ctx)
	}
	wg.Wait()
}
