// package utils

// import (
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/golang-jwt/jwt"
// )

// var (
// 	CookieName = "jwt_token"
// 	SecretKey  = "your-secret-key"
// )

// func SetJWTCookie(w http.ResponseWriter, userID uint, email string) error {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"user_id": userID,
// 		"email":   email,
// 		"exp":     time.Now().Add(time.Hour * 24).Unix(),
// 	})
// 	tokenString, err := token.SignedString([]byte(SecretKey))
// 	if err != nil {
// 		return err
// 	}
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     CookieName,
// 		Value:    tokenString,
// 		Expires:  time.Now().Add(24 * time.Hour),
// 		HttpOnly: true,
// 		Secure:   false, // true в production
// 		Path:     "/",
// 		SameSite: http.SameSiteStrictMode,
// 	})

// 	return nil

// }

// func GetUserFromCookie(r *http.Request) (jwt.MapClaims, error) {
// 	cookie, err := r.Cookie(CookieName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
// 		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
// 		}
// 		return []byte(SecretKey), nil
// 	})

// 	if err != nil || !token.Valid {
// 		return nil, err
// 	}

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok {
// 		if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
// 			return nil, errors.New("token expired")
// 		}
// 		return claims, nil
// 	}

// 	return nil, errors.New("invalid token claims")
// }

// func ClearJWTCookie(w http.ResponseWriter) {
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     CookieName,
// 		Value:    "",
// 		Expires:  time.Now().Add(-time.Hour), // Устанавливаем время в прошлом
// 		HttpOnly: true,
// 		Secure:   false,
// 		Path:     "/",
// 	})
// }

package utils

import (
	"net/http"
	"time"
)

var CookieName = "jwt_token"

func SetJWTCookie(w http.ResponseWriter, userID uint, email, role string) error {
	tokenString, err := GenerateJWT(userID, email, role)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // true в production
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

func GetUserFromCookie(r *http.Request) (*Claims, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return nil, err
	}

	claims, err := ParseJWT(cookie.Value)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func ClearJWTCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
}

// Дополнительные функции для удобства
func GetUserRoleFromCookie(r *http.Request) (string, error) {
	claims, err := GetUserFromCookie(r)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

func GetUserIDFromCookie(r *http.Request) (uint, error) {
	claims, err := GetUserFromCookie(r)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func GetUserEmailFromCookie(r *http.Request) (string, error) {
	claims, err := GetUserFromCookie(r)
	if err != nil {
		return "", err
	}
	return claims.Email, nil
}
