// файл для работы со всеми шаблонами
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proj/intel/models"
	"proj/intel/services"
	"proj/utils"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	claims, err := utils.GetUserFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	switch claims.Role {
	case "admin":
		AdminFunction(w, r)
	case "curator":
		// CuratorFunction(w, r)
	case "student":
		// StudentFunction(w, r)
	case "headman":
		// HeadmanFunction(w, r)
	default:
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func AdminFunction(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== UPLOAD HANDLER CALLED ===")
	log.Printf("Method: %s", r.Method)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))

	// Добавьте CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		log.Printf("OPTIONS request handled")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Проверяем авторизацию
	_, err := utils.GetUserFromCookie(r)
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		log.Printf("Starting file upload processing")

		// Получаем файл
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error getting file: %v", err)
			sendError(w, "No file uploaded: "+err.Error())
			return
		}
		defer file.Close()

		log.Printf("File received: %s, Size: %d", header.Filename, header.Size)

		fileType := r.FormValue("type")
		if fileType == "" {
			fileType = "students"
		}
		log.Printf("File type: %s", fileType)

		// Обрабатываем Excel файл
		result := processExcelFile(file, fileType)
		log.Printf("Processing result: %+v", result)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}

		log.Printf("Upload completed successfully")
		return
	}

	log.Printf("GET request - serving admin page")
	templates.ExecuteTemplate(w, "admin.html", nil)
}

func ListStudents(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	_, err := utils.GetUserFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Получаем реальные данные из базы
	data, err := GetAllTopicsData()
	if err != nil {
		log.Printf("Ошибка получения данных: %v", err)
		http.Error(w, "Ошибка загрузки данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Добавляем заголовок для страницы
	data["Title"] = "Управление студентами и темами"

	// Выполн+яем шаблон
	err = templates.ExecuteTemplate(w, "listStudents.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

func ruc(w http.ResponseWriter, r *http.Request) {
	_, err := utils.GetUserFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if r.Method == http.MethodPost {
		studentEmail := r.FormValue("student_email")
		headmanGroup := r.FormValue("headman_group")

		fmt.Printf("Получены данные: email=%s, group=%s\n", studentEmail, headmanGroup)

		db := services.GetDB()

		// Находим студента по email (УБЕРИ проверку роли)
		var student models.User
		result := db.Where("email = ?", studentEmail).First(&student)

		if result.Error != nil {
			fmt.Printf("Ошибка поиска студента: %v\n", result.Error)
			http.Error(w, "Студент не найден", http.StatusBadRequest)
			return
		}

		fmt.Printf("Найден студент: %s, текущая роль: %s\n", student.Name, student.Role)

		// Меняем роль на headman и назначаем группу
		student.Role = "curator"

		result = db.Save(&student)
		if result.Error != nil {
			fmt.Printf("Ошибка сохранения: %v\n", result.Error)
			http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Роль успешно изменена на: %s\n", student.Role)
		http.Redirect(w, r, "/dashboard/", http.StatusFound)
		return
	}

	templates.ExecuteTemplate(w, "rucCreate.html", nil)
}

func starosta(w http.ResponseWriter, r *http.Request) {
	_, err := utils.GetUserFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	templates.ExecuteTemplate(w, "starosta.html", nil)

}
