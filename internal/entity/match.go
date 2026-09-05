package entity

type Match struct {
	MatchID string  `json:"match_id"`
	Map     string  `json:"map"`
	Result  string  `json:"result"`
	Score   string  `json:"score"`
	Kills   int     `json:"kills"`
	Deaths  int     `json:"deaths"`
	KDRatio float64 `json:"kd_ratio"`
}
