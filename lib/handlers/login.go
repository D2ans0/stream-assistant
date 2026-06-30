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
const userCookieName = "User"

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
	con := db.GetConnection()
	defer con.Con.Close()
	userDB, err := con.GetAppUserByName(userName)
	if err == nil && userDB.Pass == common.HashPassword(userPass, userDB.Salt) {
		claims := jwt.MapClaims{
			"Name":            userDB.Name,
			"PermissionLevel": userDB.Permissions,
		}
		jwt, err := common.SignedJWT(claims)
		if err != nil {
			log.Println(err.Error())
			return
		}
		cookieAuth := http.Cookie{
			Name:     authCookieName,
			Value:    jwt,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}
		// intended to be used from JS.
		// Server will still check against cookieAuth instead
		cookieUser := http.Cookie{
			Name:     userCookieName,
			Value:    fmt.Sprintf("%s:%d", userDB.Name, userDB.Permissions),
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}
		if rememberMe == "on" {
			cookieUser.Expires = time.Now().Add(time.Hour * 365 * 24)
			cookieAuth.Expires = time.Now().Add(time.Hour * 365 * 24)

		}
		http.SetCookie(w, &cookieUser)
		http.SetCookie(w, &cookieAuth)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Error(w,
			fmt.Sprintf("%d: Unauthorized", http.StatusUnauthorized),
			http.StatusUnauthorized,
		)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	existingAuthCookie, err := r.Cookie(authCookieName)
	if err == nil {
		existingAuthCookie.Value = ""
		existingAuthCookie.Expires = time.Unix(0, 0)
		http.SetCookie(w, existingAuthCookie)
	}
	existingUserCookie, err := r.Cookie(userCookieName)
	if err == nil {
		existingUserCookie.Value = ""
		existingUserCookie.Expires = time.Unix(0, 0)
		http.SetCookie(w, existingUserCookie)
	}
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

	con := db.GetConnection()
	defer con.Con.Close()
	existingUser, err := con.GetAppUserByName(name)
	if err == nil && tokenValidity && existingUser.Name == name {
		return &name
	} else {
		return nil
	}
}
