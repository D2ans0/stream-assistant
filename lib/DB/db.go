package db

import (
	"SA/lib/common"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	_ "github.com/ncruces/go-sqlite3/driver"
)

const userTableName = "users"

// AppUser properties
type AppUser struct {
	Name  string // Username used for signing in
	Pass  string // Plain-text password only used for user creation, afterwars a hashed and salted version is returned
	Salt  string // Salt is generated on user creation, use GetAppUserByName() to get actual salt
	Admin bool   // Whether the user has application admin permisions
}

const twitchTableName = "twitchUsers"

type TwitchUser struct {
	UserID            string
	UserName          string
	AccessToken       string
	RefreshToken      string
	AccessTokenExpiry int64
}

func InitDatabase() {
	var err error
	var sqlQuery string
	var row string
	db, _ := OpenDB()

	// Create table with app users
	sqlQuery = fmt.Sprintf(`
		SELECT name FROM sqlite_schema
		WHERE type ='table' AND name 
		NOT LIKE 'sqlite_' AND name = '%s';`, userTableName)
	db.QueryRow(sqlQuery).Scan(&row)
	if row != "" {
		log.Printf("Table %s already exists, skipping...", userTableName)
	} else {
		sqlQuery = fmt.Sprintf(`
		CREATE TABLE %s (
			Name TEXT PRIMARY KEY,
			Pass TEXT,
			Salt TEXT,
			Admin BOOL
		);`, userTableName)
		_, err = db.Exec(sqlQuery)
		if err != nil {
			log.Println("Failed to create DB")
			log.Println(err.Error())
			return
		}
		log.Printf("Created table %s", userTableName)
	}

	// Create table with twitch users we have access to
	sqlQuery = fmt.Sprintf(`
		SELECT name FROM sqlite_schema
		WHERE type ='table' AND name 
		NOT LIKE 'sqlite_' AND name = '%s';`, twitchTableName)
	db.QueryRow(sqlQuery).Scan(&row)
	if row != "" {
		log.Printf("Table %s already exists, skipping...", twitchTableName)
	} else {
		sqlQuery = fmt.Sprintf(`
		CREATE TABLE %s (
			UserID TEXT PRIMARY KEY,
			UserName TEXT,
			AccessToken TEXT,
			RefreshToken TEXT,
			AccessTokenExpiry UInt64
		);`, twitchTableName)
		_, err = db.Exec(sqlQuery)
		if err != nil {
			log.Println("Failed to create DB")
			log.Println(err.Error())
			return
		}
		log.Printf("Created table %s", twitchTableName)
	}
	db.Close()

}

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:SA.db")
	if err != nil {
		log.Println("Failed to open DB")
		log.Println(err.Error())
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println("Failed to ping DB")
		log.Println(err.Error())
		return nil, err
	}
	return db, err
}

func AddAppUser(db *sql.DB, user AppUser) error {
	existingUser, _ := GetAppUserByName(db, user.Name)
	if cmp.Equal(existingUser.Name, user.Name) {
		log.Printf("User %s in table %s already exists...", user.Name, userTableName)
		return nil
	}
	user.Salt = common.GenerateSalt()
	user.Pass = common.HashPassword(user.Pass, user.Salt)
	sqlQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', '%t')",
		userTableName,
		user.Name,
		user.Pass,
		user.Salt,
		user.Admin,
	)
	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Println("Failed to create user")
		log.Println(err.Error())
		return err
	}
	log.Printf("Added user %s to %s", user.Name, userTableName)
	return nil
}

func AddTwitchUser(db *sql.DB, user TwitchUser) error {
	existingUser, _ := GetTwitchUserByID(db, user.UserID)
	if cmp.Equal(existingUser, user) {
		log.Printf("User %s in table %s already exists...", user.UserName, twitchTableName)
		return nil
	}
	sqlQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', '%s', %d)",
		twitchTableName,
		user.UserID,
		user.UserName,
		user.AccessToken,
		user.RefreshToken,
		user.AccessTokenExpiry,
	)
	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Println("Failed to create user")
		log.Println(err.Error())
		return err
	}
	log.Printf("Added user %s to %s", user.UserName, twitchTableName)
	return nil
}

func GetAppUserByName(db *sql.DB, userName string) (AppUser, error) {
	var err error
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE Name='%s'", userTableName, userName)
	err = db.Ping()
	if err != nil {
		log.Println("Failed to ping")
		log.Println(err.Error())
		return AppUser{}, err
	}
	result := db.QueryRow(sqlQuery)
	u := AppUser{}
	err = result.Scan(
		&u.Name,
		&u.Pass,
		&u.Salt,
		&u.Admin,
	)
	if err != nil {
		return AppUser{}, err
	}
	return u, nil
}

func GetTwitchUserByID(db *sql.DB, id string) (TwitchUser, error) {
	var err error
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE UserID='%s'", twitchTableName, id)
	err = db.Ping()
	if err != nil {
		log.Println("Failed to ping")
		log.Println(err.Error())
		return TwitchUser{}, err
	}
	result := db.QueryRow(sqlQuery)
	u := TwitchUser{}
	err = result.Scan(
		&u.UserID,
		&u.UserName,
		&u.AccessToken,
		&u.RefreshToken,
		&u.AccessTokenExpiry,
	)
	if err != nil {
		return TwitchUser{}, err
	}
	return u, nil
}
