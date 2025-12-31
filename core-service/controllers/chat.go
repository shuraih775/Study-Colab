package controllers

import (
	"core-service/config"
	"core-service/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"core-service/internal/file"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatHandler struct {
	server *Server
}

type ClientAttachmentDTO struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
}

type ClientMessage struct {
	Type        string                `json:"type"` // "join", "leave", "broadcast"
	Room        string                `json:"room"`
	Text        string                `json:"text"`
	Attachments []ClientAttachmentDTO `json:"attachments"`
	client      *Client               `json:"-"`
}

type ServerAttachmentDTO struct {
	ID       string `json:"id"`
	FileURL  string `json:"file_url"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
}

type ServerMessage struct {
	ID          string                `json:"id"`
	User        User                  `json:"user"`
	Text        string                `json:"text"`
	Attachments []ServerAttachmentDTO `json:"attachments"`
	Timestamp   time.Time             `json:"timestamp"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	server   *Server
	conn     *websocket.Conn
	send     chan []byte
	rooms    map[string]bool
	UserID   string
	Username string
}

type Hub struct {
	roomID     string
	server     *Server
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	stop       chan bool
}

func newHub(roomID string, server *Server) *Hub {
	return &Hub{
		roomID:     roomID,
		server:     server,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan bool),
	}
}

func NewChatHandler(server *Server) *ChatHandler {
	return &ChatHandler{server: server}
}

func (h *Hub) run() {
	defer func() { log.Printf("Hub %s stopped", h.roomID) }()
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if len(h.clients) == 0 {
					h.server.removeHub(h.roomID)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case <-h.stop:
			for client := range h.clients {
				delete(h.clients, client)
				close(client.send)
			}
			return
		}
	}
}

type Server struct {
	hubs       map[string]*Hub
	clients    map[*Client]bool
	broadcast  chan *ClientMessage
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	fileClient *file.Client
}

func NewServer(fileClient *file.Client) *Server {
	return &Server{
		hubs:       make(map[string]*Hub),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *ClientMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		fileClient: fileClient,
	}
}

func (s *Server) getHub(roomID string) (*Hub, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	hub, ok := s.hubs[roomID]
	return hub, ok
}

func (s *Server) getOrCreateHub(roomID string) *Hub {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if hub, ok := s.hubs[roomID]; ok {
		return hub
	}
	hub := newHub(roomID, s)
	s.hubs[roomID] = hub
	go hub.run()
	return hub
}

func (s *Server) removeHub(roomID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if hub, ok := s.hubs[roomID]; ok {
		close(hub.stop)
		delete(s.hubs, roomID)
	}
}

func (s *Server) fetchHistory(client *Client, roomID string) {
	var messages []models.ChatMessage

	err := config.DB.Preload("User").
		Preload("Attachments").
		Where("room_id = ?", roomID).
		Order("timestamp desc").
		Limit(50).
		Find(&messages).Error

	if err != nil {
		log.Printf("Error fetching history: %v", err)
		return
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		s.sendSingleMessageToClient(client, &msg)
	}
}

func (s *Server) sendSingleMessageToClient(client *Client, msg *models.ChatMessage) {
	var attDTOs []ServerAttachmentDTO

	for _, att := range msg.Attachments {
		downloadURL, err := s.fileClient.GenerateDownloadURL(att.FileID)
		if err != nil {
			log.Printf("failed to generate download URL: %v", err)
			continue
		}

		attDTOs = append(attDTOs, ServerAttachmentDTO{
			ID:       att.ID.String(),
			FileURL:  downloadURL,
			FileType: att.FileType,
			FileSize: att.FileSize,
		})
	}

	serverMsg := &ServerMessage{
		ID:          msg.ID.String(),
		User:        User{ID: msg.UserID.String(), Name: msg.User.Username},
		Text:        msg.Text,
		Attachments: attDTOs,
		Timestamp:   msg.Timestamp,
	}

	msgBytes, _ := json.Marshal(serverMsg)
	client.send <- msgBytes
}

func (s *Server) saveMessage(client *Client, roomID string, text string, clientAtts []ClientAttachmentDTO) (*models.ChatMessage, error) {
	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	msg := &models.ChatMessage{
		RoomID:    roomID,
		UserID:    userID,
		Text:      text,
		Timestamp: time.Now(),
	}

	for _, att := range clientAtts {
		msg.Attachments = append(msg.Attachments, models.Attachment{
			FileID:   att.FileID,
			FileType: att.FileType,
			FileSize: att.FileSize,
		})
	}

	result := config.DB.Create(msg)
	if result.Error != nil {
		return nil, result.Error
	}

	err = config.DB.Preload("User").Preload("Attachments").First(msg, "id = ?", msg.ID).Error
	return msg, err
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				for roomID := range client.rooms {
					if hub, ok := s.getHub(roomID); ok {
						hub.unregister <- client
					}
				}
				close(client.send)
			}

		case clientMsg := <-s.broadcast:
			client := clientMsg.client

			switch clientMsg.Type {
			case "join":
				hub := s.getOrCreateHub(clientMsg.Room)
				hub.register <- client
				client.rooms[clientMsg.Room] = true
				go s.fetchHistory(client, clientMsg.Room)

			case "leave":
				if hub, ok := s.getHub(clientMsg.Room); ok {
					hub.unregister <- client
					delete(client.rooms, clientMsg.Room)
				}

			case "broadcast":
				if hub, ok := s.getHub(clientMsg.Room); ok {
					if _, inRoom := hub.clients[client]; inRoom {

						savedMsg, err := s.saveMessage(client, clientMsg.Room, clientMsg.Text, clientMsg.Attachments)
						if err != nil {
							log.Printf("Error saving message: %v", err)
							continue
						}

						var attDTOs []ServerAttachmentDTO
						for _, att := range savedMsg.Attachments {
							url, err := s.fileClient.GenerateDownloadURL(att.FileID)
							if err != nil {
								continue
							}
							attDTOs = append(attDTOs, ServerAttachmentDTO{
								ID:       att.ID.String(),
								FileURL:  url,
								FileType: att.FileType,
								FileSize: att.FileSize,
							})
						}

						serverMsg := &ServerMessage{
							ID:          savedMsg.ID.String(),
							User:        User{ID: savedMsg.UserID.String(), Name: savedMsg.User.Username},
							Text:        savedMsg.Text,
							Attachments: attDTOs,
							Timestamp:   savedMsg.Timestamp,
						}

						msgBytes, _ := json.Marshal(serverMsg)
						hub.broadcast <- msgBytes
					}
				}
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("Error unmarshaling: %v", err)
			continue
		}
		clientMsg.client = c
		c.server.broadcast <- &clientMsg

	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

func (h *ChatHandler) HandleConnection(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userVal, _ := c.Get("user")

	userIDUUID := userIDVal.(uuid.UUID)
	userModel := userVal.(models.User)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		server:   h.server,
		conn:     conn,
		send:     make(chan []byte, 256),
		rooms:    make(map[string]bool),
		UserID:   userIDUUID.String(),
		Username: userModel.Username,
	}

	h.server.register <- client
	go client.writePump()
	go client.readPump()
}
