package logs

type ListInput struct {
	Source     string `query:"source" enum:"system,caddy,dns,access" default:"system"`
	Cursor     int64  `query:"cursor" minimum:"0" default:"0"`
	Limit      int    `query:"limit" minimum:"1" maximum:"500" default:"200"`
	ProviderID string `query:"provider_id" doc:"按 DNS Provider ID 筛选"`
}

type Entry struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp,omitempty"`
	Level     string         `json:"level,omitempty"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type ListResponse struct {
	Entries    []Entry `json:"entries"`
	NextCursor int64   `json:"next_cursor"`
}

type ListOutput struct{ Body ListResponse }
