package crcon

const (
	ActionMatchEnded = action("MATCH ENDED")
	ActionMatchStart = action("MATCH START")
)

type action string

type getRecentLogsRequest struct {
	End             int      `json:"end"`
	FilterActions   []action `json:"filter_action"`
	FilterPlayer    []string `json:"filter_player"`
	InclusiveFilter bool     `json:"inclusive_filter"`
}

type setMapRequest struct {
	MapId string `json:"map_name"`
}

type messagePlayerRequest struct {
	Message  string `json:"message"`
	PlayerId string `json:"player_id"`
}
