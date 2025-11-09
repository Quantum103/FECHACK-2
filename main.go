package main

import (
	"fmt"
	"log"
	"net/http"
	"proj/intel/handlers"
	"proj/intel/services"
)

func main() {
	err := services.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	handlers.LoadTemplates()
	handlers.RegisterRouter()
	fmt.Println("Сервер запустился на :8080")
	http.ListenAndServe(":8080", nil)
}
