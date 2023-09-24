package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (apiConfig *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		apiConfig.fileServerHits += 1
		next.ServeHTTP(writer, request)
	})
}
func (apiConfig *apiConfig) metricNumber() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := fmt.Sprintf(`<html>
					<body>
						<h1>Welcome, Chirpy Admin</h1>
						<p>Chirpy has been visited %d times!</p>
					</body>
				</html>
`, apiConfig.fileServerHits)
		w.Write([]byte(html))
	}
}
func (apiConfig *apiConfig) resetMetric() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		apiConfig.fileServerHits = 0
		w.Write([]byte("Hits: " + strconv.Itoa(apiConfig.fileServerHits)))
	})
}

func validateRequest(email string) string {
	if len(email) > 140 {
		return "Chirp is too long"
	} else {
		var profane = [3]string{"kerfuffle", "sharbert", "fornax"}
		body := strings.Split(email, " ")
		for _, v := range profane {
			for i, val := range body {
				if strings.ToLower(val) == v {
					body[i] = "****"
				}
			}
		}
		return strings.Join(body, " ")
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	resBody := map[string]string{"error": msg}
	dat, err := json.Marshal(resBody)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	dat, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Cannot decode the payload", 500)
		log.Fatal(err)
	}

	w.Write(dat)
}
