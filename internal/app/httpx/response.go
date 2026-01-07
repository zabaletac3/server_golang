package httpx

type StandardResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	StatusCode int         `json:"statusCode"`
	Timestamp  string      `json:"timestamp"`
	Path        string      `json:"path"`
	RequestID  string      `json:"request_id,omitempty"`
}