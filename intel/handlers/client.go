package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"proj/intel/models"
	"proj/intel/services"
	"proj/utils"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	userID uint // ID пользователя из БД
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

const (
	pongWait       = 60 * time.Second
	maxMessageSize = 512
	pingPeriod     = (pongWait * 9) / 10
	writeWait      = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	userID, err := utils.GetUserIDFromCookie(r)

	client := &Client{
		userID: userID, // Сохраняем ID пользователя
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte),
	}

	client.hub.register <- client

	go client.writePUMP()
	go client.readPUMP()
}

func (c *Client) readPUMP() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, text, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("Received: %s", string(text))

		// Парсим JSON сообщение
		var data map[string]interface{}
		if err := json.Unmarshal(text, &data); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		// Извлекаем текст сообщения
		textValue, ok := data["text"].(string)
		if !ok {
			log.Printf("Invalid message format: text field missing or not a string")
			continue
		}

		// Извлекаем recipient_id
		var recipientID uint
		if recipient, exists := data["recipient_id"]; exists {
			switch v := recipient.(type) {
			case float64:
				recipientID = uint(v)
			case string:
				if v != "" {
					if id, err := strconv.ParseUint(v, 10, 64); err == nil {
						recipientID = uint(id)
					}
				}
			default:
				recipientID = 0
			}
		}

		// Получаем имя отправителя из базы данных
		var user models.User
		if err := services.GetDB().First(&user, c.userID).Error; err != nil {
			log.Printf("Error getting user: %v", err)
			continue
		}
		// Сохраняем сообщение в БД для отправителя
		chatMessageSent := models.ChatMessage{
			SenderID:   c.userID,
			ReceiverID: recipientID,
			Content:    textValue,
			CreatedAt:  time.Now(),
		}
		services.SaveMessage(&chatMessageSent)

		c.hub.brodcast <- &Message{
			SenderID:    c.userID,
			SenderName:  user.Name,
			RecipientID: recipientID,
			Text:        textValue,
		}
	}
}

func (c *Client) writePUMP() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(msg)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(msg)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
