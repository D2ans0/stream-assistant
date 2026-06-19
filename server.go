package main

import (
	db "SA/lib/DB"
	"SA/lib/handlers"
	tw "SA/lib/twitch"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

var (
	clientID     string
	clientSecret string
)

func webServer() {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://id.twitch.tv/oauth2/authorize",
		TokenURL: "https://id.twitch.tv/oauth2/token",
	}
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:3000/auth/callback",
		Scopes:       []string{},
		Endpoint:     endpoint,
	}
	type App struct {
		config *oauth2.Config
	}
	fs := http.FileServer(http.Dir("./web"))
	app := tw.App{Config: conf}
	mux := http.NewServeMux()

	mux.Handle("/", fs)
	// mux.HandleFunc("GET /", handlers.Root)
	mux.HandleFunc("POST /getChannelID", handlers.GetChannelIDByName)
	mux.HandleFunc("GET /auth/login", app.LoginHandler)
	mux.HandleFunc("GET /auth/oauth", app.OAuthHandler)
	mux.HandleFunc("GET /auth/callback", app.OAuthCallbackHandler)
	mux.HandleFunc("GET /login", handlers.LoginGet)
	mux.HandleFunc("POST /login", handlers.LoginPost)
	mux.HandleFunc("GET /logout", handlers.Logout)
	port := ":3000"
	fmt.Println("Server is running on port" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db.InitDatabase()
	con, _ := db.OpenDB()
	newAppUser := db.AppUser{
		Name:  "Stumpy",
		Pass:  "Somepass",
		Salt:  "asalt",
		Admin: true,
	}
	db.AddAppUser(con, newAppUser)
	user, _ := db.GetAppUserByName(con, "Stumpy")
	log.Printf("Username: %s", user.Name)

	handlers.Init()
	webServer()
}
