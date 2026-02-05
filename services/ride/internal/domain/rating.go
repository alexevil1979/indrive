// Package domain — Rating and Review entities
package domain

import "time"

// Rating represents a rating given after a ride
type Rating struct {
	ID         string    `json:"id"`
	RideID     string    `json:"ride_id"`
	FromUserID string    `json:"from_user_id"` // Who gave the rating
	ToUserID   string    `json:"to_user_id"`   // Who received the rating
	Role       string    `json:"role"`         // "passenger" or "driver" (role of ToUserID)
	Score      int       `json:"score"`        // 1-5 stars
	Comment    string    `json:"comment,omitempty"`
	Tags       []string  `json:"tags,omitempty"` // e.g. ["polite", "clean_car", "fast"]
	CreatedAt  time.Time `json:"created_at"`
}

// UserRating represents aggregated rating for a user
type UserRating struct {
	UserID       string  `json:"user_id"`
	Role         string  `json:"role"` // "passenger" or "driver"
	AverageScore float64 `json:"average_score"`
	TotalRatings int     `json:"total_ratings"`
	Score5Count  int     `json:"score_5_count"`
	Score4Count  int     `json:"score_4_count"`
	Score3Count  int     `json:"score_3_count"`
	Score2Count  int     `json:"score_2_count"`
	Score1Count  int     `json:"score_1_count"`
}

// RatingTags — predefined tags for ratings
var DriverRatingTags = []string{
	"polite",         // Вежливый
	"clean_car",      // Чистая машина
	"safe_driving",   // Безопасное вождение
	"fast",           // Быстрая поездка
	"good_music",     // Хорошая музыка
	"comfortable",    // Комфортно
	"on_time",        // Вовремя
	"professional",   // Профессиональный
}

var PassengerRatingTags = []string{
	"polite",         // Вежливый
	"on_time",        // Вовремя
	"friendly",       // Дружелюбный
	"respectful",     // Уважительный
	"clean",          // Аккуратный
}

// TagLabels — human-readable labels for tags
var TagLabels = map[string]string{
	"polite":       "Вежливый",
	"clean_car":    "Чистая машина",
	"safe_driving": "Безопасное вождение",
	"fast":         "Быстрая поездка",
	"good_music":   "Хорошая музыка",
	"comfortable":  "Комфортно",
	"on_time":      "Вовремя",
	"professional": "Профессиональный",
	"friendly":     "Дружелюбный",
	"respectful":   "Уважительный",
	"clean":        "Аккуратный",
}
