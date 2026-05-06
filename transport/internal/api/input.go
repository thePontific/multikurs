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

// HandleInput обрабатывает POST /input (Input(params) от прикладного уровня)
func HandleInput(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var inputReq models.InputRequest
	if err = json.Unmarshal(body, &inputReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// request_id = timestamp (для идентификации сообщения)
	requestID := inputReq.Timestamp.Format(time.RFC3339Nano)

	fmt.Printf("📥 Input(params): от=%s, увеличение=%dx, образец=%s, request_id=%s\n",
		inputReq.Sender, inputReq.Magnification, inputReq.SampleID, requestID)

	// Здесь нужно отправить запрос агентному уровню через Transfer(params)
	// Формируем запрос к агенту
	agentRequest := map[string]interface{}{
		"request_id":    requestID,
		"magnification": inputReq.Magnification,
		"sample_id":     inputReq.SampleID,
	}

	// Отправляем агенту (вызов Transfer(params) → на агентный уровень)
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

// callAgentTransfer вызывает метод агентного уровня Transfer(params)
func callAgentTransfer(req map[string]interface{}) error {
	// Адрес агентного уровня (меняйте под свою сеть)
	agentURL := "http://192.168.123.100:5000/api/v1/transfer"

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
