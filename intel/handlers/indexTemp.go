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
	_, err := utils.GetUserFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	templates.ExecuteTemplate(w, "listStudents.html", nil)
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

// func HeadmanFunction(w http.ResponseWriter, r *http.Request) {
// 	_, err := utils.GetUserFromCookie(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}

// 	db := services.GetDB()
// 	users, err := GetUsersByGroup(db, "ИС-202")
// 	if err != nil {
// 		http.Error(w, "Ошибка получения пользователей", http.StatusInternalServerError)
// 		return
// 	}

// 	// Форматируем сегодняшнюю дату
// 	today := time.Now().Format("02.01.2006")

// 	data := TemplateData{
// 		Users: users,
// 		Date:  today,
// 	}

// 	templates.ExecuteTemplate(w, "listUsers.html", data)
// }

// func StudentFunction(w http.ResponseWriter, r *http.Request) {
// 	claims, err := utils.GetUserFromCookie(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}

// 	// Получаем информацию о студенте
// 	db := services.GetDB()
// 	var student models.User
// 	if err := db.Where("email = ?", claims.Email).First(&student).Error; err != nil {
// 		http.Error(w, "Студент не найден", http.StatusInternalServerError)
// 		return
// 	}

// 	// Получаем посещаемость за сегодня
// 	today := time.Now().Format("2006-01-02")
// 	var attendance models.Attendance
// 	result := db.Where("user_id = ? AND DATE(date) = ?", student.ID, today).First(&attendance)

// 	// Подготавливаем данные для шаблона
// 	data := struct {
// 		Student     models.User
// 		Attendance  *models.Attendance
// 		HasRecord   bool
// 		Date        string
// 		Status      string
// 		StatusClass string
// 	}{
// 		Student:   student,
// 		Date:      time.Now().Format("02.01.2006"),
// 		HasRecord: result.Error == nil,
// 	}

// 	if data.HasRecord {
// 		data.Attendance = &attendance
// 		if attendance.HoursMissed == 0 {
// 			data.Status = "Присутствовал"
// 			data.StatusClass = "present"
// 		} else {
// 			data.Status = "Отсутствовал"
// 			data.StatusClass = "absent"
// 		}
// 	} else {
// 		data.Status = "Не отмечен"
// 		data.StatusClass = "unknown"
// 	}

// 	templates.ExecuteTemplate(w, "student_dashboard.html", data)
// }

// func CuratorFunction(w http.ResponseWriter, r *http.Request) {
// 	claims, err := utils.GetUserFromCookie(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}

// 	db := services.GetDB()

// 	// Получаем текущего куратора по email из куки (так как Name нет в Claims)
// 	var curator models.User
// 	if err := db.Where("email = ?", claims.Email).First(&curator).Error; err != nil {
// 		http.Error(w, "Куратор не найден", http.StatusInternalServerError)
// 		return
// 	}

// 	// Обработка добавления новой группы
// 	if r.Method == http.MethodPost {
// 		groupName := r.FormValue("group")

// 		if groupName != "" {
// 			// Проверяем, нет ли уже такой группы у этого куратора
// 			var existingGroup models.Groupfromcur
// 			result := db.Where("name = ? AND group = ?", curator.Name, groupName).First(&existingGroup)

// 			if result.Error != nil {
// 				// Если группы нет - создаем
// 				curatorGroup := &models.Groupfromcur{
// 					Name:  curator.Name, // Используем ФИО из БД
// 					Group: groupName,
// 				}
// 				services.Add(curatorGroup)

// 			}
// 		}
// 		http.Redirect(w, r, "/dashboard/", http.StatusFound)
// 		return
// 	}

// 	// Получаем ТОЛЬКО группы текущего куратора (по ФИО из БД)
// 	var curatorGroups []models.Groupfromcur
// 	db.Where("name = ?", curator.Name).Find(&curatorGroups)

// 	data := struct {
// 		Curator models.User
// 		Groups  []models.Groupfromcur
// 	}{
// 		Curator: curator,
// 		Groups:  curatorGroups,
// 	}

// 	templates.ExecuteTemplate(w, "curator_dashboard.html", data)
// }

// // func CuratorFunction(w http.ResponseWriter, r *http.Request) {

// // 	if r.Method == http.MethodPost {
// // 		name := r.FormValue("full_name")
// // 		group := r.FormValue("group")

// // 		curatorGroup := &models.Groupfromcur{
// // 			Name:  name,
// // 			Group: group,
// // 		}
// // 		services.Add(curatorGroup)
// // 		http.Redirect(w, r, "/dashboard/", http.StatusFound)
// // 		return
// // 	}

// // 	claims, err := utils.GetUserFromCookie(r)
// // 	if err != nil {
// // 		http.Redirect(w, r, "/login", http.StatusFound)
// // 		return
// // 	}

// // 	// Получаем пользователя чтобы узнать его ФИО
// // 	var user models.User
// // 	db := services.GetDB()
// // 	db.Where("email = ?", claims.Email).First(&user)
// // 	var myGroups []models.Groupfromcur
// // 	db.Where("name = ?", user.Name).Find(&myGroups)

// // 	templates.ExecuteTemplate(w, "curator_dashboard.html", myGroups)
// // }

// func AdminFunction(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodPost {
// 		name := r.FormValue("full_name")
// 		number := r.FormValue("email")
// 		group := r.FormValue("group")
// 		telegramm := r.FormValue("telegramm")
// 		password := r.FormValue("password")
// 		email := r.FormValue("email")
// 		role := "curator"
// 		// делаем шаблон пользователя
// 		user := &models.User{
// 			Name:     name,
// 			Phone:    number,
// 			Group:    group,
// 			Telegram: telegramm,
// 			Email:    email,
// 			Password: password,
// 			Role:     role,
// 		}
// 		// добавляем пользователя в БД
// 		services.Add(user)
// 		http.Redirect(w, r, "/dashboard/", http.StatusFound)
// 	}

// 	templates.ExecuteTemplate(w, "admin_dashboard.html", nil)
// }

// // Пример использования дополнительных функций
// func Profile(w http.ResponseWriter, r *http.Request) {
// 	// Получаем отдельно роль
// 	role, err := utils.GetUserRoleFromCookie(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}

// 	// Получаем ID пользователя
// 	userID, err := utils.GetUserIDFromCookie(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}

// 	// Получаем email
// 	email, err := utils.GetUserEmailFromCookie(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}

// 	data := struct {
// 		Role  string
// 		ID    uint
// 		Email string
// 	}{
// 		Role:  role,
// 		ID:    userID,
// 		Email: email,
// 	}

// 	templates.ExecuteTemplate(w, "profile.html", data)
// }
