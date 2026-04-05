package model

import "testing"

func TestEmbeddingRequestValidationFields(t *testing.T) {
	req := EmbeddingRequest{}
	if req.IncidentID != "" || req.IncidentSummary != "" {
		t.Fatalf("expected zero values")
	}
}
