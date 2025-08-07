package http

import (
	"APIRankLolV2/internal/domain"
	"APIRankLolV2/internal/usecase"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterTeamRoutes(router *gin.Engine, svc *usecase.TeamService) {
	fmt.Println("registrando rotas de team")
	router.POST("/teams", func(c *gin.Context) {
		ctx := context.Background()
		var team domain.Team

		if err := c.ShouldBindJSON(&team); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "details": err.Error()})
			return
		}

		if err := svc.CreateTeam(ctx, team); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Team created successfully"})
	})

	router.POST("/teams/:id/players", func(c *gin.Context) {
		ctx := context.Background()
		teamIDParam := c.Param("id")
		var player domain.Player

		if err := c.ShouldBindJSON(&player); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player body", "details": err.Error()})
			return
		}

		if err := svc.AddPlayerToTeam(ctx, teamIDParam, player); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Player added successfully"})
	})

	router.GET("/players", func(c *gin.Context) {
		ctx := context.Background()
		gamerName := c.Query("gamer_name")
		if gamerName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gamer_name query parameter is required"})
			return
		}

		players, err := svc.GetPlayersByGamerName(ctx, gamerName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, players)
	})

	router.GET("/players/:id/winrate", func(c *gin.Context) {
		ctx := context.Background()
		playerID := c.Param("id")
		matchCount := c.DefaultQuery("count", "10")
		typeFilter := c.Query("type")
		queueFilter := c.Query("queue")

		result, err := svc.CalculateWinRate(ctx, playerID, matchCount, typeFilter, queueFilter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"winrate": result})
	})

	router.GET("/teams/:id/winrate", func(c *gin.Context) {
		ctx := context.Background()
		teamID := c.Param("id")
		matchCount := c.DefaultQuery("count", "10")
		typeFilter := c.Query("type")
		queueFilter := c.Query("queue")

		result, err := svc.CalculateWinRateTeam(ctx, teamID, matchCount, typeFilter, queueFilter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"winrate": result})
	})
}
