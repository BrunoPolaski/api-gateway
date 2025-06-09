package main

import (
	"log"
	"net/http"

	"github.com/BrunoPolaski/api-gateway/internal/router"
	"github.com/BrunoPolaski/go-logger/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	logger.InitLogger()

	r := router.Init()

	logger.Info("Initializing gateway on port :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
