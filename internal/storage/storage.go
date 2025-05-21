package storage

import (
	"fmt"
	"img-resizer/internal/config"
	"img-resizer/internal/models"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Storage defines the interface for image storage
type Storage interface {
	Save(id string, quality models.ImageQuality, reader io.Reader) (string, error)
	Get(id string, quality models.ImageQuality) (io.ReadCloser, error)
	Delete(id string, quality models.ImageQuality) error
}

// LocalStorage implements Storage interface for local file system
type LocalStorage struct {
	basePath string
}

// NewStorage creates a new storage based on configuration
func NewStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Storage.Type {
	case "local":
		return NewLocalStorage(cfg.Storage.LocalPath)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}

// NewLocalStorage creates a new local storage
func NewLocalStorage(basePath string) (Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

func (s *LocalStorage) getPath(id string, quality models.ImageQuality) (string, error) {
	prefix := id[:2]
	dir := filepath.Join(s.basePath, prefix)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return filepath.Join(dir, fmt.Sprintf("%s_%s.jpg", id, quality)), nil
}

func (s *LocalStorage) Save(id string, quality models.ImageQuality, reader io.Reader) (string, error) {
	path, err := s.getPath(id, quality)
	if err != nil {
		return "", err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file: %v", err)
		}
	}()

	_, err = io.Copy(file, reader)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (s *LocalStorage) Get(id string, quality models.ImageQuality) (io.ReadCloser, error) {
	path, err := s.getPath(id, quality)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *LocalStorage) Delete(id string, quality models.ImageQuality) error {
	path, err := s.getPath(id, quality)
	if err != nil {
		return err
	}

	return os.Remove(path)
}
