package models

import "time"

// ImageQuality represents the quality level of an image
type ImageQuality string

const (
	// QualityOriginal represents the original image quality
	QualityOriginal ImageQuality = "100"
	// QualityHigh represents 75% of the original quality
	QualityHigh ImageQuality = "75"
	// QualityMedium represents 50% of the original quality
	QualityMedium ImageQuality = "50"
	// QualityLow represents 25% of the original quality
	QualityLow ImageQuality = "25"
)

// ImageMetadata represents metadata for an image
type ImageMetadata struct {
	ID          string       `json:"id"`
	OriginalName string      `json:"originalName"`
	MimeType    string       `json:"mimeType"`
	Size        int64        `json:"size"`
	Width       int          `json:"width"`
	Height      int          `json:"height"`
	CreatedAt   time.Time    `json:"createdAt"`
	Qualities   []ImageQuality `json:"qualities"`
}

// ImageProcessingTask represents a task for processing an image
type ImageProcessingTask struct {
	ID        string `json:"id"`
	FilePath  string `json:"filePath"`
}