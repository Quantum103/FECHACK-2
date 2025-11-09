package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Проверяем метод
	if r.Method != "POST" {
		http.Error(w, "Только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получаем файл из формы
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Читаем файл в память
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	// 4. Отправляем в Python для обработки
	result, err := sendToPython(fileData, header.Filename)
	if err != nil {
		http.Error(w, "Ошибка обработки: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func sendToPython(fileData []byte, filename string) (map[string]interface{}, error) {
	// Создаем multipart форму
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Добавляем файл
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	part.Write(fileData)
	writer.Close()

	// Отправляем в Python
	req, err := http.NewRequest("POST", "http://localhost:8000/process", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}
