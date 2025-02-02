package models

type Response struct {
	Message string `json:"message"`
	Service Service
	Data    interface{} `json:"data,omitempty"`  // Holds any data to be returned, omitting it if nil
	Error   string      `json:"error,omitempty"` // Holds an error message if applicable
}

type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Check     []Check
}

type Check struct {
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Response   string `json:"response"`
}

type EasyResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}
