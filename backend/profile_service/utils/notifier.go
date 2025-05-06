package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func NotifyBlogsService(userID uint, username string) {
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":  userID,
		"username": username,
	})
	_, err := http.Post(
		"http://blogs-service:8081/internal/sync-username",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("blogs-service sync failed: %v", err)
	}
}
