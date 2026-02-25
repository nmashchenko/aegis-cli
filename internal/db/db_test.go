package db

import (
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestCreateAndGetActiveTask(t *testing.T) {
	db := setupTestDB(t)

	var limitSecs int64 = 1500
	id, err := db.CreateTask("coding", &limitSecs)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero task ID")
	}

	task, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if task == nil {
		t.Fatal("expected active task, got nil")
	}
	if task.Name != "coding" {
		t.Errorf("task name = %q, want %q", task.Name, "coding")
	}
	if task.EndedAt != nil {
		t.Error("expected EndedAt to be nil for active task")
	}
	if task.LimitSeconds == nil || *task.LimitSeconds != 1500 {
		t.Errorf("task limit = %v, want 1500", task.LimitSeconds)
	}
}

func TestCreateTaskNoLimit(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.CreateTask("reading", nil)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	task, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if task.LimitSeconds != nil {
		t.Errorf("expected nil LimitSeconds, got %v", *task.LimitSeconds)
	}
}

func TestStopTask(t *testing.T) {
	db := setupTestDB(t)

	id, err := db.CreateTask("coding", nil)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = db.StopTask(id)
	if err != nil {
		t.Fatalf("StopTask failed: %v", err)
	}

	task, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if task != nil {
		t.Error("expected no active task after stop")
	}
}

func TestPreventDoubleStart(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.CreateTask("coding", nil)
	if err != nil {
		t.Fatalf("first CreateTask failed: %v", err)
	}

	active, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if active == nil {
		t.Fatal("expected active task")
	}
}

func TestCreateAndCountUrges(t *testing.T) {
	db := setupTestDB(t)

	err := db.CreateUrge(nil)
	if err != nil {
		t.Fatalf("CreateUrge failed: %v", err)
	}

	taskID, err := db.CreateTask("coding", nil)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = db.CreateUrge(&taskID)
	if err != nil {
		t.Fatalf("CreateUrge with task failed: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	count, err := db.CountUrges(start, end)
	if err != nil {
		t.Fatalf("CountUrges failed: %v", err)
	}
	if count != 2 {
		t.Errorf("urge count = %d, want 2", count)
	}
}

func TestGetUrgeCountForTask(t *testing.T) {
	db := setupTestDB(t)

	taskID, _ := db.CreateTask("coding", nil)
	db.CreateUrge(&taskID)
	db.CreateUrge(&taskID)
	db.CreateUrge(nil)

	count, err := db.GetUrgeCountForTask(taskID)
	if err != nil {
		t.Fatalf("GetUrgeCountForTask failed: %v", err)
	}
	if count != 2 {
		t.Errorf("urge count = %d, want 2", count)
	}
}

func TestPauseTask(t *testing.T) {
	db := setupTestDB(t)

	id, err := db.CreateTask("coding", nil)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = db.PauseTask(id)
	if err != nil {
		t.Fatalf("PauseTask failed: %v", err)
	}

	task, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if task.PausedAt == nil {
		t.Fatal("expected PausedAt to be set")
	}
}

func TestResumeTask(t *testing.T) {
	db := setupTestDB(t)

	id, err := db.CreateTask("coding", nil)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = db.PauseTask(id)
	if err != nil {
		t.Fatalf("PauseTask failed: %v", err)
	}

	// Small sleep so paused duration > 0
	time.Sleep(10 * time.Millisecond)

	err = db.ResumeTask(id)
	if err != nil {
		t.Fatalf("ResumeTask failed: %v", err)
	}

	task, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if task.PausedAt != nil {
		t.Fatal("expected PausedAt to be nil after resume")
	}
	if task.TotalPausedSeconds < 0 {
		t.Errorf("expected non-negative TotalPausedSeconds, got %d", task.TotalPausedSeconds)
	}
}

func TestStopWhilePaused(t *testing.T) {
	db := setupTestDB(t)

	id, err := db.CreateTask("coding", nil)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = db.PauseTask(id)
	if err != nil {
		t.Fatalf("PauseTask failed: %v", err)
	}

	err = db.StopTask(id)
	if err != nil {
		t.Fatalf("StopTask failed: %v", err)
	}

	task, err := db.GetActiveTask()
	if err != nil {
		t.Fatalf("GetActiveTask failed: %v", err)
	}
	if task != nil {
		t.Error("expected no active task after stop")
	}
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)

	id1, _ := db.CreateTask("task1", nil)
	db.StopTask(id1)
	id2, _ := db.CreateTask("task2", nil)
	db.StopTask(id2)

	db.CreateUrge(nil)
	db.CreateUrge(&id1)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)

	stats, err := db.GetStats(start, end)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.TasksCompleted != 2 {
		t.Errorf("tasks completed = %d, want 2", stats.TasksCompleted)
	}
	if stats.UrgesLogged != 2 {
		t.Errorf("urges logged = %d, want 2", stats.UrgesLogged)
	}
}
