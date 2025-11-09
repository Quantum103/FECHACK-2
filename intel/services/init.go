package services

import (
	"fmt"
	"log"
	"proj/intel/models"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB() error {
	var err error
	once.Do(func() {
		dbPath := "./attendance.db"
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			log.Fatal("Ошибка подключения к БД:", err)
			return
		}

		// Автомиграция
		err = db.AutoMigrate(&models.User{}, &models.Attendance{}, &models.Groupfromcur{}, &models.ChatMessage{}, &models.Topic{})
		if err != nil {
			log.Fatal("Ошибка миграции:", err)
			return
		}

		fmt.Println("✅ База данных SQLite создана/подключена")
		createDefaultAdmin()
	})
	return err
}

func createDefaultAdmin() {
	admin := models.User{
		Name:     "Администратор",
		Email:    "admin@system.com",
		Password: "admin123",
		Role:     "admin",
	}

	db.FirstOrCreate(&admin, models.User{Email: "admin@system.com"})
}

func GetDB() *gorm.DB {
	if db == nil {
		log.Panic("Database not initialized! Call InitDB() first")
	}
	return db
}
