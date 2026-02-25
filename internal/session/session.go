package session

import (
	"fmt"
	"time"

	"github.com/nmashchenko/aegis-cli/internal/db"
	"github.com/nmashchenko/aegis-cli/internal/models"
)

type Service struct {
	db *db.DB
}

type StartResult struct {
	TaskID   int64
	TaskName string
}

type StopResult struct {
	TaskName string
	Duration time.Duration
}

type StatusResult struct {
	Active   bool
	Paused   bool
	TaskName string
	Elapsed  time.Duration
	TaskID   int64
	Limit    *time.Duration
}

type UrgeResult struct {
	TaskName string
}

type PauseResult struct {
	TaskName string
}

type ResumeResult struct {
	TaskName string
}

func NewService(database *db.DB) *Service {
	return &Service{db: database}
}

func (s *Service) Start(name string, limitSeconds *int64) (*StartResult, error) {
	active, err := s.db.GetActiveTask()
	if err != nil {
		return nil, fmt.Errorf("check active task: %w", err)
	}
	if active != nil {
		return nil, fmt.Errorf("task %q is already running. Stop it first with \"aegis stop\"", active.Name)
	}

	id, err := s.db.CreateTask(name, limitSeconds)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return &StartResult{TaskID: id, TaskName: name}, nil
}

func (s *Service) Stop() (*StopResult, error) {
	active, err := s.db.GetActiveTask()
	if err != nil {
		return nil, fmt.Errorf("check active task: %w", err)
	}
	if active == nil {
		return nil, fmt.Errorf("no active task to stop")
	}

	if err := s.db.StopTask(active.ID); err != nil {
		return nil, fmt.Errorf("stop task: %w", err)
	}

	// Compute effective duration excluding paused time
	elapsed := time.Since(active.StartedAt)
	pausedDuration := time.Duration(active.TotalPausedSeconds) * time.Second
	if active.PausedAt != nil {
		pausedDuration += time.Since(*active.PausedAt)
	}
	duration := elapsed - pausedDuration

	return &StopResult{
		TaskName: active.Name,
		Duration: duration,
	}, nil
}

func (s *Service) Pause() (*PauseResult, error) {
	active, err := s.db.GetActiveTask()
	if err != nil {
		return nil, fmt.Errorf("check active task: %w", err)
	}
	if active == nil {
		return nil, fmt.Errorf("no active task to pause")
	}
	if active.PausedAt != nil {
		return nil, fmt.Errorf("task %q is already paused", active.Name)
	}

	if err := s.db.PauseTask(active.ID); err != nil {
		return nil, fmt.Errorf("pause task: %w", err)
	}

	return &PauseResult{TaskName: active.Name}, nil
}

func (s *Service) Resume() (*ResumeResult, error) {
	active, err := s.db.GetActiveTask()
	if err != nil {
		return nil, fmt.Errorf("check active task: %w", err)
	}
	if active == nil {
		return nil, fmt.Errorf("no active task to resume")
	}
	if active.PausedAt == nil {
		return nil, fmt.Errorf("task %q is not paused", active.Name)
	}

	if err := s.db.ResumeTask(active.ID); err != nil {
		return nil, fmt.Errorf("resume task: %w", err)
	}

	return &ResumeResult{TaskName: active.Name}, nil
}

func (s *Service) Status() (*StatusResult, error) {
	active, err := s.db.GetActiveTask()
	if err != nil {
		return nil, fmt.Errorf("check active task: %w", err)
	}

	if active == nil {
		return &StatusResult{Active: false}, nil
	}

	elapsed := time.Since(active.StartedAt)
	pausedDuration := time.Duration(active.TotalPausedSeconds) * time.Second
	if active.PausedAt != nil {
		pausedDuration += time.Since(*active.PausedAt)
	}
	effectiveElapsed := elapsed - pausedDuration

	result := &StatusResult{
		Active:   true,
		Paused:   active.PausedAt != nil,
		TaskName: active.Name,
		Elapsed:  effectiveElapsed,
		TaskID:   active.ID,
	}

	if active.LimitSeconds != nil {
		limit := time.Duration(*active.LimitSeconds) * time.Second
		result.Limit = &limit
	}

	return result, nil
}

func (s *Service) LogUrge() (*UrgeResult, error) {
	active, err := s.db.GetActiveTask()
	if err != nil {
		return nil, fmt.Errorf("check active task: %w", err)
	}

	var taskID *int64
	var taskName string
	if active != nil {
		taskID = &active.ID
		taskName = active.Name
	}

	if err := s.db.CreateUrge(taskID); err != nil {
		return nil, fmt.Errorf("create urge: %w", err)
	}

	return &UrgeResult{TaskName: taskName}, nil
}

func (s *Service) GetUrgeCount(taskID int64) (int, error) {
	return s.db.GetUrgeCountForTask(taskID)
}

func (s *Service) GetRandomHighlight(moodTier string) (*models.PaperHighlight, error) {
	return s.db.GetRandomHighlight(moodTier)
}
