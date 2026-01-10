package model

type EmbeddingRequest struct {
	IncidentID      string `json:"incident_id"`
	IncidentSummary string `json:"incident_summary"`
}

type EmbeddingResponse struct {
	Status      string `json:"status"`
	EmbeddingID int64  `json:"embedding_id"`
	Model       string `json:"model"`
}

type VectorSearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type VectorSearchResult struct {
	IncidentID      string  `json:"incident_id"`
	IncidentSummary string  `json:"incident_summary"`
	Similarity      float64 `json:"similarity"`
}

type VectorSearchResponse struct {
	Results []VectorSearchResult `json:"results"`
	Model   string               `json:"model"`
}
