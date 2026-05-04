package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

type SendRequest struct {
	RequestID     int `json:"request_id"`
	SampleID      int `json:"sample_id"`
	Magnification int `json:"magnification"`
}

type SendResponse struct {
	Status string   `json:"status"`
	Images []string `json:"images,omitempty"`
	Error  string   `json:"error,omitempty"`
}

func initMinIO() error {
	endpoint := "localhost:9000"
	accessKeyID := "admin"
	secretAccessKey := "admin123456"
	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}
	minioClient = client
	log.Println("[MinIO] Подключено")
	return nil
}

func getImagesFromMinIO(sampleID int, magnification int) ([]string, error) {
	bucketName := "microscope-images"
	ctx := context.Background()
	fileName := fmt.Sprintf("sample_%d_mag_%d.jpg", sampleID, magnification)

	log.Printf("[MinIO] Ищем файл: %s", fileName)

	reader, err := minioClient.GetObject(ctx, bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("файл %s не найден: %v", fileName, err)
	}
	defer reader.Close()

	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Str)

	log.Printf("[MinIO] Загружено: %s (%d байт)", fileName, len(fileBytes))
	return []string{dataURL}, nil
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req SendRequest
	if err := json.Unmarshal(body, &req); err != nil {
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: err.Error()})
		return
	}

	log.Printf("[Send] Запрос: sample=%d, mag=%dx", req.SampleID, req.Magnification)

	// Получаем изображения из MinIO
	images, err := getImagesFromMinIO(req.SampleID, req.Magnification)
	if err != nil {
		log.Printf("[Error] %v", err)
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: err.Error()})
		return
	}

	// Возвращаем изображения напрямую
	json.NewEncoder(w).Encode(SendResponse{
		Status: "accepted",
		Images: images,
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	if err := initMinIO(); err != nil {
		log.Printf("MinIO не доступен: %v", err)
	}

	http.HandleFunc("/send", sendHandler)
	http.HandleFunc("/health", healthHandler)

	port := 8080
	log.Printf("Агент запущен на порту %d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
