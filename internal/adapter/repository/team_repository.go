package repository

import (
	"context"
	_ "database/sql"
	"fmt"

	"APIRankLolV2/internal/domain"
	"github.com/jmoiron/sqlx"
)

type TeamRepository struct {
	Db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{Db: db}
}

func (r *TeamRepository) SaveTeam(ctx context.Context, team domain.Team) (int64, error) {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	fmt.Println("inserindo...")

	var teamID int64
	err = tx.QueryRowContext(ctx, "INSERT INTO team (name) VALUES ($1) RETURNING id", team.Name).Scan(&teamID)
	if err != nil {
		return 0, err
	}
	return teamID, tx.Commit()
}

func (r *TeamRepository) SavePlayer(ctx context.Context, p domain.Player) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fmt.Println("inserindo...")

	if p.ID == 0 {
		_, err = tx.ExecContext(ctx, "INSERT INTO player (gamer_name, tag_line, team_id, puuid) VALUES ($1, $2, $3, $4)",
			p.GamerName, p.TagLine, p.TeamID, p.Puuid)
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO player (id, gamer_name, tag_line, team_id, puuid) VALUES ($1, $2, $3, $4, $5)",
			p.ID, p.GamerName, p.TagLine, p.TeamID, p.Puuid)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TeamRepository) UpdatePlayer(ctx context.Context, p domain.Player) error {
	if p.ID == 0 {
		return fmt.Errorf("ID do player não pode ser zero")
	}

	query := `
		UPDATE player
		SET gamer_name = :gamer_name,
		    tag_line = :tag_line,
		    puuid = :puuid
		WHERE id = :id
	`

	_, err := r.Db.NamedExecContext(ctx, query, p)
	if err != nil {
		return fmt.Errorf("erro ao atualizar player: %w", err)
	}

	return nil
}

func (r *TeamRepository) FindPlayersByGamerName(ctx context.Context, gamerName string) ([]domain.Player, error) {
	var players []domain.Player
	query := `
		SELECT id, gamer_name, tag_line, team_id
		FROM player
		WHERE gamer_name ILIKE '%' || $1 || '%'
	`
	err := r.Db.SelectContext(ctx, &players, query, gamerName)
	return players, err
}

func (r *TeamRepository) FindPlayersByTeamID(ctx context.Context, teamID int64) ([]domain.Player, error) {
	var players []domain.Player
	query := `
		SELECT id, gamer_name, tag_line, team_id, puuid
		FROM player
		WHERE team_id = $1
	`
	err := r.Db.SelectContext(ctx, &players, query, teamID)
	return players, err
}

func (r *TeamRepository) FindPlayerByID(ctx context.Context, playerID int64) (domain.Player, error) {
	var player domain.Player
	query := `
		SELECT id, gamer_name, tag_line, team_id, COALESCE(puuid, '') AS puuid
		FROM player
		WHERE id = $1
	`
	err := r.Db.GetContext(ctx, &player, query, playerID)

	return player, err
}
