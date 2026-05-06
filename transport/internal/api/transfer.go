package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"transport-service/internal/kafka"
	"transport-service/pkg/models"
)

// HandleTransfer обрабатывает POST /transfer (Transfer(params) от агентного уровня)
// Здесь мы получаем сегменты от агента и сохраняем их в Kafka (Producer)
func HandleTransfer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var segment models.Segment
	if err = json.Unmarshal(body, &segment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("📦 Transfer(params): request_id=%s, сегмент=%d/%d\n",
		segment.RequestID, segment.SegmentNumber, segment.TotalSegments)

	// отправляем сегмент в Kafka (Producer)
	if err := kafka.SendToKafka(segment); err != nil {
		fmt.Printf("Ошибка отправки в Kafka: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
	})
}
