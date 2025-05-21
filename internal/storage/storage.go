package storage

import (
	"fmt"
	"img-resizer/internal/config"
	"img-resizer/internal/models"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage defines the interface for image storage
type Storage interface {
	Save(id string, quality models.ImageQuality, reader io.Reader) (string, error)
	Get(id string, quality models.ImageQuality) (io.ReadCloser, error)
	Delete(id string, quality models.ImageQuality) error
	List() ([]string, error)
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
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("failed to close file after save error: %v", cerr)
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

func (s *LocalStorage) List() ([]string, error) {
	var images []string
	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".jpg" {
			// Extract ID from filename (remove quality suffix and extension)
			filename := filepath.Base(path)
			parts := strings.Split(filename, "_")
			if len(parts) > 0 {
				id := parts[0]
				if !contains(images, id) {
					images = append(images, id)
				}
			}
		}
		return nil
	})
	return images, err
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
