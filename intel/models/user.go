package models

import (
	"time"

	"gorm.io/gorm"
)

type Topic struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title"`       // Название темы
	Subject     string `json:"subject"`     // Предмет
	WorkType    string `json:"workType"`    // Вид работы: "course" или "diploma"
	Commission  string `json:"commission"`  // Цикловая комиссия
	Supervisor  string `json:"supervisor"`  // Руководитель
	Description string `json:"description"` // Описание темы (опционально)
	Status      string `json:"status"`      // Статус: "free" или "assigned"
	StudentID   uint   `json:"studentId"`   // ID студента, если назначена
	Group       string `json:"group"`       // Группа, для которой предназначена тема
}

type User struct {
	gorm.Model
	Name         string `gorm:"size:50" json:"full_name"`
	Email        string `gorm:"uniqueIndex" json:"email"`
	Password     string `gorm:"password" json:"-"`
	Role         string `gorm:"size:20;default:student" json:"role"` // admin, curator, headman, student
	Group        string `gorm:"size:20" json:"group"`
	Topic        string `gorm:"size:100" json:"topic"`
	HeadmanGroup string `gorm:"size:20" json:"headman_group"` // Группа, за которую отвечает староста

}

type Groupfromcur struct {
	gorm.Model
	Name  string `gorm:"size:50" json:"full_name"`
	Group string `gorm:"size:20" json:"group"`
}

type ChatMessage struct {
	ID         uint   `gorm:"primaryKey"`
	SenderID   uint   `gorm:"index;not null"`
	ReceiverID uint   `gorm:"index;not null"`
	Content    string `gorm:"type:text;not null"`
	CreatedAt  time.Time
}
