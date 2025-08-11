package grpc

import (
	"database/sql"
	"io"
	"log"
	"sync"

	"github.com/vandan/chat-app/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
	db           *sql.DB
	clients      map[string]chan *proto.ChatMessage
	clientsMutex sync.Mutex
	proto.UnimplementedChatServiceServer
}

func NewChatServer(db *sql.DB) *ChatServer {
	return &ChatServer{
		db:      db,
		clients: make(map[string]chan *proto.ChatMessage),
	}
}

func (s *ChatServer) Chat(stream proto.ChatService_ChatServer) error {
	userID, ok := stream.Context().Value("user_id").(string)
	if !ok {
		return status.Error(codes.Unauthenticated, "User ID not found in context")
	}

	msgChan := make(chan *proto.ChatMessage, 100)
	s.clientsMutex.Lock()
	s.clients[userID] = msgChan
	s.clientsMutex.Unlock()

	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients, userID)
		s.clientsMutex.Unlock()
		close(msgChan)
	}()

	// Send messages to client
	go func() {
		for msg := range msgChan {
			if err := stream.Send(msg); err != nil {
				log.Printf("Error sending message to %s: %v", userID, err)
				return
			}
		}
	}()

	// Receive messages from client
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Save message to database
		_, err = s.db.Exec(
			"INSERT INTO messages (sender_id, recipient_id, content, created_at) VALUES ($1, $2, $3, $4)",
			userID, msg.RecipientId, msg.Content, msg.Timestamp,
		)
		if err != nil {
			log.Printf("Failed to save message: %v", err)
		}

		// Forward message to recipient
		s.clientsMutex.Lock()
		if recipientChan, exists := s.clients[msg.RecipientId]; exists {
			recipientChan <- msg
		}
		s.clientsMutex.Unlock()
	}

}
