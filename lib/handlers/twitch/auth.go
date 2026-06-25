package twitch

import (
	db "SA/lib/DB"
	"errors"
	"os"
	"time"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

var (
	ConfigPath   string = "./config.json"
	Config       map[string]string
	clientID     string
	clientSecret string
	redirectURL  string
)

type App struct {
	Config *oauth2.Config
}

func init() {
	if keyData, err := os.ReadFile(ConfigPath); err == nil {
		if err := json.Unmarshal(keyData, &Config); err != nil {
			panic(err)
		}
	}
}

func GetConfig() App {
	clientID = Config["ClientID"]
	clientSecret = Config["ClientSecret"]
	redirectURL = Config["RedirectURL"]
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://id.twitch.tv/oauth2/authorize",
		TokenURL: "https://id.twitch.tv/oauth2/token",
	}
	conf := App{Config: &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"channel_editor"},
		Endpoint:     endpoint,
	}}
	return conf
}

// Sends the request to the oAuth provider
func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	url := a.Config.AuthCodeURL("StreamAssistant", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (a *App) getAuthClient(token *oauth2.Token) *http.Client {
	return a.Config.Client(context.Background(), token)
}

// Handles the redirect from the oAuth provider.
func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request, loggedInUser string) (db.TwitchUser, error) {
	code := r.URL.Query().Get("code")

	// Exchanging code for token
	userToken, err := a.Config.Exchange(context.Background(), code)
	if err != nil {
		return db.TwitchUser{}, err
	}

	validatedToken, err := a.OAuthValidate(*userToken)
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidKey) {

		}
		return db.TwitchUser{}, errors.New("Failed to validate oAuth token\n" + err.Error())
	}
	twitchUser := db.TwitchUser{
		UserID:            validatedToken.User_id,
		UserName:          validatedToken.Login,
		AccessToken:       userToken.AccessToken,
		RefreshToken:      userToken.RefreshToken,
		AccessTokenExpiry: userToken.Expiry.Unix(),
	}
	con, _ := db.OpenDB()
	if err := db.ModifyAppUserChannel(con, loggedInUser, twitchUser.UserName, db.ChannelAction{
		ActionType: db.Add,
		PermLevel:  db.Owner,
	}); err != nil {
		println(err.Error())
	}
	fmt.Fprintln(w, "Successfully authorized Stream Assistant!")
	fmt.Fprintf(w, "Login: %s", twitchUser.UserName)
	fmt.Fprintf(w, "UserID: %s", twitchUser.UserID)
	fmt.Fprintf(w, "ClienID (ID of this application): %s", validatedToken.Client_id)
	return twitchUser, nil
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
	// println(string(body))
	if err := json.Unmarshal(body, &validatedInfo); err != nil {
		return UserAuth{}, err
	}

	return validatedInfo, nil
}

func (a *App) refreshToken(token *oauth2.Token) *oauth2.Token {
	tokenSource := a.Config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return newToken
}

func (a *App) RefreshAccessTokenForUser(user *db.TwitchUser) error {
	token := new(oauth2.Token)
	token.AccessToken = user.AccessToken
	token.RefreshToken = user.RefreshToken
	token.Expiry = time.Unix(user.AccessTokenExpiry, 0)
	token.TokenType = "Bearer"
	newToken := a.refreshToken(token)
	if con, err := db.OpenDB(); err != nil {
		return err
	} else {
		println(newToken.AccessToken)
		db.UpdateTwitchUserAccessTokenByID(con, user.UserID, newToken.AccessToken)
		return nil
	}
}
