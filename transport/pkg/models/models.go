package models

import "time"

type InputRequest struct {
	Sender        string    `json:"sender"`
	Timestamp     time.Time `json:"timestamp"`
	Magnification int       `json:"magnification"`
	SampleID      string    `json:"sample_id"`
}

type Segment struct {
	RequestID     string `json:"request_id"`
	SegmentNumber int    `json:"segment_number"`
	TotalSegments int    `json:"total_segments"`
	ErrorFlag     bool   `json:"error_flag"`
	Magnification int    `json:"magnification"`
	SampleID      string `json:"sample_id"`
	Sender        string `json:"sender"` // ← ДОБАВИТЬ это поле
	Data          string `json:"data"`
}

type ReceiveRequest struct {
	Sender        string    `json:"sender"`
	Timestamp     time.Time `json:"timestamp"`
	ErrorFlag     bool      `json:"error_flag"`
	Magnification int       `json:"magnification"`
	SampleID      string    `json:"sample_id"`
	Payload       struct {
		Images []string `json:"images"`
	} `json:"payload"`
}
