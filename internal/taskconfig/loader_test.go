package taskconfig

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadRejectsInvalidConfigurations(t *testing.T) {
	valid := "title: Task\ndescription: Description\nflag: flag\nmaximum_points: 100\nminimum_points: 10\ndecay: 5\nstarts_at: \"2026-09-01T12:00:00\"\n"
	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{name: "no yaml", files: map[string]string{"readme.txt": "text"}, wantErr: "contains no YAML files"},
		{name: "unknown field", files: map[string]string{"task.yaml": valid + "unknown: true\n"}, wantErr: "field unknown not found"},
		{name: "timezone in task", files: map[string]string{"task.yaml": strings.Replace(valid, "12:00:00", "12:00:00+05:00", 1)}, wantErr: "invalid task start time"},
		{name: "duplicate slug", files: map[string]string{"one.yaml": "slug: duplicate\n" + valid, "two.yaml": "slug: duplicate\n" + valid}, wantErr: "is duplicated"},
		{name: "invalid points", files: map[string]string{"task.yaml": strings.Replace(valid, "minimum_points: 10", "minimum_points: 101", 1)}, wantErr: "minimum points"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			for name, contents := range tt.files {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := Load(directory, time.UTC)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
