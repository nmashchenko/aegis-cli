package session

import (
	"testing"

	"github.com/nmashchenko/aegis-cli/internal/db"
)

func setupTest(t *testing.T) *Service {
	t.Helper()
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return NewService(database)
}

func TestStartTask(t *testing.T) {
	svc := setupTest(t)

	result, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if result.TaskID == 0 {
		t.Fatal("expected non-zero task ID")
	}
	if result.TaskName != "coding" {
		t.Errorf("task name = %q, want %q", result.TaskName, "coding")
	}
}

func TestStartPreventsDuplicate(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("first Start failed: %v", err)
	}

	_, err = svc.Start("reading", nil)
	if err == nil {
		t.Fatal("expected error on double start, got nil")
	}
}

func TestStopTask(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	result, err := svc.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if result.TaskName != "coding" {
		t.Errorf("task name = %q, want %q", result.TaskName, "coding")
	}
}

func TestStopWithNoActiveTask(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Stop()
	if err == nil {
		t.Fatal("expected error on stop with no active task")
	}
}

func TestStatus(t *testing.T) {
	svc := setupTest(t)

	status, err := svc.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status.Active {
		t.Error("expected inactive status")
	}

	_, err = svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	status, err = svc.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if !status.Active {
		t.Error("expected active status")
	}
	if status.TaskName != "coding" {
		t.Errorf("task name = %q, want %q", status.TaskName, "coding")
	}
}

func TestLogUrge(t *testing.T) {
	svc := setupTest(t)

	result, err := svc.LogUrge()
	if err != nil {
		t.Fatalf("LogUrge failed: %v", err)
	}
	if result.TaskName != "" {
		t.Errorf("expected empty task name, got %q", result.TaskName)
	}

	_, err = svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	result, err = svc.LogUrge()
	if err != nil {
		t.Fatalf("LogUrge failed: %v", err)
	}
	if result.TaskName != "coding" {
		t.Errorf("task name = %q, want %q", result.TaskName, "coding")
	}
}

func TestPauseTask(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	result, err := svc.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}
	if result.TaskName != "coding" {
		t.Errorf("task name = %q, want %q", result.TaskName, "coding")
	}
}

func TestPauseWithNoActiveTask(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Pause()
	if err == nil {
		t.Fatal("expected error on pause with no active task")
	}
}

func TestResumeTask(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	_, err = svc.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	result, err := svc.Resume()
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	if result.TaskName != "coding" {
		t.Errorf("task name = %q, want %q", result.TaskName, "coding")
	}
}

func TestResumeWithNoActiveTask(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Resume()
	if err == nil {
		t.Fatal("expected error on resume with no active task")
	}
}

func TestResumeWhenNotPaused(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	_, err = svc.Resume()
	if err == nil {
		t.Fatal("expected error on resume when not paused")
	}
}

func TestStatusWhilePaused(t *testing.T) {
	svc := setupTest(t)

	_, err := svc.Start("coding", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	_, err = svc.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	status, err := svc.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if !status.Active {
		t.Error("expected active status")
	}
	if !status.Paused {
		t.Error("expected paused status")
	}
}
