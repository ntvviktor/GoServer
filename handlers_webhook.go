package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (apiConfig *apiConfig) handleWebhook(w http.ResponseWriter, req *http.Request) {
	apiString := req.Header.Get("Authorization")
	apiKey, canCut := strings.CutPrefix(apiString, "ApiKey ")
	fmt.Println(apiKey)
	if !canCut || apiKey != apiConfig.apiKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized request")
		return
	}

	type parameter struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(req.Body)
	param := parameter{}
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed request")
		return
	}
	if param.Event == "user.upgraded" {
		user, err := apiConfig.DB.UpdateWebhook(param.Data.UserID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithJSON(w, http.StatusOK, user)
		return
	}
	respondWithJSON(w, http.StatusOK, "")
	return
}
