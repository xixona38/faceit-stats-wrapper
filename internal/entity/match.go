package entity

type Match struct {
	MatchID string  `json:"match_id"`
	Map     string  `json:"map"`
	Result  string  `json:"result"`
	Score   string  `json:"score"`
	Kills   int     `json:"kills"`
	Deaths  int     `json:"deaths"`
	Assists int     `json:"assists"`
	KDRatio float64 `json:"kd_ratio"`
}
