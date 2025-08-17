package usecase

import (
	"APIRankLolV2/internal/domain"
	"APIRankLolV2/internal/infra/riot"
	"APIRankLolV2/internal/util"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var ErrPlayerNotFound = errors.New("player not found")

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
	teamID := util.StringToInt64(teamIDParam)

	playerBuscado, err := s.GetPlayerById(ctx, teamID)
	if errors.Is(err, ErrPlayerNotFound) {
		account, err := s.riotClient.GetAccountByRiotID(player.GamerName, player.TagLine)
		if err != nil {
			return fmt.Errorf("falha ao buscar PUUID na Riot API: %w", err)
		}
		player.Puuid = account.Puuid
		player.TeamID = teamID

		return s.SavePlayer(ctx, player)
	}

	player.ID = playerBuscado.ID
	return s.repo.UpdatePlayer(ctx, player)
}

func (s *TeamService) SavePlayer(ctx context.Context, player domain.Player) error {
	return s.repo.SavePlayer(ctx, player)
}

func (s *TeamService) GetPlayersByGamerName(ctx context.Context, gamerName string) ([]domain.Player, error) {
	return s.repo.FindPlayersByGamerName(ctx, gamerName)
}

func (s *TeamService) GetAllPlayersByTeamID(ctx context.Context, teamID string) ([]domain.Player, error) {
	return s.repo.FindAllPlayersByTeamID(ctx, util.StringToInt64(teamID))
}

func (s *TeamService) GetPlayerById(ctx context.Context, playerID int64) (domain.Player, error) {
	player, err := s.repo.FindPlayerByID(ctx, playerID)
	if err != nil {
		return player, fmt.Errorf("%w: %v", ErrPlayerNotFound, err)
	}

	return player, nil
}

func (s *TeamService) CalculateWinRateTeam(ctx context.Context, teamID, countStr, typeFilter, queueFilter string) (float64, error) {
	teamIdInt := util.StringToInt64(teamID)
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

	playerIDInt := util.StringToInt64(playerID)

	player, err := s.GetPlayerById(ctx, playerIDInt)
	if err != nil {
		return 0, err
	}

	matchIDs, err := s.riotClient.GetMatchIDs(player.Puuid, count, typeFilter, queueFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch match IDs: %w", err)
	}

	start := time.Now()

	matches := s.buscarPartidasConcorrentemente(matchIDs, 3)

	duration := time.Since(start)
	fmt.Printf("Tempo da execução sequencial: %v\n", duration)

	winrate := calcularWinrate(matches, player.Puuid)
	fmt.Printf("Winrate: %.2f%%\n", winrate)
	return winrate, nil
}

func (s *TeamService) buscarDetalhesPartidaComRetry(matchID string, tentativas int) (*riot.MatchResponse, error) {
	var err error
	var match *riot.MatchResponse

	for i := 0; i < tentativas; i++ {
		match, err = s.riotClient.GetMatchDetail(matchID)
		if err == nil {
			return match, nil
		}

		var httpErr *util.HttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusTooManyRequests {
			wait := time.Duration((i+1)*2) * time.Second // Ex: 2s, 4s, 6s
			fmt.Printf("Rate limit para %s. Tentativa %d. Esperando %v...\n", matchID, i+1, wait)
			time.Sleep(wait)
			continue
		}
		break
	}
	return nil, err
}

func (s *TeamService) buscarPartidasConcorrentemente(matchIDs []string, maxConcorrentes int) []riot.MatchResponse {
	var wg sync.WaitGroup
	var mu sync.Mutex

	semaforo := make(chan struct{}, maxConcorrentes) // controla a quantidade de goroutines ativas
	resultados := make([]riot.MatchResponse, 0)

	for _, matchID := range matchIDs {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()
			semaforo <- struct{}{} // ocupa vaga

			defer func() { <-semaforo }() // libera vaga

			partida, err := s.buscarDetalhesPartidaComRetry(id, 5)
			if err != nil {
				fmt.Printf("Erro ao buscar %s: %v\n", id, err)
				return
			}

			mu.Lock()
			resultados = append(resultados, *partida)
			mu.Unlock()
		}(matchID)
	}

	wg.Wait()
	return resultados
}

func calcularWinrate(matches []riot.MatchResponse, puuid string) float64 {
	total, wins := 0, 0

	for _, match := range matches {
		for _, p := range match.Info.Participants {
			if p.Puuid == puuid {
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
		return 0
	}

	return (float64(wins) / float64(total)) * 100
}
