package usecase

import (
	"APIRankLolV2/internal/domain"
	"APIRankLolV2/internal/infra/riot"
	"context"
	"fmt"
	"strconv"
)

type TeamService struct {
	repo       domain.TeamRepository
	riotClient *riot.RiotClient
}

func NewTeamService(repo domain.TeamRepository, riotClient *riot.RiotClient) *TeamService {
	return &TeamService{repo: repo, riotClient: riotClient}
}

func (s *TeamService) CreateTeam(ctx context.Context, team domain.Team) error {
	teamId, err := s.repo.SaveTeam(ctx, team)
	if err != nil {
		return err
	}

	for _, p := range team.Players {
		err := s.AddPlayerToTeam(ctx, strconv.FormatInt(teamId, 10), p)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *TeamService) AddPlayerToTeam(ctx context.Context, teamIDParam string, player domain.Player) error {
	teamID, err := strconv.ParseInt(teamIDParam, 10, 64)

	if err != nil {
		fmt.Println("Erro ao converter:", err)
		return err
	}

	account, err := s.riotClient.GetAccountByRiotID(player.GamerName, player.TagLine)
	if err != nil {
		return fmt.Errorf("falha ao buscar PUUID na Riot API: %w", err)
	}
	player.Puuid = account.Puuid
	player.TeamID = teamID

	return s.SavePlayer(ctx, player)
}

func (s *TeamService) SavePlayer(ctx context.Context, player domain.Player) error {
	return s.repo.SavePlayer(ctx, player)
}

func (s *TeamService) GetPlayersByGamerName(ctx context.Context, gamerName string) ([]domain.Player, error) {
	return s.repo.FindPlayersByGamerName(ctx, gamerName)
}

func (s *TeamService) GetPlayerById(ctx context.Context, playerID int64) (domain.Player, error) {
	player, err := s.repo.FindPlayerByID(ctx, playerID)
	if err != nil {
		return player, fmt.Errorf("player not found: %w", err)
	}

	if player.Puuid == "" {
		account, err := s.riotClient.GetAccountByRiotID(player.GamerName, player.TagLine)

		if err != nil {
			return player, err
		}

		player.Puuid = account.Puuid

		err = s.SavePlayer(ctx, player)

		if err != nil {
			return player, err
		}
	}

	return player, nil
}

func (s *TeamService) CalculateWinRateTeam(ctx context.Context, teamID, countStr, typeFilter, queueFilter string) (float64, error) {
	teamIdInt, _ := strconv.ParseInt(teamID, 10, 64)
	players, err := s.repo.FindPlayersByTeamID(ctx, teamIdInt)

	if err != nil {
		return 0, err
	}

	var winrate float64 = 0
	var total float64 = 0

	for _, player := range players {
		winrateValue, err := s.CalculateWinRate(ctx, strconv.FormatInt(player.ID, 10), countStr, typeFilter, queueFilter)
		winrate += winrateValue
		total++

		if err != nil {
			return 0, err
		}
	}

	winrateTeam := winrate / total
	fmt.Printf("Winrate time: %.1f%%\n", winrateTeam)
	return winrateTeam, nil
}

func (s *TeamService) CalculateWinRate(ctx context.Context, playerID, countStr, typeFilter, queueFilter string) (float64, error) {
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		return 0, fmt.Errorf("invalid count parameter")
	}

	playerIDInt, _ := strconv.ParseInt(playerID, 10, 64)

	player, err := s.GetPlayerById(ctx, playerIDInt)
	if err != nil {
		return 0, err
	}

	matchIDs, err := s.riotClient.GetMatchIDs(player.Puuid, count, typeFilter, queueFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch match IDs: %w", err)
	}
	wins := 0
	total := 0
	for _, matchID := range matchIDs {
		match, err := s.riotClient.GetMatchDetail(matchID)
		if err != nil {
			continue // ignora erros individuais
		}
		for _, p := range match.Info.Participants {
			if p.Puuid == player.Puuid {
				total++

				fmt.Printf("%s -> Champion: %s | Vitória? %t | Tipo: %s\n", p.RiotIdGameName, p.Champion, p.Win, match.Info.QueueName)
				if p.Win {
					wins++
				}
				break
			}
		}
	}
	if total == 0 {
		return 0, fmt.Errorf("no valid matches found")
	}

	winrate := (float64(wins) / float64(total)) * 100
	fmt.Printf("Winrate: %.2f%%\n", winrate)
	return winrate, nil
}
