package models

type SearchRequest struct {
	Query string `json:"query" binding:"required"`
	Mode  string `json:"mode"` // auto, simple, pro
}

type SearchResponse struct {
	Query          string   `json:"query"`
	Mode           string   `json:"mode"`
	Answer         string   `json:"answer"`
	Sources        []Source `json:"sources"`
	Reasoning      string   `json:"reasoning,omitempty"`
	ProcessingTime float64  `json:"processing_time"`
	Timestamp      int64    `json:"timestamp"`
	SessionID      string   `json:"session_id,omitempty"`
	ContextUsed    bool     `json:"context_used,omitempty"`
}

type Source struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Snippet     string  `json:"snippet"`
	Credibility float64 `json:"credibility,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TavilySearchRequest struct {
	Query             string `json:"query"`
	MaxResults        int    `json:"max_results"`
	IncludeRawContent bool   `json:"include_raw_content"`
}

type TavilySearchResponse struct {
	Results []TavilyResult `json:"results"`
	Query   string         `json:"query"`
}

type TavilyResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Snippet    string  `json:"snippet"`
	RawContent string  `json:"raw_content,omitempty"`
	Score      float64 `json:"score"`
}
