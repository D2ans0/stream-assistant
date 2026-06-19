package handlers

import (
	db "SA/lib/DB"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const authCookieName = "Auth"

func LoginGet(w http.ResponseWriter, r *http.Request) {
	existingCookie, err := r.Cookie(authCookieName)
	if err == nil && ValidateJWT(existingCookie.Value) {
		fmt.Fprintf(w, "Already logged in. Token: %s", existingCookie.Value)
	} else {
		http.ServeFile(w, r, "web/login.html")
	}
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming request to " + r.RequestURI)
	userName := r.FormValue("user")
	userPass := r.FormValue("pass")
	rememberMe := r.FormValue("remember")
	con, err := db.OpenDB()
	if err != nil {
		println(err.Error())
		return
	}
	userDB, err := db.GetAppUserByName(con, userName)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if userDB.Pass == userPass {
		claims := jwt.MapClaims{
			"ID":    25,
			"Name":  "Stumpy",
			"Admin": true,
		}
		jwt, err := SignedJWT(claims)
		if err != nil {
			log.Println(err.Error())
			return
		}
		cookie := http.Cookie{
			Name:     authCookieName,
			Value:    jwt,
			Path:     "/",
			Expires:  time.Now().Add(600 * time.Second),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		}
		if rememberMe == "on" {
			cookie.Expires = time.Now().Add(time.Hour * 365 * 24)

		}
		http.SetCookie(w, &cookie)
		fmt.Fprintf(w, "\nLogged in as %s", userName)
	} else {
		http.Error(w,
			fmt.Sprintf("%d: Unauthorized", http.StatusUnauthorized),
			http.StatusUnauthorized,
		)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	existingCookie, err := r.Cookie(authCookieName)
	if err != nil {
		println(err.Error())
		return
	}
	existingCookie.Value = ""
	existingCookie.Expires = time.Unix(0, 0)
	log.Println(existingCookie.Expires)
	http.SetCookie(w, existingCookie)
	http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
}
