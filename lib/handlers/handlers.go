package handlers

import (
	tw "SA/lib/twitch"
	"fmt"
	"net/http"
)

var (
	clientID     string
	clientSecret string
)

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

func GetChannelIDByName(w http.ResponseWriter, r *http.Request) {
	channelName := r.FormValue("channelName")
	token, _ := tw.GetAppBearerToken(clientID, clientSecret)
	user, err := tw.GetChannelIDByName(clientID, token.AccessToken, channelName)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	}
	fmt.Fprintf(w, "%s", user.Data[0].ID)
}
