package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"proj/intel/models"
	"proj/intel/services"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB

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

type StudentWithTopic struct {
	User  models.User  `json:"user"`
	Topic models.Topic `json:"topic"`
}

// Получение всех данных для шаблона
func GetAllTopicsData() (map[string]interface{}, error) {
	studentsWithTopics, err := GetStudentsWithTopics()
	if err != nil {
		return nil, err
	}

	studentsWithoutTopics, err := GetStudentsWithoutTopics()
	if err != nil {
		return nil, err
	}

	freeTopics, err := GetFreeTopics()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"StudentsWithTopics":    studentsWithTopics,
		"StudentsWithoutTopics": studentsWithoutTopics,
		"FreeTopics":            freeTopics,
	}, nil
}

func GetStudentsWithTopics() ([]StudentWithTopic, error) {
	db := services.GetDB()
	var studentsWithTopics []StudentWithTopic

	var users []models.User
	err := db.Where("role = ? AND id IN (SELECT student_id FROM topics WHERE student_id IS NOT NULL)", "student").
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		var topic models.Topic
		err := db.Where("student_id = ?", user.ID).First(&topic).Error
		if err == nil {
			studentsWithTopics = append(studentsWithTopics, StudentWithTopic{
				User:  user,
				Topic: topic,
			})
		}
	}

	return studentsWithTopics, nil
}

// Студенты без тем
func GetStudentsWithoutTopics() ([]models.User, error) {
	var students []models.User
	db := services.GetDB()
	err := db.Where("role = ? AND id NOT IN (SELECT student_id FROM topics WHERE student_id IS NOT NULL)", "student").
		Find(&students).Error

	return students, err
}

// Свободные темы
func GetFreeTopics() ([]models.Topic, error) {
	var topics []models.Topic
	db := services.GetDB()
	err := db.Where("status = ? OR student_id IS NULL OR student_id = 0", "free").Find(&topics).Error
	return topics, err
}

func AssignTopicToStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	studentID := r.FormValue("student_id")
	topicID := r.FormValue("topic_id")

	// Преобразуем ID в uint
	studentIDUint, err := strconv.ParseUint(studentID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	topicIDUint, err := strconv.ParseUint(topicID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	db := services.GetDB()

	// Обновляем тему - назначаем студента
	result := db.Model(&models.Topic{}).
		Where("id = ?", topicIDUint).
		Updates(map[string]interface{}{
			"student_id": uint(studentIDUint),
			"status":     "assigned",
		})

	if result.Error != nil {
		http.Error(w, "Failed to assign topic", http.StatusInternalServerError)
		return
	}

	// Обновляем студента - записываем тему
	result = db.Model(&models.User{}).
		Where("id = ?", studentIDUint).
		Update("topic", func() string {
			var topic models.Topic
			db.First(&topic, topicIDUint)
			return topic.Title
		}())

	if result.Error != nil {
		http.Error(w, "Failed to update student", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// Обработчик для назначения студента теме
func AssignStudentToTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicID := r.FormValue("topic_id")
	studentID := r.FormValue("student_id")

	// Создаем новый запрос с правильными параметрами
	r.ParseForm()
	r.Form.Set("student_id", studentID)
	r.Form.Set("topic_id", topicID)

	// Теперь вызываем обработчик с правильными данными
	AssignTopicToStudent(w, r)
}
