package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCaddyUpdateTaskPersistsProgress(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update-task.json")
	tasks := NewCaddyUpdateTasks(path)
	_, err := tasks.Start("download", "2.11.4", func(_ context.Context, report func(CaddyUpdateProgress)) (string, error) {
		report(CaddyUpdateProgress{
			Stage: "downloading", Attempt: 2, EffectiveURL: "https://mirror.example/caddy.zip",
			DownloadedBytes: 20, TotalBytes: 100, HTTPStatus: 206,
		})
		return "2.11.4", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForTask(t, tasks, "succeeded")
	restored := NewCaddyUpdateTasks(path).Current()
	if restored == nil || restored.Attempt != 2 || restored.EffectiveURL == "" || restored.Progress != 100 {
		t.Fatalf("恢复的任务不完整: %+v", restored)
	}
}

func waitForTask(t *testing.T, tasks *CaddyUpdateTasks, status string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if task := tasks.Current(); task != nil && task.Status == status {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("任务未进入 %s: %+v", status, tasks.Current())
}
