package db

import (
	"SA/lib/common"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	_ "github.com/ncruces/go-sqlite3/driver"
)

const userTableName = "users"

// Level of permissions available to user
type PermLevel int
type ChannelPerm map[string]PermLevel

const (
	User      PermLevel = iota // Can control items directly related to the broadcast
	Moderator                  // Grants moderator access to the user
	Admin                      // Can set any available options
	Owner                      // Same as admin, but permissions cannot be stripped by admins
)

// AppUser properties
type AppUser struct {
	Name        string      // Username used for signing in
	Pass        string      // Plain-text password only used for user creation, afterwars a hashed and salted version is returned
	Salt        string      // Salt is generated on user creation, use GetAppUserByName() to get actual salt
	Permissions PermLevel   // Application permission level
	Channels    ChannelPerm // Channels the user can access
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
			PermsLevel INT,
			Channels TEXT
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

// Opens Database for the app
func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:SA.db")

	if err != nil {
		log.Println("Failed to open DB")
		log.Println(err.Error())
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Println("Failed to ping DB")
		log.Println(err.Error())
		return nil, err
	}
	return db, err
}

// Adds App user
func AddAppUser(db *sql.DB, user AppUser) error {
	existingUser, _ := GetAppUserByName(db, user.Name)
	if cmp.Equal(existingUser.Name, user.Name) {
		log.Printf("User %s in table %s already exists...", user.Name, userTableName)
		return nil
	}
	user.Salt = common.GenerateSalt()
	user.Pass = common.HashPassword(user.Pass, user.Salt)
	channelsJSON, _ := json.Marshal(user.Channels)
	sqlQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', '%d', '%s')",
		userTableName,
		user.Name,
		user.Pass,
		user.Salt,
		user.Permissions,
		channelsJSON,
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

// Adds Twitch User to the database
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
		log.Printf("Failed to create user %s", user.UserName)
		log.Println(err.Error())
		return err
	}
	log.Printf("Added user %s to %s", user.UserName, twitchTableName)
	return nil
}

// Returns the App user by searching for the specified name
func GetAppUserByName(db *sql.DB, userName string) (AppUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE Name='%s'", userTableName, userName)
	if err := db.Ping(); err != nil {
		log.Println("Failed to contact DB")
		return AppUser{}, err
	}
	result := db.QueryRow(sqlQuery)
	u := AppUser{}
	var rawChannelsJSON []byte
	if err := result.Scan(
		&u.Name,
		&u.Pass,
		&u.Salt,
		&u.Permissions,
		&rawChannelsJSON,
	); err != nil {
		return AppUser{}, err
	}

	if err := json.Unmarshal(rawChannelsJSON, &u.Channels); err != nil {
		return AppUser{}, err
	}
	return u, nil
}

// Returns the App user by searching for the specified name
func GetTwitchUserByID(db *sql.DB, id string) (TwitchUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE UserID='%s'", twitchTableName, id)
	if err := db.Ping(); err != nil {
		log.Println("Failed to ping")
		log.Println(err.Error())
		return TwitchUser{}, err
	}
	result := db.QueryRow(sqlQuery)
	u := TwitchUser{}
	if err := result.Scan(
		&u.UserID,
		&u.UserName,
		&u.AccessToken,
		&u.RefreshToken,
		&u.AccessTokenExpiry,
	); err != nil {
		return TwitchUser{}, err
	}
	return u, nil
}
