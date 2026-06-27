package db

import (
	"SA/lib/common"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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

type StreamAssistantDB struct {
	Con  *sql.DB
	Path string
}

// Opens Database for the app
func GetConnection() StreamAssistantDB {
	var db StreamAssistantDB
	db.Path = "./SA.db"

	con, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", db.Path))
	db.Con = con

	if err != nil {
		log.Println("Failed to open DB")
		log.Panic(err.Error())
	}
	return db
}

func (db *StreamAssistantDB) InitDatabase() {
	var err error
	var sqlQuery string
	var row string

	// Create table with app users
	sqlQuery = fmt.Sprintf(`
		SELECT name FROM sqlite_schema
		WHERE type ='table' AND name 
		NOT LIKE 'sqlite_' AND name = '%s';`, userTableName)
	db.Con.QueryRow(sqlQuery).Scan(&row)
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
		_, err = db.Con.Exec(sqlQuery)
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
	db.Con.QueryRow(sqlQuery).Scan(&row)
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
		_, err = db.Con.Exec(sqlQuery)
		if err != nil {
			log.Println("Failed to create DB")
			log.Println(err.Error())
			return
		}
		log.Printf("Created table %s", twitchTableName)
	}
}

// Adds App user
func (db *StreamAssistantDB) AddAppUser(user AppUser) error {
	existingUser, _ := db.GetAppUserByName(user.Name)
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
	_, err := db.Con.Exec(sqlQuery)
	if err != nil {
		log.Println("Failed to create user")
		log.Println(err.Error())
		return err
	}
	log.Printf("Added user %s to %s", user.Name, userTableName)
	return nil
}

// Adds Twitch User to the database
func (db *StreamAssistantDB) AddTwitchUser(user TwitchUser) error {
	existingUser, _ := db.GetTwitchUserByID(user.UserID)
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
	_, err := db.Con.Exec(sqlQuery)
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

func (db *StreamAssistantDB) AddOrReplaceTwitchUser(user TwitchUser) error {
	err := db.AddTwitchUser(user)
	if err != nil {
		var sqlErr *sqlite3.Error
		errors.As(err, &sqlErr)
		println(err.Error())
		if sqlErr.ExtendedCode() == sqlite3.CONSTRAINT_PRIMARYKEY {
			tx, err := db.Con.Begin()
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
func (db *StreamAssistantDB) GetAppUserByName(userName string) (AppUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE Name='%s'", userTableName, userName)
	result := db.Con.QueryRow(sqlQuery)
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
func (db *StreamAssistantDB) GetAppUserChannels(userName string) (ChannelPerm, error) {
	sqlQuery := fmt.Sprintf("SELECT Channels FROM %s WHERE Name='%s'", userTableName, userName)
	result := db.Con.QueryRow(sqlQuery)
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

func (db *StreamAssistantDB) ModifyAppUserChannel(userName string, channelName string, action ChannelAction) error {
	var channelsJSON []byte
	var err error
	channels, err := db.GetUserAccessibleChannels(userName)
	logMessage := "%s channel perms for %s to %s"
	sqlQuery := "UPDATE %s SET %s = '%s' WHERE Name='%s'"

	switch action.ActionType {
	case Add:
		if _, ok := channels[channelName]; ok {
			log.Printf("Add failed: User %s already exists", channelName)
			return fmt.Errorf("Add failed: User %s already exists", channelName)
		}
		channels[channelName] = action.PermLevel
		channelsJSON, err = json.Marshal(&channels)
		println(string(channelsJSON))
		sqlQuery = fmt.Sprintf(sqlQuery, userTableName, "Channels", channelsJSON, userName)
		logMessage = fmt.Sprintf(logMessage, "Added", channelName, userName)
	case Modify:
		if _, ok := channels[channelName]; !ok {
			log.Printf("Modify failed: User %s doesn't exist", channelName)
			return fmt.Errorf("Modify failed: User %s doesn't exist", channelName)
		}
		channels[channelName] = action.PermLevel
		channelsJSON, err = json.Marshal(&channels)
		sqlQuery = fmt.Sprintf(sqlQuery, userTableName, "Channels", channelsJSON, userName)
		logMessage = fmt.Sprintf(logMessage, "Modified", channelName, userName)
	case Remove:
		if _, ok := channels[channelName]; !ok {
			log.Printf("Remove failed: User %s doesn't exist", channelName)
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
	_, err = db.Con.Exec(sqlQuery)
	if err != nil {
		return err
	}
	log.Println(logMessage)
	return nil
}

// Returns the App user by searching for the specified name
func (db *StreamAssistantDB) GetTwitchUserByID(id string) (TwitchUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE UserID='%s'", twitchTableName, id)
	result := db.Con.QueryRow(sqlQuery)
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

func (db *StreamAssistantDB) GetTwitchUserByName(name string) (*TwitchUser, error) {
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE UserName='%s'", twitchTableName, name)
	result := db.Con.QueryRow(sqlQuery)
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

func (db *StreamAssistantDB) TwitchNameToID(name string) (*string, error) {
	if user, err := db.GetTwitchUserByName(name); err == nil {
		return &user.UserID, nil
	} else {
		return nil, err
	}
}

func (db *StreamAssistantDB) TwitchIDNameToName(ID string) (*string, error) {
	if user, err := db.GetTwitchUserByID(ID); err == nil {
		return &user.UserName, nil
	} else {
		return nil, err
	}
}

func (db *StreamAssistantDB) UpdateTwitchUserAccessTokenByName(userName string, token string) error {
	var err error
	sqlQuery := "UPDATE %s SET %s = '%s' WHERE UserName='%s'"
	sqlQuery = fmt.Sprintf(sqlQuery, twitchTableName, "AccessToken", token, userName)
	_, err = db.Con.Exec(sqlQuery)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (db *StreamAssistantDB) UpdateTwitchUserAccessTokenByID(ID string, token string) error {
	var err error
	sqlQuery := "UPDATE %s SET %s = '%s' WHERE UserID='%s'"
	sqlQuery = fmt.Sprintf(sqlQuery, twitchTableName, "AccessToken", token, ID)
	_, err = db.Con.Exec(sqlQuery)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (db *StreamAssistantDB) GetUserAccessibleChannels(appUserName string) (ChannelPerm, error) {
	if user, err := db.GetAppUserByName(appUserName); err != nil {
		return nil, err
	} else {
		return user.Channels, nil
	}
}

func (db *StreamAssistantDB) IsActionAllowedForUser(appUserName string, channelName string, neededAccessLevel PermLevel) bool {
	if channelMap, err := db.GetUserAccessibleChannels(appUserName); err == nil {
		if accessLevel, ok := channelMap[channelName]; ok {
			if accessLevel >= neededAccessLevel {
				return true
			}
		}
	}
	return false
}
