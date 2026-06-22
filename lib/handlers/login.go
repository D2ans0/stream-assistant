package handlers

import (
	db "SA/lib/DB"
	"SA/lib/common"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const authCookieName = "Auth"

func LoginGet(w http.ResponseWriter, r *http.Request) {
	if loggedInUser(r) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	} else {
		http.ServeFile(w, r, "web/login.html")
	}
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	userName := r.FormValue("user")
	userPass := r.FormValue("pass")
	rememberMe := r.FormValue("remember")
	con, err := db.OpenDB()
	if err != nil {
		println(err.Error())
		return
	}
	userDB, err := db.GetAppUserByName(con, userName)
	if err == nil && userDB.Pass == common.HashPassword(userPass, userDB.Salt) {
		claims := jwt.MapClaims{
			"ID":   25,
			"Name": "Stumpy",
		}
		jwt, err := common.SignedJWT(claims)
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
			SameSite: http.SameSiteLaxMode,
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
	http.SetCookie(w, existingCookie)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

// Check if user JWT exists, is valid, and user exists in database
func loggedInUser(r *http.Request) *string {
	existingCookie, err := r.Cookie(authCookieName)
	if err != nil {
		return nil
	}
	token, tokenValidity := common.ParseJWT(existingCookie.Value)

	name, ok := token["Name"].(string)
	if !ok {
		return nil
	}

	con, err := db.OpenDB()
	existingUser, _ := db.GetAppUserByName(con, name)
	if err == nil && tokenValidity && existingUser.Name == name {
		return &name
	} else {
		return nil
	}
}
