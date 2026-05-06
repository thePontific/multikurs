package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"transport-service/pkg/models"
)

const (
	SegmentLostError = "lost"
	ReadPeriod       = 2 * time.Second
)

type MessageInProgress struct {
	Received      int            `json:"received"`
	Total         int            `json:"total"`
	LastReceived  time.Time      `json:"last_received"`
	Sender        string         `json:"sender"`
	Magnification int            `json:"magnification"`
	SampleID      string         `json:"sample_id"`
	Segments      map[int][]byte `json:"segments"`
}

var (
	storage = make(map[string]*MessageInProgress)
	mu      sync.Mutex
)

// AddSegment добавляет сегмент в хранилище
func AddSegment(segment models.Segment) {
	mu.Lock()
	defer mu.Unlock()

	reqID := segment.RequestID
	msg, exists := storage[reqID]

	if !exists {
		// первый сегмент — создаем запись и сохраняем Sender
		storage[reqID] = &MessageInProgress{
			Received:      0,
			Total:         segment.TotalSegments,
			LastReceived:  time.Now().UTC(),
			Sender:        segment.Sender, // ← сохраняем Sender из сегмента
			Magnification: segment.Magnification,
			SampleID:      segment.SampleID,
			Segments:      make(map[int][]byte),
		}
		msg = storage[reqID]
	}

	if segment.ErrorFlag {
		msg.Received = msg.Total
		msg.Segments = nil
	} else {
		chunkData, err := base64.StdEncoding.DecodeString(segment.Data)
		if err != nil {
			fmt.Printf("⚠️ Ошибка декодирования сегмента %d: %v\n", segment.SegmentNumber, err)
			return
		}
		msg.Segments[segment.SegmentNumber] = chunkData
		msg.Received++
		msg.LastReceived = time.Now().UTC()
	}

	storage[reqID] = msg
}

// assembleMessage склеивает сегменты в одно сообщение
func assembleMessage(reqID string) models.ReceiveRequest {
	msg := storage[reqID]

	var fullData []byte
	for i := 1; i <= msg.Total; i++ {
		if chunk, ok := msg.Segments[i]; ok {
			fullData = append(fullData, chunk...)
		}
	}

	var images []string
	if err := json.Unmarshal(fullData, &images); err != nil {
		fmt.Printf("⚠️ Ошибка парсинга собранных данных: %v\n", err)
		images = []string{}
	}

	// Используем сохраненный Sender
	receiveReq := models.ReceiveRequest{
		Sender:        msg.Sender, // ← теперь Sender не пустой!
		Timestamp:     mustParseTime(reqID),
		ErrorFlag:     false,
		Magnification: msg.Magnification,
		SampleID:      msg.SampleID,
	}
	receiveReq.Payload.Images = images

	fmt.Printf("📦 Собрано %d изображений для request_id=%s, sender=%s\n", len(images), reqID, msg.Sender)

	return receiveReq
}

func mustParseTime(reqID string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, reqID)
	return t
}

// ScanStorage сканирует хранилище и отправляет готовые сообщения на прикладной уровень
func ScanStorage(sendToAppLayer func(models.ReceiveRequest)) {
	mu.Lock()
	defer mu.Unlock()

	for reqID, msg := range storage {
		if msg.Received == msg.Total {
			var receiveReq models.ReceiveRequest

			if msg.Segments == nil {
				receiveReq = models.ReceiveRequest{
					Sender:        msg.Sender,
					Timestamp:     mustParseTime(reqID),
					ErrorFlag:     true,
					Magnification: msg.Magnification,
					SampleID:      msg.SampleID,
				}
			} else {
				receiveReq = assembleMessage(reqID)
			}

			fmt.Printf("✅ Готово к отправке: request_id=%s, sender=%s, error_flag=%v\n",
				reqID, receiveReq.Sender, receiveReq.ErrorFlag)
			go sendToAppLayer(receiveReq)
			delete(storage, reqID)

		} else if time.Since(msg.LastReceived) > ReadPeriod+1*time.Second {
			receiveReq := models.ReceiveRequest{
				Sender:        msg.Sender,
				Timestamp:     mustParseTime(reqID),
				ErrorFlag:     true,
				Magnification: msg.Magnification,
				SampleID:      msg.SampleID,
			}
			fmt.Printf("❌ Потеря сегментов: request_id=%s, получено=%d из %d, sender=%s\n",
				reqID, msg.Received, msg.Total, msg.Sender)
			go sendToAppLayer(receiveReq)
			delete(storage, reqID)
		}
	}
}
