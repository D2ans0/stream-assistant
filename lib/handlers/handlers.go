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

// loads config on startup
func init() {
	if keyData, err := os.ReadFile(ConfigPath); err == nil {
		if err := json.Unmarshal(keyData, &Config); err != nil {
			panic(err)
		}
	}
}

// Root handler, currently unused
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

// Serve dashboard
func Dashboard(w http.ResponseWriter, r *http.Request) {
	if user := loggedInUser(r); user != nil {
		http.ServeFile(w, r, "web/dashboard.html")
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

// Returns channel ID when passed a name
func GetChannelIDByName(w http.ResponseWriter, r *http.Request) {
	channelName := r.FormValue("channelName")
	clientID = Config["ClientID"]
	clientSecret = Config["ClientSecret"]
	token, err := tw.GetAppBearerToken(clientID, clientSecret)
	user, err := tw.GetChannelIDByName(clientID, token.AccessToken, channelName)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	} else {
		fmt.Fprintf(w, "%s", *user)
	}
}

// Needs to be passed a channel, and title form values, also JWT cookie needs to be set
func SetChannelStreamTitle(w http.ResponseWriter, r *http.Request) {
	if user := loggedInUser(r); user != nil {
		con := db.GetConnection()
		defer con.Con.Close()
		channelName := r.FormValue("channel")
		desiredTitle := r.FormValue("title")
		if con.IsActionAllowedForUser(*user, channelName, db.User) {
			log.Printf("%s requested title change to \"%s\"", *user, desiredTitle)
			if err := tw.SetStreamTitle(channelName, Config["ClientID"], desiredTitle); err != nil {
				fmt.Fprint(w, "Failed to change title")
				log.Println(err.Error())
			} else {
				fmt.Fprint(w, "Title changed!")
			}
		}
	}
}

func GetChannelStreamTitle(w http.ResponseWriter, r *http.Request) {
	if user := loggedInUser(r); user != nil {
		if channelName := r.URL.Query().Get("channel"); channelName != "" {
			title, err := tw.GetStreamTitle(channelName, Config["ClientID"])
			if err != nil {
				log.Println(err.Error())
				return
			}
			fmt.Fprint(w, *title)
		}
	}
}

// Handle beginning of oauth2 flow or redirect to login if not logged into the app
func TwitchOauth(w http.ResponseWriter, r *http.Request) {
	if loggedInUser(r) != nil {
		oAuth2 := tw.GetConfig()
		oAuth2.OAuthHandler(w, r)
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

// Handle callback from twitch, and add the token values to database
func TwitchOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if user := loggedInUser(r); user != nil {
		oAuth2 := tw.GetConfig()
		user, err := oAuth2.OAuthCallbackHandler(w, r, *user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			con := db.GetConnection()
			defer con.Con.Close()
			defer con.AddOrReplaceTwitchUser(user)
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

// Returns JSON with channels available to users and their access level to them
func GetUserChannels(w http.ResponseWriter, r *http.Request) {
	err := errors.New("Empty Error")

	if user := loggedInUser(r); user != nil {
		con := db.GetConnection()
		defer con.Con.Close()
		channelMap, err := con.GetUserAccessibleChannels(*user)
		if err == nil {
			jsonStr, err := json.Marshal(channelMap)
			if err == nil {
				w.Write(jsonStr)
				return
			}
		}
	} else {
		http.Error(w, "401: Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Failed to get user Channels:\n%s", err.Error())
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
