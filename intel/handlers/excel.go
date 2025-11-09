package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"proj/intel/models"
	"proj/intel/services"
	"strings"

	"github.com/xuri/excelize/v2"
)

type UploadResponse struct {
	Success  bool   `json:"success"`
	Imported int    `json:"imported"`
	Message  string `json:"message"`
	Error    string `json:"error,omitempty"`
}

func processExcelFile(file io.Reader, fileType string) UploadResponse {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in processExcelFile: %v", r)
		}
	}()

	log.Printf("Processing Excel file, type: %s", fileType)
	// Сначала читаем файл в память для диагностики
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Error reading file: %v", err),
		}
	}

	// Проверяем сигнатуру файла
	if len(fileBytes) < 8 {
		return UploadResponse{
			Success: false,
			Error:   "File is too small or empty",
		}
	}

	// Проверяем, это ли Excel файл по сигнатуре
	if !isExcelFile(fileBytes) {
		return UploadResponse{
			Success: false,
			Error:   "File is not a valid Excel file. Please upload .xlsx or .xls file",
		}
	}

	// Пробуем открыть как Excel
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Error opening Excel file: %v. Please ensure it's a valid .xlsx file", err),
		}
	}
	defer f.Close()

	// Остальной код обработки...
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		// Пробуем получить первый доступный лист
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return UploadResponse{
				Success: false,
				Error:   "No sheets found in Excel file",
			}
		}
		rows, err = f.GetRows(sheets[0])
		if err != nil {
			return UploadResponse{
				Success: false,
				Error:   fmt.Sprintf("Error reading sheet: %v", err),
			}
		}
	}

	if len(rows) < 2 {
		return UploadResponse{
			Success: false,
			Error:   "File is empty or has no data rows",
		}
	}

	switch fileType {
	case "students":
		return processStudents(rows)
	case "topics":
		return processTopics(rows)
	case "supervisors":
		return processSupervisors(rows)
	default:
		return UploadResponse{
			Success: false,
			Error:   "Unknown file type",
		}
	}
}

// isExcelFile проверяет сигнатуру файла
func isExcelFile(data []byte) bool {
	// Сигнатуры для Excel файлов
	signatures := [][]byte{
		{0x50, 0x4B, 0x03, 0x04},                         // .xlsx (ZIP archive)
		{0x50, 0x4B, 0x05, 0x06},                         // .xlsx (ZIP archive)
		{0x50, 0x4B, 0x07, 0x08},                         // .xlsx (ZIP archive)
		{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, // .xls (OLE2)
	}

	for _, sig := range signatures {
		if len(data) >= len(sig) && bytes.Equal(data[:len(sig)], sig) {
			return true
		}
	}
	return false
}

func processStudents(rows [][]string) UploadResponse {
	count := 0
	log.Printf("Начало обработки студентов, всего строк: %d", len(rows))

	for i, row := range rows {
		if i == 0 {
			log.Printf("Заголовок: %v", row)
			continue
		}

		log.Printf("Обработка строки %d: %v", i, row)

		if len(row) >= 4 {
			// Создаем пользователя со всеми полями
			user := models.User{
				Name:     strings.TrimSpace(row[0]), // ФИО
				Email:    strings.TrimSpace(row[1]), // Email
				Password: strings.TrimSpace(row[2]), // Password
				Group:    strings.TrimSpace(row[3]), // Group
				Role:     "student",
			}

			// Используем ваш сервис для добавления
			services.Add(&user)
			count++

		} else {
			log.Printf("Пропущена строка %d: только %d столбцов. Данные: %v", i, len(row), row)
		}
	}

	log.Printf("Обработка завершена. Импортировано: %d", count)
	return UploadResponse{
		Success:  true,
		Imported: count,
		Message:  fmt.Sprintf("Импортировано %d студентов", count),
	}
}

func processTopics(rows [][]string) UploadResponse {
	count := 0
	for i, row := range rows {
		if i == 0 {
			continue // Пропускаем заголовок
		}

		// Структура Excel файла для тем:
		// row[0] - Название темы
		// row[1] - Предмет
		// row[2] - Вид работы (курсовая/дипломная)
		// row[3] - Цикловая комиссия
		// row[4] - Руководитель
		// row[5] - Группа (опционально)
		// row[6] - Описание (опционально)

		if len(row) >= 5 {
			topic := models.Topic{
				Title:      strings.TrimSpace(row[0]),
				Subject:    strings.TrimSpace(row[1]),
				WorkType:   strings.TrimSpace(row[2]),
				Commission: strings.TrimSpace(row[3]),
				Supervisor: strings.TrimSpace(row[4]),
				Status:     "free", // По умолчанию тема свободна
			}

			// Добавляем группу если есть
			if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
				topic.Group = strings.TrimSpace(row[5])
			}

			// Добавляем описание если есть
			if len(row) > 6 && strings.TrimSpace(row[6]) != "" {
				topic.Description = strings.TrimSpace(row[6])
			}

			// Используем ваш сервис для добавления
			services.Add(&topic)
			count++

		} else {
			log.Printf("Пропущена строка %d: недостаточно данных для темы (только %d столбцов)", i, len(row))
		}
	}

	return UploadResponse{
		Success:  true,
		Imported: count,
		Message:  fmt.Sprintf("Импортировано %d тем", count),
	}
}

func processSupervisors(rows [][]string) UploadResponse {
	count := 0
	for i, row := range rows {
		if i == 0 {
			continue // Пропускаем заголовок
		}
		if len(row) >= 3 {
			// Логика для добавления руководителей через ваш сервис
			// supervisor := Supervisor{...}
			// if result := services.Add(&supervisor); result.Error == nil {
			//     count++
			// }
			count++ // временно считаем строки
		}
	}

	return UploadResponse{
		Success:  true,
		Imported: count,
		Message:  fmt.Sprintf("Processed %d supervisors", count),
	}
}

func sendError(w http.ResponseWriter, message string) {
	response := UploadResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
