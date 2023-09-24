package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github/ntvviktor/GoServer/internal/auth"
	"github/ntvviktor/GoServer/internal/database"
	"log"
	"net/http"
	"sort"
	"strconv"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (apiConfig *apiConfig) handlePostChirps(w http.ResponseWriter, req *http.Request) {
	tokenString, canCut := getTokenFromHeader(req)
	if !canCut {
		respondWithError(w, http.StatusBadRequest, "Malformed  JSON token")
		return
	}
	authorID, _, err := auth.ValidateJWT(tokenString, apiConfig.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid Token")
		return
	}
	decoder := json.NewDecoder(req.Body)
	type parameter struct {
		Body string `json:"body"`
	}
	param := parameter{}
	err = decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed  JSON token")
		return
	}
	chirp, err := apiConfig.DB.CreateChirp(param.Body, authorID)
	type response struct {
		Chirp
	}
	respondWithJSON(w, http.StatusCreated, response{
		Chirp{
			ID:       chirp.ID,
			Body:     param.Body,
			AuthorID: chirp.AuthorID,
		},
	})
}

func (apiConfig *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {
	var responseChirps []database.Chirp
	var authorID int
	authorIDString := req.URL.Query().Get("author_id")

	chirps, err := apiConfig.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if authorIDString == "" {
		responseChirps = chirps
	} else {
		authorID, err = strconv.Atoi(authorIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Internal error")
			return
		}
		for _, v := range chirps {
			if v.AuthorID == authorID {
				chirp := database.Chirp{
					ID:       v.ID,
					Body:     v.Body,
					AuthorID: v.AuthorID,
				}
				responseChirps = append(responseChirps, chirp)
			}
		}
	}

	sortCriteria := req.URL.Query().Get("sort")
	if sortCriteria == "desc" {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].ID > responseChirps[j].ID
		})
	} else {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].ID < responseChirps[j].ID
		})
	}
	respondWithJSON(w, http.StatusOK, responseChirps)
}

func (apiConfig *apiConfig) handleDeleteChirp(w http.ResponseWriter, req *http.Request) {
	tokenString, canCut := getTokenFromHeader(req)
	if !canCut {
		respondWithError(w, http.StatusBadRequest, "Malformed  JSON token")
		return
	}
	authorID, _, err := auth.ValidateJWT(tokenString, apiConfig.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid Token")
		return
	}
	paramID := chi.URLParam(req, "chirpID")
	chirpID, err := strconv.Atoi(paramID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed URL request")
		return
	}
	chirp, err := apiConfig.DB.DeleteChirp(chirpID, authorID)
	if err != nil {
		respondWithJSON(w, http.StatusForbidden, "Cannot delete")
		return
	}
	respondWithJSON(w, http.StatusOK, chirp)
}

func (apiConfig *apiConfig) getChirpsById(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "chirpID")
	chirps, err := apiConfig.DB.GetChirps()
	if err != nil {
		http.Error(w, "Internal Error", 500)
		log.Fatal(err)
	}
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})
	chirpID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Internal Error", 500)
		log.Fatal(err)
	}
	for _, v := range chirps {
		if chirpID == v.ID {
			w.WriteHeader(200)
			dat, _ := json.Marshal(v)
			w.Write(dat)
			fmt.Println(string(dat))
			return
		}
	}
	w.WriteHeader(404)
}
