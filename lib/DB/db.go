package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	_ "github.com/ncruces/go-sqlite3/driver"
)

const userTableName = "users"

type AppUser struct {
	Name  string
	Pass  string
	Salt  string
	Admin bool
}

const twitchTableName = "twitchUsers"

type TwitchUser struct {
	TW_user_id       string
	TW_user_name     string
	TW_token_access  string
	TW_token_refresh string
	TW_token_expiry  int64
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
			tw_user_id TEXT PRIMARY KEY,
			tw_user_name TEXT,
			tw_token_access TEXT,
			tw_token_refresh TEXT,
			tw_token_expiry UInt64
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
	if cmp.Equal(existingUser, user) {
		log.Printf("User %s in table %s already exists", user.Name, userTableName)
		return nil
	}
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
	existingUser, _ := GetTwitchUserByID(db, user.TW_user_id)
	if cmp.Equal(existingUser, user) {
		log.Printf("User %s in table %s already exists", user.TW_user_name, twitchTableName)
		return nil
	}
	sqlQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', '%s', %d)",
		twitchTableName,
		user.TW_user_id,
		user.TW_user_name,
		user.TW_token_access,
		user.TW_token_refresh,
		user.TW_token_expiry,
	)
	_, err := db.Exec(sqlQuery)
	if err != nil {
		log.Println("Failed to create user")
		log.Println(err.Error())
		return err
	}
	log.Printf("Added user %s to %s", user.TW_user_name, twitchTableName)
	return nil
}

func GetAppUserByName(db *sql.DB, id string) (AppUser, error) {
	var err error
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE Name='%s'", userTableName, id)
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
		log.Println("Failed to get user")
		log.Println(err.Error())
		return AppUser{}, err
	}
	return u, nil
}

func GetTwitchUserByID(db *sql.DB, id string) (TwitchUser, error) {
	var err error
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE tw_user_id='%s'", twitchTableName, id)
	err = db.Ping()
	if err != nil {
		log.Println("Failed to ping")
		log.Println(err.Error())
		return TwitchUser{}, err
	}
	result := db.QueryRow(sqlQuery)
	u := TwitchUser{}
	err = result.Scan(
		&u.TW_user_id,
		&u.TW_user_name,
		&u.TW_token_access,
		&u.TW_token_refresh,
		&u.TW_token_expiry,
	)
	if err != nil {
		log.Println("Failed to get user")
		log.Println(err.Error())
		return TwitchUser{}, err
	}
	return u, nil
}
