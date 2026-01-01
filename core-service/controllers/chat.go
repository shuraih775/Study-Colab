package controllers

import (
	"context"
	"core-service/config"
	"core-service/models"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"core-service/internal/file"
	"core-service/internal/observability/logging"
	"core-service/internal/observability/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var chatTracer = otel.Tracer("controllers.chat")

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
	log      *zap.Logger
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
	h.server.log.Info("hub started", zap.String("room_id", h.roomID))
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
	log        *zap.Logger
}

func NewServer(fileClient *file.Client, log *zap.Logger) *Server {
	return &Server{
		hubs:       make(map[string]*Hub),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *ClientMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		fileClient: fileClient,
		log:        log,
	}
}

func (s *Server) ActiveConnections() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.clients)
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
	ctx := context.Background()

	ctx, span := chatTracer.Start(ctx, "chat.history.fetch")
	span.SetAttributes(
		attribute.String("room.id", roomID),
		attribute.String("user.id", client.UserID),
	)
	defer span.End()

	var messages []models.ChatMessage

	err := config.DB.WithContext(ctx).Preload("User").
		Preload("Attachments").
		Where("room_id = ?", roomID).
		Order("timestamp desc").
		Limit(50).
		Find(&messages).Error

	if err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "db query failed")

		client.log.Error(
			"failed to fetch chat history",
			zap.String("room_id", roomID),
			zap.Error(err),
		)
		return
	}

	client.log.Info(
		"chat history loaded",
		zap.String("room_id", roomID),
		zap.Int("count", len(messages)),
	)
	span.SetAttributes(attribute.Int("messages.count", len(messages)))

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		s.sendSingleMessageToClient(client, &msg)

	}
}

func (s *Server) sendSingleMessageToClient(client *Client, msg *models.ChatMessage) {
	ctx := context.Background()

	var attDTOs []ServerAttachmentDTO

	for _, att := range msg.Attachments {
		_, span := chatTracer.Start(ctx, "url.download.generate")
		span.SetAttributes(attribute.String("file.id", att.FileID))

		downloadURL, err := s.fileClient.GenerateDownloadURL(att.FileID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "file service failed")
			span.End()
			continue
		}

		span.End()

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
	ctx := context.Background()

	ctx, span := chatTracer.Start(ctx, "chat.message.save")
	span.SetAttributes(
		attribute.String("room.id", roomID),
		attribute.String("user.id", client.UserID),
		attribute.Int("attachments.count", len(clientAtts)),
	)
	defer span.End()

	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid user id")
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

	metrics.ChatMessagesSent.Add(ctx, 1)

	result := config.DB.WithContext(ctx).Create(msg)
	if result.Error != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db insert failed")
		return nil, result.Error
	}

	if err := config.DB.WithContext(ctx).Preload("User").
		Preload("Attachments").
		First(msg, "id = ?", msg.ID).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db reload failed")
		return nil, err
	}

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
				span := trace.SpanFromContext(context.Background())
				span.AddEvent("room.joined", trace.WithAttributes(attribute.String("room.id", clientMsg.Room)))
				hub := s.getOrCreateHub(clientMsg.Room)
				hub.register <- client
				client.rooms[clientMsg.Room] = true
				go s.fetchHistory(client, clientMsg.Room)

			case "leave":
				span := trace.SpanFromContext(context.Background())
				span.AddEvent("room.leave", trace.WithAttributes(attribute.String("room.id", clientMsg.Room)))
				if hub, ok := s.getHub(clientMsg.Room); ok {
					hub.unregister <- client
					delete(client.rooms, clientMsg.Room)
				}

			case "broadcast":
				span := trace.SpanFromContext(context.Background())
				span.AddEvent("room.joined", trace.WithAttributes(attribute.String("room.id", clientMsg.Room)))
				if hub, ok := s.getHub(clientMsg.Room); ok {
					if _, inRoom := hub.clients[client]; inRoom {

						savedMsg, err := s.saveMessage(client, clientMsg.Room, clientMsg.Text, clientMsg.Attachments)
						if err != nil {
							client.log.Error(
								"failed to save message",
								zap.String("room_id", clientMsg.Room),
								zap.Error(err),
							)
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
		c.log.Info("client disconnected")
		c.server.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.log.Warn("read error", zap.Error(err))
			return
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			c.log.Warn("invalid client message", zap.Error(err))
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
		c.log.Info("client writePump stopped")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.log.Warn("send channel closed, disconnecting client")
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.log.Warn("failed to write websocket message", zap.Error(err))
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Warn("websocket ping failed", zap.Error(err))
				return
			}
		}
	}
}

func (h *ChatHandler) HandleConnection(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	userIDVal, _ := c.Get("user_id")
	userVal, _ := c.Get("user")

	userID := userIDVal.(uuid.UUID)
	user := userVal.(models.User)

	ctx, span := chatTracer.Start(ctx, "chat.websocket.connect")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("user.username", user.Username),
	)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "websocket upgrade failed")
		log.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	span.AddEvent("websocket.upgraded")

	client := &Client{
		server:   h.server,
		conn:     conn,
		send:     make(chan []byte, 256),
		rooms:    make(map[string]bool),
		UserID:   userID.String(),
		Username: user.Username,
		log:      log,
	}

	h.server.register <- client
	span.AddEvent("client.registered")

	go client.writePump()
	go client.readPump()

	span.SetStatus(codes.Ok, "connection established")
}
