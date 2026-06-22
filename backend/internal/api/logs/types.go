package logs

type ListInput struct {
	Source string `query:"source" enum:"system,caddy" default:"system"`
	Cursor int64  `query:"cursor" minimum:"0" default:"0"`
	Limit  int    `query:"limit" minimum:"1" maximum:"500" default:"200"`
}

type Entry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp,omitempty"`
	Level     string `json:"level,omitempty"`
	Message   string `json:"message"`
}

type ListResponse struct {
	Entries    []Entry `json:"entries"`
	NextCursor int64   `json:"next_cursor"`
}

type ListOutput struct{ Body ListResponse }
