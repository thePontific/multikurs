package models

import "time"

// InputRequest — запрос от прикладного уровня (метод Input(params))
type InputRequest struct {
	Sender        string    `json:"sender"`
	Timestamp     time.Time `json:"timestamp"`
	Magnification int       `json:"magnification"` // вариант 7: x10, x40, x100, x400, x1000
	SampleID      string    `json:"sample_id"`     // вариант 7: номер образца
}

// Segment — сегмент для Kafka (получаем от агента через Send(segment))
type Segment struct {
	RequestID     string `json:"request_id"`     // = timestamp из InputRequest
	SegmentNumber int    `json:"segment_number"` // номер сегмента (1..N)
	TotalSegments int    `json:"total_segments"` // всего сегментов
	ErrorFlag     bool   `json:"error_flag"`     // true, если агент не нашел снимки
	Magnification int    `json:"magnification"`  // вариант 7
	SampleID      string `json:"sample_id"`      // вариант 7
	Data          struct {
		Images []string `json:"images"` // снимки в base64
	} `json:"data"`
}

// ReceiveRequest — отправка на прикладной уровень (метод Receive(message))
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
