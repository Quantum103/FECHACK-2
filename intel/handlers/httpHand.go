// маршруты и их функции (переходят в файл indexTemp.go)
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"proj/intel/models"
	"proj/intel/services"
	"proj/middleware"
	"proj/utils"
	"strconv"
)

// маршруты и их функции (переходят в файл indexTemp.go)

func RegisterRouter() {
	hub := NewHub()
	go hub.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// ──────  публичные (без авторизации)  ──────
	http.HandleFunc("/register/", register)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)

	// ──────  защищённые  ──────
	http.Handle("/dashboard/", middleware.CheckAuth(Dashboard))
	http.Handle("/", middleware.CheckAuth(Dashboard)) // главная = админ‑панель
	http.Handle("/students/", middleware.CheckAuth(ListStudents))
	http.Handle("/ruc/", middleware.CheckAuth(ruc))
	http.Handle("/starosta/", middleware.CheckAuth(starosta))
	http.Handle("/upload", http.HandlerFunc(AdminFunction))

	http.Handle("/api/users", middleware.CheckAuth(usersList))
	http.Handle("/api/messages/", middleware.CheckAuth(messageHistory))

	log.Printf("Server started, listening on %s", os.Getenv("ADDR"))

}

func usersList(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := services.GetDB().Find(&users).Error; err != nil {
		http.Error(w, "Ошибка получения пользователей", http.StatusInternalServerError)
		return
	}

	// Преобразуем в JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func messageHistory(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя
	_, err := utils.GetUserFromCookie(r)
	if err != nil {
		http.Error(w, "Неавторизован", http.StatusUnauthorized)
		return
	}
	userID, err := utils.GetUserIDFromCookie(r)

	// Получаем ID собеседника
	companionIDStr := r.URL.Query().Get("companion_id")
	companionID, err := strconv.ParseUint(companionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Загружаем сообщения между пользователями
	var messages []models.ChatMessage
	if err := services.GetDB().Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, companionID, companionID, userID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		http.Error(w, "Ошибка загрузки сообщений", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
