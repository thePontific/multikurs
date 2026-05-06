package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"transport-service/internal/kafka"
	"transport-service/pkg/models"
)

func HandleTransfer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("📨 Получен сегмент (raw): %s\n", string(body))

	var segment models.Segment
	if err = json.Unmarshal(body, &segment); err != nil {
		fmt.Printf("❌ Ошибка парсинга JSON: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ⚠️ ПОТЕРЯ СЕГМЕНТА С ВЕРОЯТНОСТЬЮ 7% (по варианту R=7)
	if rand.Intn(100) < 7 {
		fmt.Printf("💀 СЕГМЕНТ ПОТЕРЯН: request_id=%s, сегмент=%d/%d (7%% потерь)\n",
			segment.RequestID, segment.SegmentNumber, segment.TotalSegments)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "lost"})
		return
	}

	fmt.Printf("📦 Transfer(params): request_id=%s, сегмент=%d/%d, sender=%s\n",
		segment.RequestID, segment.SegmentNumber, segment.TotalSegments, segment.Sender)

	if err := kafka.SendToKafka(segment); err != nil {
		fmt.Printf("❌ Ошибка отправки в Kafka: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
	})
}
