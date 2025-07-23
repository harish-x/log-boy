package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func NewAuthStreamInterceptor(db *gorm.DB, privateKey string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		serviceNameValues := md.Get("servicename")
		if len(serviceNameValues) == 0 {
			return status.Error(codes.Unauthenticated, "servicename header is not provided")
		}
		projectKey := serviceNameValues[0]

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		token := authHeaders[0]

		var storedKey struct {
			Key       string `json:"key"`
			Value     string `json:"value"`
			Timestamp int64  `json:"timestamp"`
		}

		result := db.WithContext(ss.Context()).
			Table("key_stores").
			Select("key", "value", "timestamp").
			Where("key = ?", projectKey).
			First(&storedKey)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				log.Printf("Authentication failed: project name not found: %s", projectKey)
				return status.Error(codes.Unauthenticated, "invalid credentials")
			}
			log.Printf("Database error during auth: %v", result.Error)
			return status.Error(codes.Internal, "database error")
		}
		token = strings.TrimPrefix(token, "Bearer ")

		payload := fmt.Sprintf("%s.%s.%d", projectKey, storedKey.Value, storedKey.Timestamp)

		if !validateToken(privateKey, token, payload) {
			log.Printf("Authentication failed: invalid token: %s", token)
			return status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return handler(srv, ss)
	}
}

func validateToken(privateKey string, expectedToken string, payload string) bool {
	h := hmac.New(sha256.New, []byte(privateKey))
	h.Write([]byte(payload))
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(expectedToken))
}
