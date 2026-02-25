package models

import "time"

type Task struct {
	ID                 int64
	Name               string
	StartedAt          time.Time
	EndedAt            *time.Time
	DurationSeconds    *int64
	LimitSeconds       *int64
	PausedAt           *time.Time
	TotalPausedSeconds int64
}

type Urge struct {
	ID        int64
	Timestamp time.Time
	TaskID    *int64
}

type Stats struct {
	Period         string
	TasksCompleted int
	TotalFocusSecs int64
	AvgTaskSecs    int64
	UrgesLogged    int
}

type TaskHistory struct {
	ID              int64
	Name            string
	StartedAt       time.Time
	DurationSeconds int64
	UrgeCount       int
}

type DailyUrgeCount struct {
	Date  time.Time
	Count int
}

type StatBucket struct {
	Label string
	Tasks int
	Urges int
}

type PaperHighlight struct {
	Title     string
	URL       string
	Highlight string
}
