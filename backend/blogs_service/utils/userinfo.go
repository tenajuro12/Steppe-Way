package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func GetUsername(userID uint) (string, error) {
	url := fmt.Sprintf("%s/user/profiles/%d",
		os.Getenv("PROFILE_SERVICE_URL"), userID)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Username, nil
}
