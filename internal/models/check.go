package models

// CheckRequest — algorithm, limit, window_secs are optional
// when a stored rule exists for the client.
type CheckRequest struct {
	ClientID   string `json:"client_id"`
	Algorithm  string `json:"algorithm,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	WindowSecs int    `json:"window_secs,omitempty"`
}

// CheckResponse — rule_id shows which stored rule was applied.
type CheckResponse struct {
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
	Algorithm string `json:"algorithm"`
	ClientID  string `json:"client_id"`
	RuleID    string `json:"rule_id,omitempty"`
	Error     string `json:"error,omitempty"`
}
