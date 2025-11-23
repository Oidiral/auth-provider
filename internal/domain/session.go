package domain

type Session struct {
	UserID       string `json:"userId"`
	RefreshToken string `json:"refreshToken"`
}
