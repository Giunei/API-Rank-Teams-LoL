package main

import (
	"APIRankLolV2/internal/infra/riot"
	"log"
	"os"

	"APIRankLolV2/configs"
	httpHandler "APIRankLolV2/internal/adapter/http"
	"APIRankLolV2/internal/adapter/repository"
	"APIRankLolV2/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.dev")
	apiKey := os.Getenv("RIOT_API_KEY")

	db, err := configs.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	repo := repository.NewTeamRepository(db)
	riotClient := riot.NewRiotClient(apiKey)
	svc := usecase.NewTeamService(repo, riotClient)

	r := gin.Default()
	httpHandler.RegisterTeamRoutes(r, svc)
	r.Run(":8080")
}
