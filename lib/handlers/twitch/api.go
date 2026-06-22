package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func GetAppBearerToken(clientID string, clientSecret string) (BearerToken, error) {
	var result BearerToken
	hc := http.Client{}
	url := "https://id.twitch.tv/oauth2/token?client_id=" + clientID + "&client_secret=" + clientSecret + "&grant_type=client_credentials"
	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return result, err
	}
	res, err := hc.Do(req)

	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("ERROR: Cannot unmarshal JSON")
	}
	return result, nil
}

func GetChannelIDByName(clientID string, token string, channelName string) (*User, error) {
	var result User
	hc := http.Client{}
	url := "https://api.twitch.tv/helix/users?login=" + channelName
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("client-id", clientID)
	req.Header.Add("Authorization", "Bearer "+token)
	if err != nil {
		return nil, err
	}

	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		log.Println(res.StatusCode)
		return nil, errors.New(string(body))
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("ERROR: Cannot unmarshal JSON")
		return nil, errors.New("ERROR: Cannot unmarshal JSON")
	}
	if len(result.Data) == 0 {
		return nil, errors.New("User not found")
	}
	return &result, nil
}

func SetStreamTitle(channelID string, token string, clientID string, title string) {
	var result User
	broadcastProperties := BroadcastProperties{
		Title: &title,
	}
	var reqBody []byte
	reqBody, err := json.Marshal(broadcastProperties)
	log.Println(string(reqBody))
	// return
	// var writer io.Writer
	hc := http.Client{}
	url := "https://api.twitch.tv/helix/channels?broadcaster_id=" + channelID
	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(reqBody)))
	req.Header.Add("client-id", clientID)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("err.Error(): %v\n", err.Error())
	}

	res, err := hc.Do(req)
	if err != nil {
		fmt.Printf("err.Error(): %v\n", err.Error())
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	log.Println(req.Header.Get("client-id"))
	log.Println(url)
	log.Println(string(body))
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("ERROR: Cannot unmarshal JSON")
	}
}
