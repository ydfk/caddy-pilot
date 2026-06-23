package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrCaddyUpdateRunning = errors.New("已有 Caddy 更新或上传任务正在执行")

type CaddyUpdateTask struct {
	ID              string     `json:"id"`
	Kind            string     `json:"kind"`
	Status          string     `json:"status"`
	Stage           string     `json:"stage,omitempty"`
	TargetVersion   string     `json:"target_version,omitempty"`
	Progress        int        `json:"progress"`
	Attempt         int        `json:"attempt,omitempty"`
	EffectiveURL    string     `json:"effective_url,omitempty"`
	DownloadedBytes int64      `json:"downloaded_bytes,omitempty"`
	TotalBytes      int64      `json:"total_bytes,omitempty"`
	HTTPStatus      int        `json:"http_status,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

type CaddyUpdateTasks struct {
	mu      sync.RWMutex
	current *CaddyUpdateTask
	path    string
	loaded  bool
}

var caddyUpdateTasks = &CaddyUpdateTasks{}

func ManagedCaddyUpdateTasks() *CaddyUpdateTasks {
	return caddyUpdateTasks
}

func NewCaddyUpdateTasks(path string) *CaddyUpdateTasks {
	return &CaddyUpdateTasks{path: path}
}

func (tasks *CaddyUpdateTasks) Current() *CaddyUpdateTask {
	tasks.mu.Lock()
	defer tasks.mu.Unlock()
	tasks.loadLocked()
	if tasks.current == nil {
		return nil
	}
	copy := *tasks.current
	return &copy
}

func (tasks *CaddyUpdateTasks) Start(kind, target string, work func(context.Context, func(CaddyUpdateProgress)) (string, error)) (*CaddyUpdateTask, error) {
	tasks.mu.Lock()
	tasks.loadLocked()
	if tasks.current != nil && !isFinishedTask(tasks.current.Status) {
		tasks.mu.Unlock()
		return nil, ErrCaddyUpdateRunning
	}
	task := &CaddyUpdateTask{
		ID: uuid.NewString(), Kind: kind, Status: "queued", Stage: "queued",
		TargetVersion: target, StartedAt: time.Now(),
	}
	tasks.current = task
	tasks.persistLocked()
	tasks.mu.Unlock()

	go tasks.run(task.ID, work)
	return tasks.Current(), nil
}

func (tasks *CaddyUpdateTasks) run(id string, work func(context.Context, func(CaddyUpdateProgress)) (string, error)) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	version, err := work(ctx, func(progress CaddyUpdateProgress) {
		tasks.updateProgress(id, progress)
	})
	tasks.finish(id, version, err)
}

func (tasks *CaddyUpdateTasks) updateProgress(id string, progress CaddyUpdateProgress) {
	tasks.mu.Lock()
	defer tasks.mu.Unlock()
	if tasks.current == nil || tasks.current.ID != id {
		return
	}
	if progress.Stage != "" {
		tasks.current.Stage = progress.Stage
		if isVisibleTaskStatus(progress.Stage) {
			tasks.current.Status = progress.Stage
		}
	}
	tasks.current.Attempt = progress.Attempt
	tasks.current.DownloadedBytes = progress.DownloadedBytes
	tasks.current.TotalBytes = progress.TotalBytes
	tasks.current.HTTPStatus = progress.HTTPStatus
	if progress.EffectiveURL != "" {
		tasks.current.EffectiveURL = progress.EffectiveURL
	}
	if progress.TotalBytes > 0 {
		tasks.current.Progress = min(99, int(progress.DownloadedBytes*100/progress.TotalBytes))
	}
	tasks.persistLocked()
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
		var downloadErr *CaddyDownloadError
		if errors.As(taskErr, &downloadErr) {
			tasks.current.Stage = downloadErr.Stage
			tasks.current.Attempt = downloadErr.Attempt
			tasks.current.EffectiveURL = downloadErr.EffectiveURL
			tasks.current.DownloadedBytes = downloadErr.DownloadedBytes
			tasks.current.HTTPStatus = downloadErr.HTTPStatus
		}
		tasks.persistLocked()
		return
	}
	tasks.current.Status = "succeeded"
	tasks.current.Stage = "succeeded"
	tasks.current.Progress = 100
	tasks.current.TargetVersion = version
	tasks.persistLocked()
}

func (tasks *CaddyUpdateTasks) loadLocked() {
	if tasks.loaded {
		return
	}
	tasks.loaded = true
	payload, err := os.ReadFile(tasks.persistencePath())
	if err != nil {
		return
	}
	var task CaddyUpdateTask
	if json.Unmarshal(payload, &task) != nil || task.ID == "" {
		return
	}
	if !isFinishedTask(task.Status) {
		now := time.Now()
		task.Status = "failed"
		task.ErrorMessage = "CaddyPilot 重启，之前的更新任务已中断，可重新发起更新"
		task.FinishedAt = &now
	}
	tasks.current = &task
	tasks.persistLocked()
}

func (tasks *CaddyUpdateTasks) persistLocked() {
	if tasks.current == nil {
		return
	}
	payload, err := json.MarshalIndent(tasks.current, "", "  ")
	if err != nil {
		return
	}
	_ = writeFileAtomic(tasks.persistencePath(), payload, 0o600)
}

func (tasks *CaddyUpdateTasks) persistencePath() string {
	if tasks.path != "" {
		return tasks.path
	}
	runtimeDir := environmentValue("CADDYPILOT_RUNTIME_DIR", filepath.Join("data", "runtime"))
	return filepath.Join(runtimeDir, "caddy", "update-task.json")
}

func isFinishedTask(status string) bool {
	return status == "succeeded" || status == "failed"
}

func isVisibleTaskStatus(status string) bool {
	switch status {
	case "queued", "downloading", "verifying", "installing", "restarting":
		return true
	default:
		return false
	}
}
