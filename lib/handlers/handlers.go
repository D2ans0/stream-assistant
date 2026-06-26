package handlers

import (
	db "SA/lib/DB"
	tw "SA/lib/handlers/twitch"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	ConfigPath   string = "./config.json"
	Config       map[string]string
	clientID     string
	clientSecret string
)

func init() {
	if keyData, err := os.ReadFile(ConfigPath); err == nil {
		if err := json.Unmarshal(keyData, &Config); err != nil {
			panic(err)
		}
	}
}

func Root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, `
			<body>
			<form action="/getChannelID" method="post" id="channelNameForm">
				<label for="channelName">Channel Name</label>
				<input type="text" id="channelName" name="channelName">
				<input type="submit" value="Submit">
			</form>
			</body>
			`)
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/dashboard.html")
}

func GetChannelIDByName(w http.ResponseWriter, r *http.Request) {
	channelName := r.FormValue("channelName")
	// token, _ := tw.GetAppBearerToken(clientID, clientSecret)
	clientID = Config["ClientID"]
	clientSecret = Config["ClientSecret"]
	token, err := tw.GetAppBearerToken(clientID, clientSecret)
	user, err := tw.GetChannelIDByName(clientID, token.AccessToken, channelName)
	// TODO: make a proper handler for setting stream title
	tw.SetStreamTitle("d2ans0", clientID, "this is a test")
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	} else {
		fmt.Fprintf(w, "%s", user.Data[0].ID)
	}
}

func TwitchOauth(w http.ResponseWriter, r *http.Request) {
	if loggedInUser(r) != nil {
		oAuth2 := tw.GetConfig()
		oAuth2.OAuthHandler(w, r)
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

func TwitchOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if user := loggedInUser(r); user != nil {
		oAuth2 := tw.GetConfig()
		user, err := oAuth2.OAuthCallbackHandler(w, r, *user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			con, _ := db.OpenDB()
			defer db.AddOrReplaceTwitchUser(con, user)
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

func GetUserChannels(w http.ResponseWriter, r *http.Request) {
	err := errors.New("Empty Error")
	if user := loggedInUser(r); user != nil {
		con, err := db.OpenDB()
		if err == nil {
			channelMap, err := db.GetUserAccessibleChannels(con, *user)
			if err == nil {
				jsonStr, err := json.Marshal(channelMap)
				if err == nil {
					w.Write(jsonStr)
					return
				}
			}
		}
	} else {
		http.Error(w, "401: Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Failed to get user Channels:\n%s", err.Error())
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
