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
