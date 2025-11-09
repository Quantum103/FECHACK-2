// package middleware

// import (
// 	"net/http"
// 	"proj/utils"
// )

// func CheckAuth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		// Получаем токен из заголовка
// 		tokenString, _ := utils.GetUserFromCookie(r)
// 		if tokenString == nil {
// 			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
// 			return
// 		}

// 		if tokenString["user_id"] == nil || tokenString["email"] == nil {
// 			http.Error(w, "Invalid token content", http.StatusUnauthorized)
// 			return
// 		}
// 		endpoint(w, r)
// 	})
// }

package middleware

import (
	"net/http"
	"proj/utils"
)

// CheckAuth - основной middleware для проверки аутентификации
func CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := utils.GetUserFromCookie(r)
		if err != nil {
			// Если нет токена, перенаправляем на страницу логина
			http.Redirect(w, r, "/login/", http.StatusFound)
			return
		}

		// Проверяем, что в токене есть необходимые данные
		// Теперь claims - это *utils.Claims, а не jwt.MapClaims
		if claims.UserID == 0 || claims.Email == "" || claims.Role == "" {
			http.Redirect(w, r, "/login/", http.StatusFound)
			return
		}

		// Если пользователь аутентифицирован, передаем управление следующему обработчику
		next(w, r)
	}
}

// AdminOnly - middleware для проверки прав администратора
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := utils.GetUserFromCookie(r)
		if err != nil || claims.Role != "admin" {
			http.Error(w, "Доступ запрещен. Требуются права администратора.", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// CuratorOnly - middleware для проверки прав куратора
func CuratorOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := utils.GetUserFromCookie(r)
		if err != nil || (claims.Role != "curator" && claims.Role != "admin") {
			http.Error(w, "Доступ запрещен. Требуются права куратора или администратора.", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// HeadmanOrAbove - middleware для проверки прав старосты и выше
func HeadmanOrAbove(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := utils.GetUserFromCookie(r)
		if err != nil {
			http.Redirect(w, r, "/login/", http.StatusFound)
			return
		}

		if claims.Role != "headman" && claims.Role != "curator" && claims.Role != "admin" {
			http.Error(w, "Доступ запрещен. Требуются права старосты или выше.", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
