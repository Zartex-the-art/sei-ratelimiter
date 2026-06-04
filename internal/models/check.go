package models

type CheckRequest struct {
	ClientID   string `json:"client_id"`
	Algorithm  string `json:"algorithm"`
	Limit      int    `json:"limit"`
	WindowSecs int    `json:"window_secs"`
}
type CheckResponse struct {
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
	Algorithm string `json:"algorithm"`
	ClientID  string `json:"client_id"`
	Error     string `json:"error,omitempty"`
}
