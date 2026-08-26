package taskconfig

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	location := time.FixedZone("UTC+5", 5*60*60)
	tasks, err := Load("../../tasks", location)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) == 0 {
		t.Fatal("Load() returned no tasks")
	}

	var exampleIndex = -1
	for index := range tasks {
		if tasks[index].Slug == "example" {
			exampleIndex = index
			break
		}
	}
	if exampleIndex == -1 {
		t.Fatal("example task not found")
	}
	task := tasks[exampleIndex]
	if task.MaximumPoints != 1000 || task.MinimumPoints != 100 || task.Decay != 25 {
		t.Errorf("points configuration = %d/%d/%d, want 1000/100/25", task.MaximumPoints, task.MinimumPoints, task.Decay)
	}
	wantStartsAt := time.Date(2026, time.September, 1, 12, 0, 0, 0, location)
	if !task.StartsAt.Equal(wantStartsAt) {
		t.Errorf("StartsAt = %v, want %v", task.StartsAt, wantStartsAt)
	}
	if task.File == nil || filepath.Base(task.File.StoragePath) != "example.txt" {
		t.Errorf("File = %#v, want example.txt", task.File)
	}
}
