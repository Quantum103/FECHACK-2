// маршруты и их функции (переходят в файл indexTemp.go)
package handlers

import (
	"log"
	"net/http"
	"os"
	"proj/middleware"
)

// маршруты и их функции (переходят в файл indexTemp.go)

func RegisterRouter() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// ──────  публичные (без авторизации)  ──────
	http.HandleFunc("/register/", register)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)

	http.HandleFunc("/assign-topic", AssignTopicToStudent)
	http.HandleFunc("/assign-student", AssignStudentToTopic)

	http.HandleFunc("/export-list", exportList)
	http.HandleFunc("/export", exportHandler)
	http.HandleFunc("/export-form", exportFormHandler)
	http.HandleFunc("/export-supervisor", exportSupervisorHandler)
	http.HandleFunc("/export-supervisor-form", exportSupervisorFormHandler)

	http.HandleFunc("/student", StudentFunction)

	// ──────  защищённые  ──────
	http.Handle("/dashboard/", middleware.CheckAuth(Dashboard))
	http.Handle("/", middleware.CheckAuth(Dashboard)) // главная = админ‑панель
	http.Handle("/students", middleware.RecoveryMiddleware(ListStudents))
	http.Handle("/ruc/", middleware.CheckAuth(ruc))
	// ✅ ИСПРАВЛЕННЫЕ маршруты для старост:
	http.Handle("/addStarosta/", middleware.CheckAuth(http.HandlerFunc(addStatosta))) // Со слешем

	http.Handle("/admin-upload", http.HandlerFunc(AdminFunction)) // Уникальный путь

	http.Handle("/auto-assign", middleware.CheckAuth(http.HandlerFunc(AutoAssignTopics)))
	http.Handle("/remove-assignment", middleware.CheckAuth(http.HandlerFunc(RemoveAssignment)))
	http.Handle("/studentsStar", middleware.CheckAuth(http.HandlerFunc(StudentsForStarosta)))

	log.Printf("Server started, listening on %s", os.Getenv("ADDR"))
}
