package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Job struct {
	Name    string
	Fn      func(ctx context.Context) error
}

type Pool struct {
	workers    int
	jobs       chan Job
	logger     zerolog.Logger
	wg         sync.WaitGroup
}

func NewPool(workers int, queueSize int, logger zerolog.Logger) *Pool {
	return &Pool{
		workers: workers,
		jobs:    make(chan Job, queueSize),
		logger:  logger,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func(id int) {
			defer p.wg.Done()
			p.logger.Debug().Int("worker_id", id).Msg("Worker started")

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-p.jobs:
					if !ok {
						return
					}
					start := time.Now()
					if err := job.Fn(ctx); err != nil {
						p.logger.Error().Err(err).
							Str("job", job.Name).
							Dur("duration", time.Since(start)).
							Msg("Job failed")
					} else {
						p.logger.Debug().
							Str("job", job.Name).
							Dur("duration", time.Since(start)).
							Msg("Job completed")
					}
				}
			}
		}(i)
	}
}

func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
