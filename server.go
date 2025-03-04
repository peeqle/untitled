package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

const (
	UserNotification string = "/user/{username}/notifications"
	UserChat                = "/user/{username}"
	UserStatus              = "/user/{username}/status"
	UserActivity            = "/user/{username}/activity"

	GeneralChat        = "/chat/general"
	NotificationSystem = "/notifications/system"
	ChatGroup          = "/chat/group/{groupName}"
	ChatGroupAdmin     = "/chat/group/{groupName}/admin"

	EventsServer       = "/events/server"
	EventsErrorsServer = "/events/error"
)

type User struct {
	ID   string
	Conn *websocket.Conn
}

type Server struct {
	users      map[string]*User
	topics     map[string]map[string]bool
	userMutex  sync.RWMutex
	topicMutex sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		users:  make(map[string]*User),
		topics: make(map[string]map[string]bool),
	}
}

func (s *Server) AddUser(userID string, conn *websocket.Conn) {
	s.userMutex.Lock()
	defer s.userMutex.Unlock()
	s.users[userID] = &User{ID: userID, Conn: conn}
}

func (s *Server) RemoveUser(userID string) {
	s.userMutex.Lock()
	defer s.userMutex.Unlock()
	delete(s.users, userID)
}

func (s *Server) SubscribeToTopic(userID, topic string) {
	s.topicMutex.Lock()
	defer s.topicMutex.Unlock()

	if s.topics[topic] == nil {
		s.topics[topic] = make(map[string]bool)
	}
	s.topics[topic][userID] = true
}

func (s *Server) UnsubscribeFromTopic(userID, topic string) {
	s.topicMutex.Lock()
	defer s.topicMutex.Unlock()

	if s.topics[topic] != nil {
		delete(s.topics[topic], userID)
	}
}

func (s *Server) BroadcastToTopic(topic string, message []byte) {
	s.topicMutex.RLock()
	defer s.topicMutex.RUnlock()

	for userID := range s.topics[topic] {
		s.userMutex.RLock()
		if user, ok := s.users[userID]; ok {
			if err := user.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to send message to user %s: %v", userID, err)
			}
		}
		s.userMutex.RUnlock()
	}
}

func (s *Server) SendDirectMessage(userID string, message []byte) error {
	s.userMutex.RLock()
	defer s.userMutex.RUnlock()

	if user, ok := s.users[userID]; ok {
		if err := user.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send message to user %s: %v", userID, err)
			return err
		}
	}
	return nil
}

func (s *Server) HandleWebSocket(conn *websocket.Conn, c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		log.Println("User ID is required")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "User ID is required"))
		return
	}

	s.AddUser(userID, conn)
	defer s.RemoveUser(userID)

	log.Printf("User %s connected", userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("User %s disconnected unexpectedly: %v", userID, err)
			} else {
				log.Printf("User %s disconnected", userID)
			}
			break
		}

		log.Printf("Received message from user %s: %s", userID, message)

		var msg struct {
			Type    string `json:"type"`
			Topic   string `json:"topic,omitempty"`
			To      string `json:"to,omitempty"`
			Message string `json:"message,omitempty"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse message from user %s: %v", userID, err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid message format"}`))
			continue
		}

		// Обработка сообщения в зависимости от типа
		switch msg.Type {
		case "subscribe":
			if msg.Topic == "" {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "topic is required for subscription"}`))
				continue
			}
			s.SubscribeToTopic(userID, msg.Topic)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"status": "subscribed", "topic": "`+msg.Topic+`"}`))

		case "direct":
			if msg.To == "" || msg.Message == "" {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "both 'to' and 'message' are required for direct messages"}`))
				continue
			}
			if err := s.SendDirectMessage(msg.To, []byte(msg.Message)); err != nil {
				log.Printf("Failed to send direct message from user %s to user %s: %v", userID, msg.To, err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "failed to send message"}`))
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"status": "message sent"}`))
			}

		default:
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unknown message type"}`))
		}
	}
}
