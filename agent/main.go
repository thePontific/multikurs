package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

type SendRequest struct {
	RequestID     interface{} `json:"request_id"`
	SampleID      interface{} `json:"sample_id"`
	Magnification int         `json:"magnification"`
	Sender        string      `json:"sender"` // ← ДОБАВИТЬ поле sender
}

type SendResponse struct {
	Status string   `json:"status"`
	Images []string `json:"images,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// Segment — сегмент для отправки на транспортный уровень
type Segment struct {
	RequestID     string `json:"request_id"`
	SegmentNumber int    `json:"segment_number"`
	TotalSegments int    `json:"total_segments"`
	ErrorFlag     bool   `json:"error_flag"`
	Magnification int    `json:"magnification"`
	SampleID      string `json:"sample_id"`
	Sender        string `json:"sender"` // ← добавить
	Data          string `json:"data"`
}

const (
	TRANSPORT_URL = "http://transport:8000"
	SEGMENT_SIZE  = 50000 // байт
)

func initMinIO() error {
	endpoint := "minio:9000"
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

func toInt(val interface{}) (int, error) {
	switch v := val.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("неизвестный тип: %T", val)
	}
}

func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return fmt.Sprintf("%v", v)
	}
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

// Разбивает данные на сегменты по maxSize байт
func splitIntoChunks(data []byte, maxSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += maxSize {
		end := i + maxSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

// Отправляет сегменты на транспортный уровень
func sendSegmentsToTransport(requestID string, imagesData []byte, magnification int, sampleID int, sender string) error {
	// Разбиваем на сегменты
	chunks := splitIntoChunks(imagesData, SEGMENT_SIZE)

	log.Printf("[Segment] Размер данных: %d байт, разбито на %d сегментов (каждый до %d байт)",
		len(imagesData), len(chunks), SEGMENT_SIZE)

	// Отправляем каждый сегмент
	for i, chunk := range chunks {
		// Кодируем chunk в base64 для безопасной передачи в JSON
		chunkBase64 := base64.StdEncoding.EncodeToString(chunk)

		segment := Segment{
			RequestID:     requestID,
			SegmentNumber: i + 1,
			TotalSegments: len(chunks),
			ErrorFlag:     false,
			Magnification: magnification,
			SampleID:      strconv.Itoa(sampleID),
			Sender:        sender, // ← теперь sender передается из запроса
			Data:          chunkBase64,
		}

		segmentJSON, err := json.Marshal(segment)
		if err != nil {
			log.Printf("[Error] Ошибка сериализации сегмента %d: %v", i+1, err)
			continue
		}

		// Отправляем на транспортный уровень
		resp, err := http.Post(TRANSPORT_URL+"/send", "application/json", bytes.NewBuffer(segmentJSON))
		if err != nil {
			log.Printf("[Error] Ошибка отправки сегмента %d: %v", i+1, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("[Segment] Отправлен сегмент %d/%d", i+1, len(chunks))
		} else {
			log.Printf("[Segment] Ошибка сегмента %d, статус: %d", i+1, resp.StatusCode)
		}

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: err.Error()})
		return
	}
	defer r.Body.Close()

	var req SendRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[Error] Ошибка парсинга JSON: %v", err)
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: err.Error()})
		return
	}

	sampleID, err := toInt(req.SampleID)
	if err != nil {
		log.Printf("[Error] Неверный sample_id: %v", err)
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: "Неверный sample_id"})
		return
	}

	requestID := toString(req.RequestID)
	sender := req.Sender // ← получаем sender из запроса

	log.Printf("[Send] Запрос: request_id=%s, sender=%s, sample=%d, mag=%dx", requestID, sender, sampleID, req.Magnification)

	// Получаем изображения из MinIO
	images, err := getImagesFromMinIO(sampleID, req.Magnification)
	if err != nil {
		log.Printf("[Error] %v", err)
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: err.Error()})
		return
	}

	// Сериализуем изображения в JSON
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		log.Printf("[Error] Ошибка сериализации изображений: %v", err)
		json.NewEncoder(w).Encode(SendResponse{Status: "error", Error: err.Error()})
		return
	}

	// Отправляем сегменты на транспортный уровень (асинхронно) с передачей sender
	go sendSegmentsToTransport(requestID, imagesJSON, req.Magnification, sampleID, sender)

	// Возвращаем ответ, что запрос принят
	json.NewEncoder(w).Encode(SendResponse{
		Status: "accepted",
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

	http.HandleFunc("/transfer", sendHandler)
	http.HandleFunc("/health", healthHandler)

	port := 8080
	log.Printf("╔══════════════════════════════════════════════════════════════╗")
	log.Printf("║     Агентный уровень (с разбиением на сегменты) ЗАПУЩЕН    ║")
	log.Printf("╠══════════════════════════════════════════════════════════════╣")
	log.Printf("║  HTTP:      http://localhost:%d/send                         ║")
	log.Printf("║  Health:    http://localhost:%d/health                       ║")
	log.Printf("║  Транспорт: %s/transfer                                      ║")
	log.Printf("║  Размер сегмента: %d байт                                    ║")
	log.Printf("╚══════════════════════════════════════════════════════════════╝", port, port, TRANSPORT_URL, SEGMENT_SIZE)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), nil))
}
