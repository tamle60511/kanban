package models

import "time"

// AccessLog represents a log of user access to operations
type AccessLog struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	OperationID  int       `json:"operation_id"`
	AccessTime   time.Time `json:"access_time"`
	SearchParams string    `json:"search_params,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Status       string    `json:"status"`
}
