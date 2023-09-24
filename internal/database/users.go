package database

import (
	"errors"
	"time"
)

type RevokeToken struct {
	ID   string `json:"id"`
	Time string `json:"time"`
}

type User struct {
	ID          int         `json:"id"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	RevokeToken RevokeToken `json:"revoke_token"`
	IsChirpyRed bool        `json:"is_chirpy_red"`
}

var ErrAlreadyExists = errors.New("user already exist")

func (db *DB) CreateUser(email string, hashedPassword string) (User, error) {
	if _, err := db.GetUserByEmail(email); !errors.Is(err, ErrNotExist) {
		return User{}, ErrAlreadyExists
	}
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	id := len(dbStruct.Users) + 1
	user := User{
		ID:          id,
		Email:       email,
		Password:    hashedPassword,
		IsChirpyRed: false,
	}
	dbStruct.Users[id] = user
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (db *DB) GetUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotExist
}

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	updatedUser := User{
		ID:       id,
		Email:    email,
		Password: password,
	}
	dbStruct.Users[id] = updatedUser
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	return updatedUser, nil
}

func (db *DB) CreateRevokeToken(id int, time time.Time, tokenID string) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, err := db.GetUser(id)
	if err != nil {
		return User{}, ErrNotExist
	}

	user.RevokeToken = RevokeToken{
		ID:   tokenID,
		Time: time.UTC().String(),
	}
	dbStruct.Users[id] = user
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpdateWebhook(id int) (User, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, err := db.GetUser(id)
	if err != nil {
		return User{}, ErrNotExist
	}
	user.IsChirpyRed = true
	dbStruct.Users[id] = user
	err = db.writeDB(dbStruct)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
