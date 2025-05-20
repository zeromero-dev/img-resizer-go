package storage

import (
	"fmt"
	"img-resizer/internal/config"
	"img-resizer/internal/models"
	"io"
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

// getPath returns the path for an image with the given ID and quality
func (s *LocalStorage) getPath(id string, quality models.ImageQuality) string {
	// Create a directory structure based on the first few characters of the ID
	// to avoid having too many files in a single directory
	prefix := id[:2]
	dir := filepath.Join(s.basePath, prefix)

	// Create directory if it doesn't exist
	os.MkdirAll(dir, 0755)

	return filepath.Join(dir, fmt.Sprintf("%s_%s.jpg", id, quality))
}

// Save saves an image to storage
func (s *LocalStorage) Save(id string, quality models.ImageQuality, reader io.Reader) (string, error) {
	path := s.getPath(id, quality)

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy data to file
	_, err = io.Copy(file, reader)
	if err != nil {
		return "", err
	}

	return path, nil
}

// Get retrieves an image from storage
func (s *LocalStorage) Get(id string, quality models.ImageQuality) (io.ReadCloser, error) {
	path := s.getPath(id, quality)

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Delete removes an image from storage
func (s *LocalStorage) Delete(id string, quality models.ImageQuality) error {
	path := s.getPath(id, quality)

	// Remove file
	return os.Remove(path)
}
