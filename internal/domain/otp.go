package domain

import "time"

type OTP struct {
	UserId    string    `json:"userId"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"` 
}
