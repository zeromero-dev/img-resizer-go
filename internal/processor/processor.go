package processor

import (
	"bytes"
	"fmt"
	"img-resizer/internal/models"
	"io"

	"github.com/h2non/bimg"
)

// Processor handles image processing
type Processor struct{}

// NewProcessor creates a new image processor
func NewProcessor() *Processor {
	return &Processor{}
}

// ProcessImage processes an image and returns different quality variants
func (p *Processor) ProcessImage(original []byte) (map[models.ImageQuality][]byte, error) {
	// Check if the image is valid
	if !bimg.IsTypeSupported(bimg.DetermineImageType(original)) {
		return nil, fmt.Errorf("unsupported image type")
	}

	// Create a map to store different quality variants
	variants := make(map[models.ImageQuality][]byte)

	// Store the original image
	variants[models.QualityOriginal] = original

	// Process image with different quality levels
	qualities := []models.ImageQuality{
		models.QualityHigh,
		models.QualityMedium,
		models.QualityLow,
	}

	for _, quality := range qualities {
		// Convert quality string to int
		var qualityInt int
		switch quality {
		case models.QualityHigh:
			qualityInt = 75
		case models.QualityMedium:
			qualityInt = 50
		case models.QualityLow:
			qualityInt = 25
		default:
			continue
		}

		// Create options for processing
		options := bimg.Options{
			Quality: qualityInt,
			Type:    bimg.JPEG,
		}

		// Process image
		processed, err := bimg.NewImage(original).Process(options)
		if err != nil {
			return nil, fmt.Errorf("failed to process image with quality %s: %w", quality, err)
		}

		variants[quality] = processed
	}

	return variants, nil
}

// GetImageInfo returns information about an image
func (p *Processor) GetImageInfo(data []byte) (bimg.ImageSize, error) {
	return bimg.NewImage(data).Size()
}

// ReadAll reads all data from a reader
func (p *Processor) ReadAll(reader io.Reader) ([]byte, error) {
	return io.ReadAll(reader)
}

// CreateReader creates a reader from a byte slice
func (p *Processor) CreateReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}
