package main

import (
	db "SA/lib/DB"
	"SA/lib/common"
	"SA/lib/handlers"
	tw "SA/lib/twitch"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var (
	clientID     string
	clientSecret string
	conf         *oauth2.Config
)

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
	app := tw.App{Config: conf}
	mux := http.NewServeMux()
	mux.Handle("/", logWrapperHandler(fs))
	mux.HandleFunc("POST /twitch/getChannelID", logWrapperFunc(handlers.GetChannelIDByName))
	mux.HandleFunc("GET /twitch/oauth", logWrapperFunc(app.OAuthHandler))
	mux.HandleFunc("GET /twitch/callback", logWrapperFunc(app.OAuthCallbackHandler))
	mux.HandleFunc("GET /login", logWrapperFunc(handlers.LoginGet))
	mux.HandleFunc("POST /login", logWrapperFunc(handlers.LoginPost))
	mux.HandleFunc("GET /logout", logWrapperFunc(handlers.Logout))
	port := ":3000"
	fmt.Println("Server is running on port" + port)

	log.Fatal(http.ListenAndServe(port, mux))
}

// Create and setup everything
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db.InitDatabase()
	common.InitJWT()
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://id.twitch.tv/oauth2/authorize",
		TokenURL: "https://id.twitch.tv/oauth2/token",
	}
	conf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:3000/twitch/callback",
		Scopes:       []string{},
		Endpoint:     endpoint,
	}
	type App struct {
		config *oauth2.Config
	}
}

func main() {
	con, _ := db.OpenDB()
	newAppUser := db.AppUser{
		Name:        "Stumpy",
		Pass:        "Somepass",
		Permissions: db.Owner,
		Channels:    db.ChannelPerm{"Stumpy": 4},
	}

	db.AddAppUser(con, newAppUser)
	if user, err := db.GetAppUserByName(con, "Stumpy"); err != nil {
		println(err.Error())
		os.Exit(1)
	} else {
		log.Printf("Username: %s", user.Name)
		log.Printf("Channels: %d", user.Channels["d2ans0"])
	}

	webServer()
}
