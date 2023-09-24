package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github/ntvviktor/GoServer/internal/database"
	"log"
	"net/http"
	"os"
)

type apiConfig struct {
	fileServerHits int
	DB             *database.DB
	jwtSecret      string
	apiKey         string
}

func main() {
	r := chi.NewRouter()
	corsMux := MiddlewareCors(r)
	db, err := database.NewDB("database.json")
	if err != nil {
		fmt.Println("Database gets error")
		log.Fatal(err)
	}

	_ = godotenv.Load("local.env")
	jwtSecret := os.Getenv("JWT_SECRET")
	apiKey := os.Getenv("API_KEY")
	apiCfg := apiConfig{
		fileServerHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
		apiKey:         apiKey,
	}
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	// Split the /admin and /api router
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricNumber())

	apiRouter := chi.NewRouter()
	apiRouter.Handle("/reset", apiCfg.resetMetric())
	apiRouter.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	apiRouter.Post("/users", apiCfg.createUser)
	apiRouter.Put("/users", apiCfg.validateUser)
	apiRouter.Post("/login", apiCfg.authenticateUser)
	apiRouter.Post("/refresh", apiCfg.postRefreshToken)
	apiRouter.Post("/revoke", apiCfg.postRevokeToken)
	apiRouter.Post("/chirps", apiCfg.handlePostChirps)
	apiRouter.Post("/polka/webhooks", apiCfg.handleWebhook)
	apiRouter.Get("/chirps", apiCfg.handleGetChirps)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.getChirpsById)
	apiRouter.Delete("/chirps/{chirpID}", apiCfg.handleDeleteChirp)

	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	http.ListenAndServe(":8080", corsMux)
}
