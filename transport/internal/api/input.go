package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"transport-service/pkg/models"
)

func HandleInput(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var inputReq models.InputRequest
	if err = json.Unmarshal(body, &inputReq); err != nil {
		fmt.Printf("❌ Ошибка парсинга InputRequest: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requestID := inputReq.Timestamp.Format(time.RFC3339Nano)

	fmt.Printf("📥 Input(params): от=%s, увеличение=%dx, образец=%s, request_id=%s\n",
		inputReq.Sender, inputReq.Magnification, inputReq.SampleID, requestID)

	// Отправляем запрос агенту с Sender
	agentRequest := map[string]interface{}{
		"request_id":    requestID,
		"magnification": inputReq.Magnification,
		"sample_id":     inputReq.SampleID,
		"sender":        inputReq.Sender, // ← добавить sender
	}

	if err := callAgentTransfer(agentRequest); err != nil {
		fmt.Printf("Ошибка вызова агента: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "accepted",
		"request_id": requestID,
	})
}

func callAgentTransfer(req map[string]interface{}) error {
	//agentURL := "http://localhost:8080/transfer"
	agentURL := "http://agent:8080/transfer" // имя сервиса и правильный путь!
	jsonData, _ := json.Marshal(req)
	resp, err := http.Post(agentURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent returned %d", resp.StatusCode)
	}
	return nil
}
