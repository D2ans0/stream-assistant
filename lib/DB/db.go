package db

import (
	"SA/lib/common"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/google/go-cmp/cmp"
	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
)

var ErrUserNotFound = errors.New("User not found")

const userTableName = "users"

type ChannelPerm map[string]PermLevel
type ChannelActionType int

const (
	Add    ChannelActionType = iota // Adds access to a channel
	Modify                          // Modifies the level of access to a channel
	Remove                          // Removes all access to the channel
)

type ChannelAction struct {
	ActionType ChannelActionType // Add/Modify/Remove
	PermLevel  PermLevel         // Optional for Delete channel action
}

// Level of permissions available to user, same "system" used for both platform perms and channel perms
type PermLevel int

const (
	Undefined PermLevel = iota
	User                // Can control items directly related to the broadcast
	Moderator           // Grants moderator access to the user
	Admin               // Can set any available options
	Owner               // Same as admin, but permissions cannot be stripped by admins
)

// AppUser properties
type AppUser struct {
	Name        string      // Username used for signing in
	Pass        string      // Plain-text password only used for user creation, afterwars a hashed and salted version is returned
	Salt        string      // Salt is generated on user creation, use GetAppUserByName() to get actual salt
	Permissions PermLevel   // Application permission level
	Channels    ChannelPerm // Channels the user can access and the access level
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
			AccessTokenExpiry Int64
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
		var sqlErr *sqlite3.Error
		errors.As(err, &sqlErr)
		if sqlErr.ExtendedCode() == sqlite3.CONSTRAINT_PRIMARYKEY {
			log.Printf("Channel %s already exists!", user.UserName)
		} else {
			log.Println(err.Error())
		}
		return err
	}
	log.Printf("Added user %s to %s", user.UserName, twitchTableName)
	return nil
}

func AddOrReplaceTwitchUser(db *sql.DB, user TwitchUser) error {
	err := AddTwitchUser(db, user)
	if err != nil {
		var sqlErr *sqlite3.Error
		errors.As(err, &sqlErr)
		println(err.Error())
		if sqlErr.ExtendedCode() == sqlite3.CONSTRAINT_PRIMARYKEY {
			tx, err := db.Begin()
			if err != nil {
				println(err.Error())
				return err
			}
			println(fmt.Sprintf("DELETE from %s where UserID=%s", twitchTableName, user.UserID))
			tx.Exec(fmt.Sprintf("DELETE from %s where UserID=%s", twitchTableName, user.UserID))
			sqlQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', '%s', %d)",
				twitchTableName,
				user.UserID,
				user.UserName,
				user.AccessToken,
				user.RefreshToken,
				user.AccessTokenExpiry,
			)
			tx.Exec(sqlQuery)
			err = tx.Commit()
			if err != nil {
				println(err.Error())
				return err
			}
		}
	}
	return nil
}

// Returns the App user by searching for the specified name
func GetAppUserByName(db *sql.DB, userName string) (AppUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE Name='%s'", userTableName, userName)
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

// Returns a map with what channels the user has access to
func GetAppUserChannels(db *sql.DB, userName string) (ChannelPerm, error) {
	sqlQuery := fmt.Sprintf("SELECT Channels FROM %s WHERE Name='%s'", userTableName, userName)
	result := db.QueryRow(sqlQuery)
	var channels ChannelPerm
	var rawChannelsJSON []byte
	if err := result.Scan(&rawChannelsJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(rawChannelsJSON, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

func ModifyAppUserChannel(db *sql.DB, userName string, channelName string, action ChannelAction) error {
	var channelsJSON []byte
	var err error
	channels := make(ChannelPerm)
	logMessage := "%s channel perms for %s to %s"
	sqlQuery := "UPDATE %s SET %s = '%s' WHERE Name='%s'"
	channelsErr := fmt.Errorf("Remove failed: User %s doesn't exist", channelName)
	switch action.ActionType {
	case Add:
		if _, ok := channels[channelName]; ok {
			return fmt.Errorf("Add failed: User %s already exists", channelName)
		}
		channels[channelName] = action.PermLevel
		channelsJSON, err = json.Marshal(&channels)
		println(channelsJSON)
		sqlQuery = fmt.Sprintf(sqlQuery, userTableName, "Channels", channelsJSON, userName)
		logMessage = fmt.Sprintf(logMessage, "Added", channelName, userName)
	case Modify:
		if _, ok := channels[channelName]; !ok || channelsErr != nil {
			return fmt.Errorf("Modify failed: User %s doesn't exist", channelName)
		}
		channels[channelName] = action.PermLevel
		channelsJSON, err = json.Marshal(&channels)
		sqlQuery = fmt.Sprintf(sqlQuery, userTableName, "Channels", channelsJSON, userName)
		logMessage = fmt.Sprintf(logMessage, "Modified", channelName, userName)
	case Remove:
		if _, ok := channels[channelName]; !ok || channelsErr != nil {
			return fmt.Errorf("Remove failed: User %s doesn't exist", channelName)
		}
		delete(channels, channelName)
		channelsJSON, err = json.Marshal(&channels)
		sqlQuery = fmt.Sprintf(sqlQuery, userTableName, "Channels", channelsJSON, userName)
		logMessage = fmt.Sprintf(logMessage, "Removed", channelName, userName)
	default:
		return errors.New("Invalid ActionType")
	}

	if err != nil {
		return err
	}
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return err
	}
	log.Println(logMessage)
	return nil
}

// Returns the App user by searching for the specified name
func GetTwitchUserByID(db *sql.DB, id string) (TwitchUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE UserID='%s'", twitchTableName, id)
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

func GetTwitchUserByName(db *sql.DB, name string) (*TwitchUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE UserName='%s'", twitchTableName, name)
	result := db.QueryRow(sqlQuery)
	u := TwitchUser{}
	if err := result.Scan(
		&u.UserID,
		&u.UserName,
		&u.AccessToken,
		&u.RefreshToken,
		&u.AccessTokenExpiry,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateTwitchUserAccessTokenByName(db *sql.DB, userName string, token string) error {
	var err error
	sqlQuery := "UPDATE %s SET %s = '%s' WHERE UserName='%s'"
	sqlQuery = fmt.Sprintf(sqlQuery, twitchTableName, "AccessToken", token, userName)
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func UpdateTwitchUserAccessTokenByID(db *sql.DB, ID string, token string) error {
	var err error
	sqlQuery := "UPDATE %s SET %s = '%s' WHERE UserID='%s'"
	sqlQuery = fmt.Sprintf(sqlQuery, twitchTableName, "AccessToken", token, ID)
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func GetUserAccessibleChannels(db *sql.DB, appUserName string) (map[string]string, error) {
	users := make(map[string]string)
	if user, err := GetAppUserByName(db, appUserName); err != nil {
		return nil, err
	} else {
		for key, val := range user.Channels {
			users[key] = strconv.Itoa(int(val))
		}
		return users, nil
	}
}
