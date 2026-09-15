package repository

import (
	"context"
	"encoding/json"
	"faceit_stats_wrapper/internal/entity"
	"faceit_stats_wrapper/internal/service"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type faceitAPI struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewFaceitAPI(client *http.Client, apiKey string) service.FaceitRepository {
	return &faceitAPI{
		client:  client,
		apiKey:  apiKey,
		baseURL: "https://open.faceit.com/data/v4",
	}
}

type playerDTO struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Games    struct {
		CS2 struct {
			SkillLevel int `json:"skill_level"`
			FaceitElo  int `json:"faceit_elo"`
		} `json:"cs2"`
	} `json:"games"`
}

type historyDTO struct {
	Items []struct {
		MatchID string `json:"match_id"`
	} `json:"items"`
}

type MatchStatsDTO struct {
	Rounds []struct {
		RoundStats struct {
			Map   string `json:"Map"`
			Score string `json:"Score"`
		} `json:"round_stats"`
		Teams []struct {
			Players []struct {
				PlayerID    string `json:"player_id"`
				PlayerStats struct {
					Kills   string `json:"Kills"`
					Deaths  string `json:"Deaths"`
					Assists string `json:"Assists"`
					KDRatio string `json:"K/D Ratio"`
					Result  string `json:"Result"`
				} `json:"player_stats"`
			} `json:"players"`
		} `json:"teams"`
	} `json:"rounds"`
}

func (api *faceitAPI) GetPlayerByNickname(ctx context.Context, nickname string) (*entity.Player, error) {
	params := url.Values{}
	params.Set("nickname", nickname)

	url := api.baseURL + "/players?" + params.Encode()

	// url := fmt.Sprintf("%s/players?nickname=%s", api.baseURL, nickname)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+api.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var dto playerDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &entity.Player{
		ID:       dto.PlayerID,
		Nickname: dto.Nickname,
		Level:    dto.Games.CS2.SkillLevel,
		Elo:      dto.Games.CS2.FaceitElo,
		Avatar:   dto.Avatar,
	}, nil
}

func (api *faceitAPI) GetPlayerLastMatch(ctx context.Context, playerID string) (*entity.Match, error) {
	matchIDs, err := api.getPlayerMatchIDs(ctx, playerID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get match ID: %w", err)
	}

	if len(matchIDs) == 0 {
		return nil, fmt.Errorf("no matches found for player %s", playerID)
	}

	matchID := matchIDs[0]

	return api.getMatchStats(ctx, matchID, playerID)
}

func (api *faceitAPI) getMatchStats(ctx context.Context, matchID string, playerID string) (*entity.Match, error) {
	matchURL := fmt.Sprintf("%s/matches/%s/stats", api.baseURL, matchID)

	matchReq, err := http.NewRequestWithContext(ctx, http.MethodGet, matchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create match request: %w", err)
	}

	matchReq.Header.Set("Authorization", "Bearer "+api.apiKey)

	matchResp, err := api.client.Do(matchReq)
	if err != nil {
		return nil, fmt.Errorf("match stats http request failed: %w", err)
	}
	defer matchResp.Body.Close()

	if matchResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("faceit match stats: unexpected status code: %d", matchResp.StatusCode)
	}

	var statsDTO MatchStatsDTO
	if err := json.NewDecoder(matchResp.Body).Decode(&statsDTO); err != nil {
		return nil, fmt.Errorf("failed to decode match stats: %w", err)
	}

	if len(statsDTO.Rounds) == 0 {
		return nil, fmt.Errorf("no rounds found in match stats")
	}

	round := statsDTO.Rounds[0]

	for _, team := range round.Teams {
		for _, player := range team.Players {
			if player.PlayerID == playerID {
				kills, _ := strconv.Atoi(player.PlayerStats.Kills)
				deaths, _ := strconv.Atoi(player.PlayerStats.Deaths)
				kd, _ := strconv.ParseFloat(player.PlayerStats.KDRatio, 64)

				resultTxt := "Loss"
				if player.PlayerStats.Result == "1" {
					resultTxt = "Win"
				}

				return &entity.Match{
					MatchID: matchID,
					Map:     round.RoundStats.Map,
					Result:  resultTxt,
					Score:   round.RoundStats.Score,
					Kills:   kills,
					Deaths:  deaths,
					KDRatio: kd,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("player %s not found in match %s", playerID, matchID)
}

func (api *faceitAPI) getPlayerMatchIDs(ctx context.Context, playerID string, limit int) ([]string, error) {
	historyURL := fmt.Sprintf("%s/players/%s/history?game=cs2&offset=0&limit=%d", api.baseURL, playerID, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, historyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create history request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+api.apiKey)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("history http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("faceit history: unexpected status code: %d", resp.StatusCode)
	}
	var history historyDTO

	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to decode history: %w", err)
	}

	matchIDs := make([]string, 0, 20)

	if len(history.Items) == 0 {
		return matchIDs, nil
	}

	for _, v := range history.Items {
		matchIDs = append(matchIDs, v.MatchID)
	}

	return matchIDs, nil

}

func (api *faceitAPI) GetPlayerMatches(ctx context.Context, playerID string, limit int) ([]entity.Match, error) {
	matchIDs, err := api.getPlayerMatchIDs(ctx, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get match IDs for player %s: %w", playerID, err)
	}

	matchStats := make([]entity.Match, 0, len(matchIDs))

	for _, v := range matchIDs {
		match, err := api.getMatchStats(ctx, v, playerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get match stats: %w", err)
		}
		matchStats = append(matchStats, *match)
	}

	return matchStats, nil
}
