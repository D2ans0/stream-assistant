package main

import (
	db "SA/lib/DB"
	"SA/lib/common"
	"SA/lib/handlers"

	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var conf *oauth2.Config

func logWrapperFunc(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf(`%s: "%s %s" - "%s"`, r.RemoteAddr, r.Method, r.RequestURI, r.UserAgent())
		handler(w, r)
	}
}

func logWrapperHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(`%s: "%s %s" - "%s"`, r.RemoteAddr, r.Method, r.RequestURI, r.UserAgent())
		handler.ServeHTTP(w, r)
	})
}

func webServer() {
	fs := http.FileServer(http.Dir("./web"))
	mux := http.NewServeMux()
	mux.Handle("/", logWrapperHandler(fs))
	mux.HandleFunc("GET /login", logWrapperFunc(handlers.LoginGet))
	mux.HandleFunc("POST /login", logWrapperFunc(handlers.LoginPost))
	mux.HandleFunc("GET /logout", logWrapperFunc(handlers.Logout))
	mux.HandleFunc("GET /dashboard", logWrapperFunc(handlers.Dashboard))
	mux.HandleFunc("GET /user/GetUserChannels", logWrapperFunc(handlers.GetUserChannels))
	mux.HandleFunc("GET /twitch/oauth", logWrapperFunc(handlers.TwitchOauth))
	mux.HandleFunc("GET /twitch/callback", logWrapperFunc(handlers.TwitchOAuthCallback))
	mux.HandleFunc("POST /twitch/getChannelID", logWrapperFunc(handlers.GetChannelIDByName))
	port := ":3000"
	fmt.Println("Server is running on port" + port)

	log.Fatal(http.ListenAndServe(port, mux))
}

// Create and setup everything
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db.InitDatabase()
	common.InitJWT()
}

func main() {
	con, _ := db.OpenDB()
	// actions := db.ChannelAction{
	// 	ActionType: db.Add,
	// 	PermLevel:  db.User,
	// }
	// db.ModifyAppUserChannel(con, "Stumpy", "Poppies", actions)
	// actions.PermLevel = db.Owner
	// db.ModifyAppUserChannel(con, "Stumpy", "d2ans0", actions)
	// db.ModifyAppUserChannel(con, "Stumpy", "Poppies2", actions)
	// actions.ActionType = db.Modify
	// actions.PermLevel = db.Admin
	// db.ModifyAppUserChannel(con, "Stumpy", "d2ans0", actions)
	// actions.ActionType = db.Remove
	// db.ModifyAppUserChannel(con, "Stumpy", "Poppies", actions)
	// os.Exit(0)
	// newAppUser := db.AppUser{
	// 	Name:        "Stumpy",
	// 	Pass:        "123",
	// 	Permissions: db.Owner,
	// 	Channels:    db.ChannelPerm{"Stumpy": 4},
	// }

	// db.AddAppUser(con, newAppUser)
	// app := tw.GetConfig()
	// user, _ := db.GetTwitchUserByName(con, "d2ans0")
	// app.RefreshAccessTokenForUser(user)
	// os.Exit(1)
	if user, err := db.GetAppUserByName(con, "Stumpy"); err != nil {
		println(err.Error())
		os.Exit(1)
	} else {
		log.Printf("Username: %s", user.Name)
		log.Printf("Channels: %d", user.Channels["d2ans0"])
	}

	webServer()
}
