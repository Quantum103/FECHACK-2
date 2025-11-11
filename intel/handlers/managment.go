package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"proj/intel/models"
	"proj/intel/services"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB

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

// Функция для рандомного распределения тем

func AutoAssignTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := services.GetDB()

	// Получаем студентов без тем
	var studentsWithoutTopics []models.User
	err := db.Where("role = ? AND (topic = '' OR topic IS NULL)", "student").Find(&studentsWithoutTopics).Error
	if err != nil {
		http.Error(w, "Ошибка получения студентов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем свободные темы
	var freeTopics []models.Topic
	err = db.Where("status = 'free' OR student_id IS NULL OR student_id = 0").Find(&freeTopics).Error
	if err != nil {
		http.Error(w, "Ошибка получения тем: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, что есть что распределять
	if len(studentsWithoutTopics) == 0 {
		http.Error(w, "Нет студентов без тем", http.StatusBadRequest)
		return
	}

	if len(freeTopics) == 0 {
		http.Error(w, "Нет свободных тем", http.StatusBadRequest)
		return
	}

	// Перемешиваем студентов и темы для случайного распределения
	shuffledStudents := shuffleStudents(studentsWithoutTopics)
	shuffledTopics := shuffleTopics(freeTopics)

	// Распределяем темы (берем минимум из количества студентов и тем)
	count := min(len(shuffledStudents), len(shuffledTopics))
	assignedCount := 0

	for i := 0; i < count; i++ {
		student := shuffledStudents[i]
		topic := shuffledTopics[i]

		// Назначаем тему студенту
		err := db.Model(&models.Topic{}).
			Where("id = ?", topic.ID).
			Updates(map[string]interface{}{
				"student_id": student.ID,
				"status":     "assigned",
			}).Error

		if err != nil {
			log.Printf("Ошибка назначения темы %d студенту %d: %v", topic.ID, student.ID, err)
			continue
		}

		// Обновляем тему у студента
		err = db.Model(&models.User{}).
			Where("id = ?", student.ID).
			Update("topic", topic.Title).Error

		if err != nil {
			log.Printf("Ошибка обновления студента %d: %v", student.ID, err)
			continue
		}

		assignedCount++
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"message":        fmt.Sprintf("Успешно распределено %d тем из %d возможных", assignedCount, count),
		"assigned":       assignedCount,
		"total_possible": count,
	})
}

// Функции для перемешивания
func shuffleStudents(students []models.User) []models.User {
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]models.User, len(students))
	copy(shuffled, students)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func shuffleTopics(topics []models.Topic) []models.Topic {
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]models.Topic, len(topics))
	copy(shuffled, topics)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// удаляет назначенные темы
func RemoveAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	studentID := r.FormValue("student_id")
	topicID := r.FormValue("topic_id")

	if studentID == "" || topicID == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	db := services.GetDB()

	// Освобождаем тему
	err := db.Model(&models.Topic{}).
		Where("id = ?", topicID).
		Updates(map[string]interface{}{
			"student_id": nil,
			"status":     "free",
		}).Error

	if err != nil {
		http.Error(w, "Ошибка освобождения темы", http.StatusInternalServerError)
		return
	}

	// Очищаем тему у студента
	err = db.Model(&models.User{}).
		Where("id = ?", studentID).
		Update("topic", "").Error

	if err != nil {
		http.Error(w, "Ошибка обновления студента", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Назначение удалено",
	})
}

// Переназначение темы
