package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-fiber-starter/internal/caddygen"
	"go-fiber-starter/internal/model/configversion"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStartupCaddyConfigPrefersActiveFile(t *testing.T) {
	database := configPersistenceDB(t)
	history, _ := caddygen.Generate(nil)
	if err := database.Create(&configversion.ConfigVersion{
		Version: 1, BusinessConfig: "[]", CaddyJSON: string(history), Status: ConfigStatusPublished,
	}).Error; err != nil {
		t.Fatal(err)
	}
	active := []byte(`{"apps":{"http":{"servers":{}}}}`)
	path := filepath.Join(t.TempDir(), "active.json")
	if err := os.WriteFile(path, active, 0o600); err != nil {
		t.Fatal(err)
	}
	payload, err := StartupCaddyConfig(context.Background(), database, path)
	if err != nil || !caddygen.HasManagementEntry(payload) {
		t.Fatalf("活动配置恢复失败: %v, %s", err, payload)
	}
}

func TestCaddyUpdateTasksRecordsFailure(t *testing.T) {
	tasks := &CaddyUpdateTasks{}
	task, err := tasks.Start("download", "2.11.4", func(context.Context, func(string, int64, int64)) (string, error) {
		return "", context.DeadlineExceeded
	})
	if err != nil || task.Status != "queued" {
		t.Fatalf("启动任务失败: %+v, %v", task, err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current := tasks.Current()
		if current.Status == "failed" {
			if current.ErrorMessage == "" || current.FinishedAt == nil {
				t.Fatalf("失败任务信息不完整: %+v", current)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("更新任务未结束")
}

func configPersistenceDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&configversion.ConfigVersion{}); err != nil {
		t.Fatal(err)
	}
	return database
}
