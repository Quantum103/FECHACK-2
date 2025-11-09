package services

import (
	"fmt"
	"proj/intel/models"
)

func Add(user interface{}) {
	res := db.Create(user)
	if res.Error != nil {
		panic(res)
	}
	fmt.Println("операция прошла успешно!")
}

func SaveMessage(message *models.ChatMessage) {
	res := db.Create(message)
	if res.Error != nil {
		panic(res)
	}
	fmt.Println("сообщение сохранено!")
}
