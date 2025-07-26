package riot

import (
	"APIRankLolV2/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
)

type RiotClient struct {
	apiKey string
	client *http.Client
}

func NewRiotClient(apiKey string) *RiotClient {
	return &RiotClient{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

type AccountDto struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

func (r *RiotClient) GetAccountByRiotID(gameName, tagLine string) (*AccountDto, error) {
	url := fmt.Sprintf("https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s", gameName, tagLine)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Riot-Token", r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
	}

	var account AccountDto
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *RiotClient) GetMatchIDs(puuid string, count int, typeFilter, queueFilter string) ([]string, error) {
	url :=
		fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?count=%d", puuid, count)
	if typeFilter != "" {
		url += "&type=" + typeFilter
	}
	if queueFilter != "" {
		url += "&queue=" + queueFilter
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Riot-Token", r.apiKey)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
	}
	var matchIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
		return nil, err
	}
	return matchIDs, nil
}

type MatchResponse struct {
	Info struct {
		QueueID      int    `json:"queueId"`
		QueueName    string `json:"-"`
		Participants []struct {
			Puuid          string `json:"puuid"`
			Win            bool   `json:"win"`
			Champion       string `json:"championName"`
			RiotIdGameName string `json:"riotIdGameName"`
		} `json:"participants"`
	} `json:"info"`
}

func (r *RiotClient) GetMatchDetail(matchID string) (*MatchResponse, error) {
	url := fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/%s", matchID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Riot-Token", r.apiKey)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("riot API error: %d", resp.StatusCode)
	}
	var match MatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&match); err != nil {
		return nil, err
	}

	q := util.NewQueueIdentifier()
	match.Info.QueueName = q.GetQueueNameByID(match.Info.QueueID)

	return &match, nil
}
