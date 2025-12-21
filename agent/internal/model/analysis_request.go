package model

type AlertAnalysisRequest struct {
	Alert       Alert  `json:"alert"`
	ThreadTS    string `json:"thread_ts"`
	CallbackURL string `json:"callback_url"`
}
