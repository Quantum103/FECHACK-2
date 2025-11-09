// файл для работы со всеми шаблонами
package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"proj/intel/models"

	"proj/intel/services"

	"proj/utils"
)

var (
	templates *template.Template
)

func LoadTemplates() {
	// Получаем абсолютный путь к папке с шаблонами
	wd, _ := os.Getwd()
	templateDir := filepath.Join(wd, "intel", "handlers", "templates")

	// Загружаем все шаблоны из папки
	templates = template.Must(template.ParseGlob(filepath.Join(templateDir, "*.html")))
}
func logout(w http.ResponseWriter, r *http.Request) {
	utils.ClearJWTCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard/", http.StatusSeeOther)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// получаем данные из формы регистрации
		name := r.FormValue("full_name")
		group := r.FormValue("group")
		password := r.FormValue("password")
		email := r.FormValue("email")
		// делаем шаблон пользователя
		user := &models.User{
			Name:     name,
			Group:    group,
			Email:    email,
			Password: password,
		}
		// добавляем пользователя в БД
		services.Add(user)

		if err := utils.SetJWTCookie(w, user.ID, user.Email, user.Role); err != nil {
			http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard/", http.StatusSeeOther)

	}

	// загрузка шаблонов
	templates.ExecuteTemplate(w, "Register.html", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	var db = services.GetDB()
	if r.Method == http.MethodPost {

		email := r.FormValue("email")
		password := r.FormValue("password")

		var user models.User
		result := db.Where("email=? and password=?", email, password).First(&user)
		if result.Error == nil && user.ID != 0 {
			// Теперь передаем роль в JWT
			utils.SetJWTCookie(w, user.ID, user.Email, user.Role)
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
	}

	templates.ExecuteTemplate(w, "Login.html", nil)
}
