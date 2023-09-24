package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github/ntvviktor/GoServer/internal/auth"
	"github/ntvviktor/GoServer/internal/database"
	"net/http"
	"strings"
	"time"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     string `json:"-"`
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
}

func (apiConfig *apiConfig) createUser(w http.ResponseWriter, req *http.Request) {
	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		//IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(req.Body)
	param := parameter{}
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed JSON Request")
		return
	}

	hashedPassword, _ := auth.HashUserPassword(param.Password)
	user, err := apiConfig.DB.CreateUser(param.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyExists) {
			respondWithError(w, http.StatusConflict, "User already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, 201, response{
		ID:    user.ID,
		Email: user.Email,
		//IsChirpyRed: user.IsChirpyRed,
	})
}

func (apiConfig *apiConfig) authenticateUser(w http.ResponseWriter, req *http.Request) {
	type parameter struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		ExpiresTime int    `json:"expires_in_seconds"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(req.Body)
	param := parameter{}
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed JSON Request")
		return
	}

	authUser, err := apiConfig.DB.GetUserByEmail(param.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "User not found!")
			return
		}
		http.Error(w, "Internal Error", 500)
		return
	}
	err = auth.AuthenticateUser(param.Password, authUser.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Wrong Password")
		return
	}
	accessToken, err := auth.GenerateJWT(authUser.ID, apiConfig.jwtSecret, auth.AccessToken)
	refreshToken, err := auth.GenerateJWT(authUser.ID, apiConfig.jwtSecret, auth.RefreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating JWT token")
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		User{
			ID:           authUser.ID,
			Email:        authUser.Email,
			IsChirpyRed:  authUser.IsChirpyRed,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

func (apiConfig *apiConfig) validateUser(w http.ResponseWriter, req *http.Request) {
	tokenString, canCut := getTokenFromHeader(req)
	if !canCut {
		respondWithError(w, http.StatusUnauthorized, "Malformed request token string")
		return
	}

	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Malformed request, token is not valid!")
		return
	}

	validID, issuer, err := auth.ValidateJWT(tokenString, apiConfig.jwtSecret)
	if err != nil || issuer == "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Cannot validate token")
		return
	}

	hashedPassword, err := auth.HashUserPassword(param.Password)
	user, err := apiConfig.DB.UpdateUser(validID, param.Email, hashedPassword)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
	}

	type response struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}
	respondWithJSON(w, 200, response{
		Email: user.Email,
		ID:    user.ID,
	})
}

func (apiConfig *apiConfig) postRefreshToken(w http.ResponseWriter, req *http.Request) {
	tokenString, canCut := getTokenFromHeader(req)
	if !canCut {
		respondWithError(w, http.StatusUnauthorized, "Malformed request token string")
		return
	}
	validID, issuer, err := auth.ValidateJWT(tokenString, apiConfig.jwtSecret)
	if err != nil || issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized JSON Token")
		return
	}

	user, err := apiConfig.DB.GetUser(validID)
	if err != nil || issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusNotFound, "User Not found")
		return
	}
	if user.RevokeToken.ID != "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized JSON Token")
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, 200, response{
		Token: tokenString,
	})
}

func (apiConfig *apiConfig) postRevokeToken(w http.ResponseWriter, req *http.Request) {
	tokenString, canCut := getTokenFromHeader(req)
	if !canCut {
		respondWithError(w, http.StatusUnauthorized, "Malformed request token string")
		return
	}
	validID, issuer, err := auth.ValidateJWT(tokenString, apiConfig.jwtSecret)

	if err != nil || issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized JSON Token")
		return
	}

	user, err := apiConfig.DB.CreateRevokeToken(validID, time.Now(), tokenString)
	if err != nil || issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusNotFound, "User Not found")
		return
	}

	respondWithJSON(w, 200, user.Email)
}

func getTokenFromHeader(req *http.Request) (string, bool) {
	authorizedToken := req.Header.Get("Authorization")
	tokenString, canCut := strings.CutPrefix(authorizedToken, "Bearer ")
	fmt.Println(tokenString)

	return tokenString, canCut
}
