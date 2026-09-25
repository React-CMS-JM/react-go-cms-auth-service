package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"react-go-cms-auth-service/internal/auth"
	"react-go-cms-auth-service/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("8081")
	db, err := platform.OpenDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database: %v", err)
	}

	svc := auth.New(db, cfg)
	h := auth.NewHandler(svc, cfg.JWTSecret, cfg.JWTIssuer)
	mux := http.NewServeMux()
	h.Register(mux)

	addr := ":" + cfg.Port
	log.Printf("react-go-cms-auth-service listening on %s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           platform.CORS(cfg.CORSOrigins, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
