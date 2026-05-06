package storage

import (
	"fmt"
	"sync"
	"time"
	"transport-service/pkg/models"
)

const (
	SegmentLostError = "lost"
	ReadPeriod       = 2 * time.Second // N секунд по варианту
)

// MessageInProgress — сообщение в процессе сборки
type MessageInProgress struct {
	Received      int                 // количество полученных сегментов
	Total         int                 // всего сегментов
	LastReceived  time.Time           // время последнего полученного сегмента
	Sender        string              // отправитель
	Magnification int                 // увеличение
	SampleID      string              // номер образца
	Segments      map[int]interface{} // карта: номер сегмента -> данные
}

var (
	storage = make(map[string]*MessageInProgress) // key = request_id
	mu      sync.Mutex
)

// AddSegment добавляет сегмент в хранилище
func AddSegment(segment models.Segment) {
	mu.Lock()
	defer mu.Unlock()

	reqID := segment.RequestID
	msg, exists := storage[reqID]

	if !exists {
		// первый сегмент — создаем запись
		storage[reqID] = &MessageInProgress{
			Received:      0,
			Total:         segment.TotalSegments,
			LastReceived:  time.Now().UTC(),
			Sender:        "", // будет заполнено позже (можно из первого сегмента)
			Magnification: segment.Magnification,
			SampleID:      segment.SampleID,
			Segments:      make(map[int]interface{}),
		}
		msg = storage[reqID]
	}

	// если пришел сегмент с ошибкой — помечаем всё сообщение как ошибочное
	if segment.ErrorFlag {
		msg.Received = msg.Total // искусственно считаем, что все сегменты "пришли"
		msg.Segments = nil
	} else {
		msg.Segments[segment.SegmentNumber] = segment.Data
		msg.Received++
		msg.LastReceived = time.Now().UTC()
	}

	storage[reqID] = msg
}

// assembleMessage склеивает сегменты в одно сообщение
func assembleMessage(reqID string) models.ReceiveRequest {
	msg := storage[reqID]

	// собираем все изображения в порядке сегментов
	var allImages []string
	for i := 1; i <= msg.Total; i++ {
		if segData, ok := msg.Segments[i]; ok {
			if dataMap, ok := segData.(map[string]interface{}); ok {
				if images, ok := dataMap["images"].([]interface{}); ok {
					for _, img := range images {
						if imgStr, ok := img.(string); ok {
							allImages = append(allImages, imgStr)
						}
					}
				}
			}
		}
	}

	receiveReq := models.ReceiveRequest{
		Sender:        msg.Sender,
		Timestamp:     mustParseTime(reqID), // request_id = timestamp
		ErrorFlag:     false,
		Magnification: msg.Magnification,
		SampleID:      msg.SampleID,
	}
	receiveReq.Payload.Images = allImages

	return receiveReq
}

// mustParseTime парсит request_id обратно в time.Time
func mustParseTime(reqID string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, reqID)
	return t
}

// ScanStorage сканирует хранилище и отправляет готовые сообщения на прикладной уровень
func ScanStorage(sendToAppLayer func(models.ReceiveRequest)) {
	mu.Lock()
	defer mu.Unlock()

	for reqID, msg := range storage {
		// случай 1: все сегменты получены
		if msg.Received == msg.Total {
			var receiveReq models.ReceiveRequest

			if msg.Segments == nil {
				// это был сегмент с ошибкой от агента
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

			fmt.Printf("✅ Собрано сообщение: request_id=%s, images=%d\n",
				reqID, len(receiveReq.Payload.Images))
			go sendToAppLayer(receiveReq)
			delete(storage, reqID)

			// случай 2: таймаут (потеря сегментов)
		} else if time.Since(msg.LastReceived) > ReadPeriod+1*time.Second {
			receiveReq := models.ReceiveRequest{
				Sender:        msg.Sender,
				Timestamp:     mustParseTime(reqID),
				ErrorFlag:     true, // признак ошибки из-за потери сегментов
				Magnification: msg.Magnification,
				SampleID:      msg.SampleID,
			}
			fmt.Printf("❌ Потеря сегментов: request_id=%s, получено=%d из %d\n",
				reqID, msg.Received, msg.Total)
			go sendToAppLayer(receiveReq)
			delete(storage, reqID)
		}
	}
}
