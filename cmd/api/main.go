package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/inodinwetrust10/mumbleBackend/internal/database"
	"github.com/inodinwetrust10/mumbleBackend/internal/server"
)

func main() {
	godotenv.Load()
	addr := fmt.Sprintf(":%d", os.Getenv("PORT"))
	database.Migrate()
	userDB, err := database.NewPostgresUser()
	if err != nil {
		log.Println(err)
	}
	messageDB, err := database.NewPostgresMessage()
	if err != nil {
		log.Println(err)
	}
	ser := server.NewServer(addr, userDB, messageDB)
	ser.Run()
}
