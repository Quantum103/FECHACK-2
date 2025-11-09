package handlers

import (
	"bytes"
	"html/template"
	"log"
	"sync"
)

type Message struct {
	SenderID    uint   // ID отправителя
	SenderName  string // Имя отправителя
	RecipientID uint   // ID получателя (0 = всем)
	Text        string // Текст сообщения
}

type WSMessage struct {
	Text    string      `json:"text"`
	Headers interface{} `json:"HEADERS"`
}

type Hub struct {
	sync.RWMutex
	clients    map[*Client]bool
	users      map[uint]*Client
	brodcast   chan *Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		users:      make(map[uint]*Client), // Добавьте эту строку
		brodcast:   make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client] = true
			h.users[client.userID] = client // Сохраняем клиента по ID пользователя
			h.Unlock()

			log.Printf("client registered %d", client.userID)

		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.users, client.userID) // Удаляем из users
				close(client.send)
				log.Printf("client unregistered %d", client.userID)
			}
			h.Unlock()

		case msg := <-h.brodcast:
			h.RLock()
			if msg.RecipientID != 0 {
				// Отправляем сообщение обоим участникам чата
				if sender, exists := h.users[msg.SenderID]; exists {
					select {
					case sender.send <- getMessageTemplate(msg):
					default:
						close(sender.send)
						delete(h.clients, sender)
						delete(h.users, sender.userID)
					}
				}
				if recipient, exists := h.users[msg.RecipientID]; exists {
					select {
					case recipient.send <- getMessageTemplate(msg):
					default:
						close(recipient.send)
						delete(h.clients, recipient)
						delete(h.users, recipient.userID)
					}
				}
			} else {
				// Broadcast всем
				for client := range h.clients {
					select {
					case client.send <- getMessageTemplate(msg):
					default:
						close(client.send)
						delete(h.clients, client)
						delete(h.users, client.userID)
					}
				}
			}
			h.RUnlock()
		}
	}
}

func getMessageTemplate(msg *Message) []byte {
	tmpl, err := template.ParseFiles("intel/handlers/templates/message.html")
	if err != nil {
		log.Printf("template parsing: %s", err)
		return []byte("")
	}

	var renderedMessage bytes.Buffer
	err = tmpl.Execute(&renderedMessage, msg)
	if err != nil {
		log.Printf("template execution: %s", err)
		return []byte("")
	}

	return renderedMessage.Bytes()
}
