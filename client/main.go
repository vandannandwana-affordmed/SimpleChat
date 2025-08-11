package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vandan/chat-app/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewChatServiceClient(conn)

	// Replace with a valid JWT token from /login
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQ4NDkzNzMsInVzZXJfaWQiOiIyIn0.fpn1VQEGsBsFBh8tLXyeelC9Mu0eZ89HBu-XQJ-NwSw" // e.g., "eyJhbGciOiJIUzI1NiJ9..."
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)

	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("Failed to start chat: %v", err)
	}

	// Receive messages
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Println("Stream closed by server")
				return
			}
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}
			fmt.Printf("[%s] From %s: %s\n", time.Unix(msg.Timestamp, 0).Format(time.RFC3339), msg.UserId, msg.Content)
		}
	}()

	// Send messages
	reader := bufio.NewReader(os.Stdin)
	for {
		var recipientID string
		fmt.Print("Enter recipient ID: ")
		recipientID, _ = reader.ReadString('\n')
		recipientID = strings.TrimSpace(recipientID)

		fmt.Print("Enter message: ")
		content, _ := reader.ReadString('\n')
		content = strings.TrimSpace(content)

		err := stream.Send(&proto.ChatMessage{
			UserId:      "2", // Must match user_id in JWT token
			RecipientId: recipientID,
			Content:     content,
			Timestamp:   time.Now().Unix(),
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
			return
		}
		fmt.Printf("Sent to %s: %s\n", recipientID, content)
	}
}