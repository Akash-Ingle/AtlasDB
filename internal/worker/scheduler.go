package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

type Scheduler struct {
	tasks  []ScheduledTask
	logger zerolog.Logger
}

func NewScheduler(logger zerolog.Logger) *Scheduler {
	return &Scheduler{logger: logger}
}

func (s *Scheduler) Register(task ScheduledTask) {
	s.tasks = append(s.tasks, task)
}

func (s *Scheduler) Run(ctx context.Context) {
	for _, task := range s.tasks {
		go s.runTask(ctx, task)
	}

	s.logger.Info().Int("tasks", len(s.tasks)).Msg("Scheduler started")
	<-ctx.Done()
	s.logger.Info().Msg("Scheduler stopped")
}

func (s *Scheduler) runTask(ctx context.Context, task ScheduledTask) {
	// Run once immediately
	s.executeTask(ctx, task)

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.executeTask(ctx, task)
		}
	}
}

func (s *Scheduler) executeTask(ctx context.Context, task ScheduledTask) {
	start := time.Now()
	if err := task.Fn(ctx); err != nil {
		s.logger.Error().Err(err).
			Str("task", task.Name).
			Dur("duration", time.Since(start)).
			Msg("Scheduled task failed")
	} else {
		s.logger.Debug().
			Str("task", task.Name).
			Dur("duration", time.Since(start)).
			Msg("Scheduled task completed")
	}
}
