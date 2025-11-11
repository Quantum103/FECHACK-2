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
			continue
		}

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
			continue
		}
		if len(row) >= 3 {
			count++
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

func exportHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Обработчик exportHandler вызван")

	if r.Method != "POST" {
		log.Println("Метод не POST, показываем форму")
		showExportForm(w)
		return
	}

	// Обязательно парсим форму
	if err := r.ParseForm(); err != nil {
		log.Printf("Ошибка парсинга формы: %v", err)
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	group := r.FormValue("group")
	log.Printf("Получена группа: %s", group)

	if group == "" {
		http.Error(w, "Группа не указана", http.StatusBadRequest)
		return
	}

	// Получаем пользователей из БД
	var users []models.User
	db := services.GetDB()

	// Используем экранирование для поля group
	result := db.Where("`group` = ?", group).Find(&users)
	if result.Error != nil {
		log.Printf("Ошибка БД: %v", result.Error)
		http.Error(w, "Ошибка базы данных: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		log.Printf("Нет данных для группы: %s", group)
		http.Error(w, "Для группы '"+group+"' нет данных", http.StatusNotFound)
		return
	}

	log.Printf("Найдено %d пользователей", len(users))

	// Создаем Excel файл
	f := excelize.NewFile()

	// Устанавливаем заголовки
	f.SetCellValue("Sheet1", "A1", "Имя")
	f.SetCellValue("Sheet1", "B1", "Тема")
	f.SetCellValue("Sheet1", "C1", "Роль")

	// Заполняем данные
	for i, user := range users {
		row := i + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), user.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), user.Topic)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), user.Role)
	}

	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=group_%s.xlsx", group))
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// Записываем файл в ответ
	if err := f.Write(w); err != nil {
		log.Printf("Ошибка записи Excel: %v", err)
		http.Error(w, "Ошибка создания файла", http.StatusInternalServerError)
		return
	}

	log.Printf("Успешно отправлен Excel файл для группы %s", group)
}

func showExportForm(w http.ResponseWriter) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Выгрузка данных</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 500px; margin: 0 auto; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Выгрузка данных студентов</h2>
        <form method="POST" action="/export">
            <div class="form-group">
                <label for="group">Введите группу:</label>
                <input type="text" id="group" name="group" required placeholder="Например: ИС-202">
            </div>
            <button type="submit">Скачать Excel</button>
        </form>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Ошибка отправки формы: %v", err)
	}
}

func exportFormHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "exportList.html", nil)
}

// excel.go - добавьте эти функции

func exportSupervisorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Обработчик exportSupervisorHandler вызван")

	if r.Method != "POST" {
		log.Println("Метод не POST, показываем форму для руководителя")
		showExportSupervisorForm(w)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("Ошибка парсинга формы: %v", err)
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	supervisor := r.FormValue("supervisor")
	log.Printf("Получен руководитель: %s", supervisor)

	if supervisor == "" {
		http.Error(w, "Руководитель не указан", http.StatusBadRequest)
		return
	}

	// Получаем темы из БД по руководителю
	var topics []models.Topic
	db := services.GetDB()

	result := db.Where("supervisor = ?", supervisor).Find(&topics)
	if result.Error != nil {
		log.Printf("Ошибка БД при поиске тем: %v", result.Error)
		http.Error(w, "Ошибка базы данных: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if len(topics) == 0 {
		log.Printf("Нет тем для руководителя: %s", supervisor)
		http.Error(w, "Для руководителя '"+supervisor+"' нет данных", http.StatusNotFound)
		return
	}

	log.Printf("Найдено %d тем", len(topics))

	// Создаем Excel файл
	f := excelize.NewFile()

	// Устанавливаем заголовки
	f.SetCellValue("Sheet1", "A1", "Название темы")
	f.SetCellValue("Sheet1", "B1", "Предмет")
	f.SetCellValue("Sheet1", "C1", "Тип работы")
	f.SetCellValue("Sheet1", "D1", "Цикловая комиссия")
	f.SetCellValue("Sheet1", "E1", "Статус")
	f.SetCellValue("Sheet1", "F1", "Группа")
	f.SetCellValue("Sheet1", "G1", "Описание")

	// Заполняем данные
	for i, topic := range topics {
		row := i + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), topic.Title)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), topic.Subject)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), topic.WorkType)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), topic.Commission)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), topic.Status)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), topic.Group)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), topic.Description)
	}

	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=supervisor_%s.xlsx", strings.ReplaceAll(supervisor, " ", "_")))
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// Записываем файл в ответ
	if err := f.Write(w); err != nil {
		log.Printf("Ошибка записи Excel: %v", err)
		http.Error(w, "Ошибка создания файла", http.StatusInternalServerError)
		return
	}

	log.Printf("Успешно отправлен Excel файл для руководителя %s", supervisor)
}

func showExportSupervisorForm(w http.ResponseWriter) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Выгрузка данных по руководителю</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 0;
            padding: 0;
            background: #2B2726;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container { 
            max-width: 500px; 
            width: 90%;
            background: #403541;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 8px 32px rgba(0,0,0,0.3);
        }
        h2 {
            color: #F0EC8B;
            text-align: center;
            margin-bottom: 30px;
            font-size: 28px;
        }
        .form-group { 
            margin-bottom: 25px; 
        }
        label { 
            display: block; 
            margin-bottom: 10px; 
            font-weight: bold;
            color: #F0EC8B;
            font-size: 16px;
        }
        input { 
            width: 100%; 
            padding: 14px; 
            border: 2px solid #403541; 
            border-radius: 8px;
            font-size: 16px;
            background: #2B2726;
            color: #F0EC8B;
            transition: all 0.3s ease;
        }
        input:focus {
            outline: none;
            border-color: #8E43ED;
            box-shadow: 0 0 0 3px rgba(142, 67, 237, 0.2);
        }
        input::placeholder {
            color: #F0EC8B;
            opacity: 0.6;
        }
        button { 
            background: #8E43ED; 
            color: #2B2726; 
            padding: 16px 24px; 
            border: none; 
            border-radius: 8px; 
            cursor: pointer; 
            font-size: 18px;
            font-weight: bold;
            width: 100%;
            transition: all 0.3s ease;
            margin-top: 10px;
        }
        button:hover { 
            background: #F0EC8B;
            color: #2B2726;
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(142, 67, 237, 0.4);
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Выгрузка данных по руководителю</h2>
        <form method="POST" action="/export-supervisor">
            <div class="form-group">
                <label for="supervisor">Введите ФИО руководителя:</label>
                <input type="text" id="supervisor" name="supervisor" required placeholder="Например: Иванов И.И.">
            </div>
            <button type="submit">Скачать Excel</button>
        </form>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Ошибка отправки формы: %v", err)
	}
}

func exportSupervisorFormHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "exportSupervisor.html", nil)
}
