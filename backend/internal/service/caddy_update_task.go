package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrCaddyUpdateRunning = errors.New("已有 Caddy 更新或上传任务正在执行")

type CaddyUpdateTask struct {
	ID              string     `json:"id"`
	Kind            string     `json:"kind"`
	Status          string     `json:"status"`
	TargetVersion   string     `json:"target_version,omitempty"`
	Progress        int        `json:"progress"`
	DownloadedBytes int64      `json:"downloaded_bytes,omitempty"`
	TotalBytes      int64      `json:"total_bytes,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

type CaddyUpdateTasks struct {
	mu      sync.RWMutex
	current *CaddyUpdateTask
}

var caddyUpdateTasks = &CaddyUpdateTasks{}

func ManagedCaddyUpdateTasks() *CaddyUpdateTasks {
	return caddyUpdateTasks
}

func (tasks *CaddyUpdateTasks) Current() *CaddyUpdateTask {
	tasks.mu.RLock()
	defer tasks.mu.RUnlock()
	if tasks.current == nil {
		return nil
	}
	copy := *tasks.current
	return &copy
}

func (tasks *CaddyUpdateTasks) Start(kind, target string, work func(context.Context, func(string, int64, int64)) (string, error)) (*CaddyUpdateTask, error) {
	tasks.mu.Lock()
	if tasks.current != nil && !isFinishedTask(tasks.current.Status) {
		tasks.mu.Unlock()
		return nil, ErrCaddyUpdateRunning
	}
	task := &CaddyUpdateTask{
		ID: uuid.NewString(), Kind: kind, Status: "queued", TargetVersion: target,
		StartedAt: time.Now(),
	}
	tasks.current = task
	tasks.mu.Unlock()

	go tasks.run(task.ID, work)
	return tasks.Current(), nil
}

func (tasks *CaddyUpdateTasks) run(id string, work func(context.Context, func(string, int64, int64)) (string, error)) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	version, err := work(ctx, func(status string, downloaded, total int64) {
		tasks.updateProgress(id, status, downloaded, total)
	})
	tasks.finish(id, version, err)
}

func (tasks *CaddyUpdateTasks) updateProgress(id, status string, downloaded, total int64) {
	tasks.mu.Lock()
	defer tasks.mu.Unlock()
	if tasks.current == nil || tasks.current.ID != id {
		return
	}
	tasks.current.Status = status
	if downloaded > 0 {
		tasks.current.DownloadedBytes = downloaded
	}
	if total > 0 {
		tasks.current.TotalBytes = total
		tasks.current.Progress = min(99, int(downloaded*100/total))
	}
}

func (tasks *CaddyUpdateTasks) finish(id, version string, taskErr error) {
	tasks.mu.Lock()
	defer tasks.mu.Unlock()
	if tasks.current == nil || tasks.current.ID != id {
		return
	}
	now := time.Now()
	tasks.current.FinishedAt = &now
	if taskErr != nil {
		tasks.current.Status = "failed"
		tasks.current.ErrorMessage = taskErr.Error()
		return
	}
	tasks.current.Status = "succeeded"
	tasks.current.Progress = 100
	tasks.current.TargetVersion = version
}

func isFinishedTask(status string) bool {
	return status == "succeeded" || status == "failed"
}
