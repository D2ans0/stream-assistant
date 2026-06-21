package twitch

import (
	db "SA/lib/DB"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

type App struct {
	Config *oauth2.Config
}

// Sends the request to the oAuth provider
func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	url := a.Config.AuthCodeURL("StreamAssistant", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Handles the redirect from the oAuth provider.
func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	// Exchanging code for token
	userToken, err := a.Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("GOT ERROR %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validatedToken, err := a.OAuthValidate(*userToken)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(w, "Failed to validate oAuth token")
		return
	}
	twitchUser := db.TwitchUser{
		UserID:            validatedToken.User_id,
		UserName:          validatedToken.Login,
		AccessToken:       userToken.AccessToken,
		RefreshToken:      userToken.AccessToken,
		AccessTokenExpiry: userToken.Expiry.Unix(),
	}
	con, _ := db.OpenDB()
	defer db.AddTwitchUser(con, twitchUser)
	fmt.Fprintln(w, "Successfully authorized Stream Assistant!")
	fmt.Fprintf(w, "Login: %s", validatedToken.Login)
	fmt.Fprintf(w, "UserID: %s", validatedToken.User_id)
	fmt.Fprintf(w, "ClienID (ID of this application): %s", validatedToken.Client_id)
}

// Validates token and returns user object
func (a *App) OAuthValidate(token oauth2.Token) (UserAuth, error) {
	hc := http.Client{}
	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err.Error())
		return UserAuth{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token.AccessToken))
	res, err := hc.Do(req)
	if err != nil {
		return UserAuth{}, err
	}
	validatedInfo := UserAuth{}
	body, _ := io.ReadAll(res.Body)
	println(string(body))
	if err := json.Unmarshal(body, &validatedInfo); err != nil {
		log.Println(err.Error())
	}

	println(validatedInfo.Client_id)
	println(validatedInfo.Expires_in)
	println(validatedInfo.Login)
	println(validatedInfo.Scopes)
	println(validatedInfo.User_id)

	return validatedInfo, nil
}
