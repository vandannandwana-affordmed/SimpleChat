package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/vandan/chat-app/proto"
	"github.com/vandan/chat-app/server/db"
	myGrpc "github.com/vandan/chat-app/server/grpc"
	"github.com/vandan/chat-app/server/rest"
	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	
	// Initialize REST server
	router := gin.Default()
	rest.RegisterRoutes(router, dbConn)

	// Start REST server
	restServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST server failed: %v", err)
		}
	}()

	// Initialize gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(myGrpc.AuthInterceptor), grpc.StreamInterceptor(myGrpc.StreamAuthInterceptor))
	chatServer := myGrpc.NewChatServer(dbConn)
	proto.RegisterChatServiceServer(grpcServer, chatServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()
	restServer.Shutdown(context.Background())
	log.Println("Servers shut down gracefully")
}
