package services

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tieubaoca/chatbot-be/types"
)

type WebSocketService struct {
	ai       *OpenAIService
	upgrader websocket.Upgrader
}

func NewWebSocketService(ai *OpenAIService) *WebSocketService {
	return &WebSocketService{
		ai: ai,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins (adjust for production)
			},
		},
	}
}

func (s *WebSocketService) HandleChat(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	// Set connection properties
	conn.SetReadLimit(512 * 1024) // 512KB max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	// Create done channel for graceful shutdown
	done := make(chan struct{})
	defer close(done)

	// Start ping ticker in separate goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("ping error: %v", err)
					return
				}
			}
		}
	}()

	for {
		// Read message from client
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			continue
		}
		var req types.WebsocketRequest
		if err := json.Unmarshal(p, &req); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}
		payloadBytes, err := json.Marshal(req.Payload)
		if err != nil {
			log.Println("Marshal error:", err)
			continue
		}
		switch req.Type {
		case types.TypeWebsocketChat:
			{
				// Process message with AI
				var payload types.WebSocketChatPayload
				err := json.Unmarshal(payloadBytes, &payload)

				if err != nil {
					log.Println("Unmarshal error:", err)
					continue
				}
				// Stream AI responses back to client
				res, err := s.ai.Chat(r.Context(), payload.Messages)
				if err != nil {
					log.Println("AI error:", err)
					conn.WriteMessage(messageType, []byte("Error processing message"))
					continue
				}
				botMessage := types.WebSocketResponse{
					Type:    types.TypeWebsocketChat,
					Payload: types.WebSocketChatResponse{Message: res.Content},
				}
				if err := conn.WriteJSON(botMessage); err != nil {
					log.Println("Write error:", err)
					continue
				}

			}
		case types.TypeWebsocketPing:
			{
				// Send a pong message back to the client
				pongRes := types.WebSocketResponse{
					Type:    types.TypeWebsocketPong,
					Payload: nil,
				}
				if err := conn.WriteJSON(pongRes); err != nil {
					log.Println("Write error:", err)
				}
				continue
			}
		default:
			{
				log.Println("Invalid message type")
				continue
			}
		}

	}
}

func (s *WebSocketService) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
