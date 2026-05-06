package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"transport-service/internal/api"
	"transport-service/internal/kafka"
	"transport-service/internal/storage"
	"transport-service/pkg/models"

	"github.com/gorilla/mux"
)

const (
	AppLayerReceiveURL = "http://192.168.123.140:8080/api/v1/receive"
)

// sendToAppLayer отправляет собранное сообщение на прикладной уровень (Receive(message))
func sendToAppLayer(req models.ReceiveRequest) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Ошибка маршалинга: %v\n", err)
		return
	}

	resp, err := http.Post(AppLayerReceiveURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка отправки на прикладной уровень: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Отправлено на прикладной уровень: error_flag=%v\n", req.ErrorFlag)
	} else {
		fmt.Printf("⚠️ Прикладной уровень вернул код: %d\n", resp.StatusCode)
	}
}

func main() {
	// 1. Запуск Kafka consumer в горутине
	go func() {
		if err := kafka.StartConsumer(storage.AddSegment); err != nil {
			fmt.Printf("Consumer stopped: %v\n", err)
		}
	}()

	// 2. Горутина для периодического сканирования хранилища (отложенная отправка)
	go func() {
		ticker := time.NewTicker(storage.ReadPeriod)
		defer ticker.Stop()
		for {
			<-ticker.C
			storage.ScanStorage(sendToAppLayer)
		}
	}()

	// 3. Настройка HTTP сервера
	r := mux.NewRouter()
	r.HandleFunc("/input", api.HandleInput).Methods(http.MethodPost, http.MethodOptions)       // Input(params)
	r.HandleFunc("/transfer", api.HandleTransfer).Methods(http.MethodPost, http.MethodOptions) // Transfer(params)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	http.Handle("/", r)

	srv := &http.Server{
		Handler:           r,
		Addr:              ":8000",
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("🚀 Транспортный уровень запущен на :8000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// 4. Graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh

	fmt.Println("\n🛑 Получен сигнал завершения, остановка сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при остановке: %v\n", err)
	}
	fmt.Println("✅ Сервис остановлен")
}
