package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github/ntvviktor/GoServer/internal/database"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type apiConfig struct {
	fileServerHits int
}

func (config *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.fileServerHits += 1
		next.ServeHTTP(writer, request)
	})
}
func (config *apiConfig) metricNumber() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := fmt.Sprintf(`<html>
					<body>
						<h1>Welcome, Chirpy Admin</h1>
						<p>Chirpy has been visited %d times!</p>
					</body>
				</html>
`, config.fileServerHits)
		w.Write([]byte(html))
	}
}
func (config *apiConfig) resetMetric() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		config.fileServerHits = 0
		w.Write([]byte("Hits: " + strconv.Itoa(config.fileServerHits)))
	})
}

func responseWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	resBody := map[string]string{"error": msg}
	dat, err := json.Marshal(resBody)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload string, db *database.DB) {
	w.WriteHeader(code)
	data, err := db.CreateChirp(payload)
	if err != nil {
		fmt.Println(data)
		log.Fatal(err)
		return
	}
	sentData := map[string]interface{}{
		"body": data.Body,
		"id":   data.ID,
	}
	dat, err := json.Marshal(sentData)

	w.Write(dat)
}

func main() {
	r := chi.NewRouter()
	corsMux := middlewareCors(r)
	apiCfg := apiConfig{
		fileServerHits: 0,
	}
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	db, err := database.NewDB("database.json")
	if err != nil {
		fmt.Println("Database gets error")
		log.Fatal(err)
	}
	router := chi.NewRouter()
	router.Handle("/reset", apiCfg.resetMetric())
	router.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricNumber())
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)
	r.Mount("/api", router)
	r.Mount("/admin", adminRouter)
	r.Post("/api/chirps", func(w http.ResponseWriter, req *http.Request) {
		type parameter struct {
			Body string `json:"body"`
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(req.Body)
		param := parameter{}
		err := decoder.Decode(&param)
		if err != nil {
			responseWithError(w, 500, "Something went wrong")
			log.Fatal(err)
			return
		}
		if len(param.Body) > 140 {
			responseWithError(w, 400, "Chirp is too long")
		} else {
			var profane = [3]string{"kerfuffle", "sharbert", "fornax"}
			body := strings.Split(param.Body, " ")
			for _, v := range profane {
				for i, val := range body {
					if strings.ToLower(val) == v {
						body[i] = "****"
					}
				}
			}
			payload := strings.Join(body, " ")
			respondWithJSON(w, 201, payload, db)
		}
	})
	r.Get("/api/chirps", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		chirps, err := db.GetChirps()
		if err != nil {
			http.Error(w, "Internal Error", 500)
			log.Fatal(err)
		}
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
		dat, _ := json.Marshal(chirps)
		w.Write(dat)
	})
	http.ListenAndServe(":8080", corsMux)
}
func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
