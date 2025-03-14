package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Profile struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Bio        string `json:"bio"`
	ProfileImg string `json:"profile_img"`
}

type ProfileClient interface {
	GetProfileByUserID(userID uint) (*Profile, error)
}

type profileClient struct {
	baseURL string
}

func NewProfileClient(baseURL string) *profileClient {
	return &profileClient{
		baseURL: baseURL,
	}
}

func (c *profileClient) GetProfileByUserID(userID uint) (*Profile, error) {
	resp, err := http.Get(fmt.Sprintf("%s/user/profiles/%d", c.baseURL, userID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get profile, status: %d", resp.StatusCode)
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}
