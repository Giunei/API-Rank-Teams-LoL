package domain

import "context"

type TeamRepository interface {
	SaveTeam(ctx context.Context, team Team) (int64, error)
	SavePlayer(ctx context.Context, p Player) error
	UpdatePlayer(ctx context.Context, p Player) error
	FindPlayersByGamerName(ctx context.Context, gamerName string) ([]Player, error)
	FindAllPlayersByTeamID(ctx context.Context, teamID int64) ([]Player, error)
	FindPlayerByID(ctx context.Context, playerID int64) (Player, error)
	FindPlayersByTeamID(ctx context.Context, teamID int64) ([]Player, error)
}
