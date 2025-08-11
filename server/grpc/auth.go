package grpc

import (
	"context"
	"log"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("No metadata provided")
		return nil, status.Error(codes.Unauthenticated, "Metadata not provided")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		log.Println("Authorization header missing")
		return nil, status.Error(codes.Unauthenticated, "Authorization header required")
	}

	parts := strings.Split(authHeader[0], " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Println("Invalid authorization header format")
		return nil, status.Error(codes.Unauthenticated, "Invalid authorization header")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		return []byte("vandan"), nil
	})

	if err != nil || !token.Valid {
		log.Printf("Token validation failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Invalid token claims")
		return nil, status.Error(codes.Unauthenticated, "Invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		log.Println("User ID not found in claims")
		return nil, status.Error(codes.Unauthenticated, "User ID not found in claims")
	}

	ctx = context.WithValue(ctx, "user_id", userID)
	return handler(ctx, req)
}

func StreamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
    md, ok := metadata.FromIncomingContext(ss.Context())
    if !ok {
        return status.Error(codes.Unauthenticated, "Metadata not provided")
    }

    authHeader, ok := md["authorization"]
    if !ok || len(authHeader) == 0 {
        return status.Error(codes.Unauthenticated, "Authorization header required")
    }

    parts := strings.Split(authHeader[0], " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return status.Error(codes.Unauthenticated, "Invalid authorization header")
    }

    token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
        return []byte("vandan"), nil
    })
    if err != nil || !token.Valid {
        return status.Error(codes.Unauthenticated, "Invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return status.Error(codes.Unauthenticated, "Invalid token claims")
    }

    userID, ok := claims["user_id"].(string)
    if !ok {
        return status.Error(codes.Unauthenticated, "User ID not found in claims")
    }

    // inject user_id into context
    newCtx := context.WithValue(ss.Context(), "user_id", userID)

    // wrap the original stream with the new context
    wrapped := &wrappedStream{ServerStream: ss, ctx: newCtx}

    return handler(srv, wrapped)
}

type wrappedStream struct {
    grpc.ServerStream
    ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
    return w.ctx
}
