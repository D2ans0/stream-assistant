package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

type App struct {
	Config *oauth2.Config
}

type UserAuth struct {
	Client_id  string   `json:"client_id"`
	Login      string   `json:"login"`
	Scopes     []string `json:"scopes"`
	User_id    string   `json:"user_id"`
	Expires_in int      `json:"expires_in"`
}

// Login page for twitch app
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	existingCookie, _ := a.OAuthGetCookie(w, r)
	w.Header().Add("Content-Type", "text/html")
	if existingCookie != nil {
		fmt.Fprintf(w, "Already logged in. Token: %s", existingCookie.Value)
	} else {
		fmt.Fprint(w, "Not logged in, navigate to <a href=\"/auth/oauth\">/auth/oauth</a>")
	}
}

// Sends the request to the oAuth provider
func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	// authCode, _ := oauth2.AuthCodeOption
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

	// Save cookie in user's browser
	cookie := http.Cookie{
		Name:     "SAoAuth2",
		Value:    userToken.AccessToken,
		Path:     "/",
		Expires:  userToken.Expiry,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	fmt.Println(w.Header().Get("Set-Cookie"))
	fmt.Fprintln(w, "Successfully authorized Stream Assistant!")
	fmt.Fprintf(w, "Token type: %s\n", userToken.TokenType)
	fmt.Fprintf(w, "Access token: %s\n", userToken.AccessToken)
	fmt.Fprintf(w, "Refresh token: %s\n", userToken.RefreshToken)
	fmt.Fprintf(w, "Expires in %d seconds\n", userToken.ExpiresIn)
	fmt.Fprintln(w, "DO NOT SHARE THIS INFO")
	a.OAuthValidate(*userToken)
}

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

func (a *App) oAuthSetCookie(w http.ResponseWriter, r *http.Request, cookie http.Cookie) {
	existingCookie, _ := a.OAuthGetCookie(w, r)
	log.Printf("Setting cookie: %v", cookie)
	if existingCookie != nil {
		log.Printf("Found cookie: %s", existingCookie)
		return
	}
	http.SetCookie(w, &cookie)
}

func (a *App) OAuthGetCookie(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie("SAoAuth2")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			return nil, err
		default:
			log.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
	}
	return cookie, nil
}
