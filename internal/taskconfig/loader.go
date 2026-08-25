package taskconfig

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"RedButton-bot/internal/dto"
	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type taskFile struct {
	Path     string `yaml:"path"`
	Name     string `yaml:"name"`
	MIMEType string `yaml:"mime_type"`
}

type task struct {
	Slug          string    `yaml:"slug"`
	Title         string    `yaml:"title"`
	Description   string    `yaml:"description"`
	Flag          string    `yaml:"flag"`
	MaximumPoints int       `yaml:"maximum_points"`
	MinimumPoints int       `yaml:"minimum_points"`
	Decay         int       `yaml:"decay"`
	StartsAt      string    `yaml:"starts_at"`
	File          *taskFile `yaml:"file"`
}

// Функция загрузки тасков из YAML-файлов.
// Возвращает проверенные определения тасков и ошибку.
func Load(directory string) ([]dto.TaskDefinition, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("прочитать директорию тасков: %w", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension == ".yaml" || extension == ".yml" {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("в директории %q отсутствуют YAML-файлы тасков", directory)
	}
	sort.Strings(paths)

	definitions := make([]dto.TaskDefinition, 0, len(paths))
	slugs := make(map[string]string, len(paths))
	for _, path := range paths {
		definition, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		if previous, exists := slugs[definition.Slug]; exists {
			return nil, fmt.Errorf(
				"идентификатор таска %q повторяется в %q и %q",
				definition.Slug,
				previous,
				path,
			)
		}
		slugs[definition.Slug] = path
		definitions = append(definitions, definition)
	}

	return definitions, nil
}

func loadFile(path string) (dto.TaskDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return dto.TaskDefinition{}, fmt.Errorf("открыть YAML-файл %q: %w", path, err)
	}
	defer file.Close()

	var source task
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&source); err != nil {
		return dto.TaskDefinition{}, fmt.Errorf("прочитать YAML-файл %q: %w", path, err)
	}

	definition, err := mapTask(path, source)
	if err != nil {
		return dto.TaskDefinition{}, fmt.Errorf("проверить YAML-файл %q: %w", path, err)
	}

	return definition, nil
}

func mapTask(path string, source task) (dto.TaskDefinition, error) {
	slug := strings.TrimSpace(source.Slug)
	if slug == "" {
		slug = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if !slugPattern.MatchString(slug) {
		return dto.TaskDefinition{}, fmt.Errorf("некорректный slug %q", slug)
	}
	if strings.TrimSpace(source.Title) == "" {
		return dto.TaskDefinition{}, errors.New("название таска не задано")
	}
	if strings.TrimSpace(source.Description) == "" {
		return dto.TaskDefinition{}, errors.New("описание таска не задано")
	}
	if source.Flag == "" {
		return dto.TaskDefinition{}, errors.New("флаг таска не задан")
	}
	if source.MaximumPoints <= 0 {
		return dto.TaskDefinition{}, errors.New("максимальная стоимость должна быть больше нуля")
	}
	if source.MinimumPoints < 0 || source.MinimumPoints > source.MaximumPoints {
		return dto.TaskDefinition{}, errors.New("минимальная стоимость должна быть от нуля до максимальной")
	}
	if source.Decay <= 0 {
		return dto.TaskDefinition{}, errors.New("распад должен быть больше нуля")
	}

	startsAt, err := time.Parse(time.RFC3339, source.StartsAt)
	if err != nil {
		return dto.TaskDefinition{}, fmt.Errorf("некорректное время открытия: %w", err)
	}

	definition := dto.TaskDefinition{
		Slug:          slug,
		Title:         strings.TrimSpace(source.Title),
		Description:   strings.TrimSpace(source.Description),
		Flag:          source.Flag,
		MaximumPoints: source.MaximumPoints,
		MinimumPoints: source.MinimumPoints,
		Decay:         source.Decay,
		StartsAt:      startsAt,
	}
	if source.File != nil {
		file, err := mapFile(filepath.Dir(path), *source.File)
		if err != nil {
			return dto.TaskDefinition{}, err
		}
		definition.File = &file
	}

	return definition, nil
}

func mapFile(directory string, source taskFile) (dto.TaskFileDefinition, error) {
	if strings.TrimSpace(source.Path) == "" {
		return dto.TaskFileDefinition{}, errors.New("путь к файлу таска не задан")
	}

	storagePath := source.Path
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(directory, storagePath)
	}
	storagePath, err := filepath.Abs(storagePath)
	if err != nil {
		return dto.TaskFileDefinition{}, fmt.Errorf("получить полный путь к файлу: %w", err)
	}
	info, err := os.Stat(storagePath)
	if err != nil {
		return dto.TaskFileDefinition{}, fmt.Errorf("проверить файл таска: %w", err)
	}
	if !info.Mode().IsRegular() {
		return dto.TaskFileDefinition{}, fmt.Errorf("путь %q не является файлом", storagePath)
	}

	fileName := strings.TrimSpace(source.Name)
	if fileName == "" {
		fileName = filepath.Base(storagePath)
	}
	mimeType := strings.TrimSpace(source.MIMEType)
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileName))
	}

	var mimeTypePointer *string
	if mimeType != "" {
		mimeTypePointer = &mimeType
	}

	return dto.TaskFileDefinition{
		StoragePath: storagePath,
		FileName:    fileName,
		MIMEType:    mimeTypePointer,
		FileSize:    info.Size(),
	}, nil
}
