package pg

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"github.com/alexevil1979/indrive/services/ride/internal/domain"
)

// RatingRepo handles rating persistence
type RatingRepo struct {
	db *sql.DB
}

// NewRatingRepo creates a new rating repository
func NewRatingRepo(db *sql.DB) *RatingRepo {
	return &RatingRepo{db: db}
}

// Create inserts a new rating
func (r *RatingRepo) Create(ctx context.Context, rating *domain.Rating) error {
	query := `
		INSERT INTO ratings (ride_id, from_user_id, to_user_id, role, score, comment, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		rating.RideID,
		rating.FromUserID,
		rating.ToUserID,
		rating.Role,
		rating.Score,
		nullString(rating.Comment),
		pq.Array(rating.Tags),
	).Scan(&rating.ID, &rating.CreatedAt)
}

// GetByRideID returns all ratings for a ride
func (r *RatingRepo) GetByRideID(ctx context.Context, rideID string) ([]domain.Rating, error) {
	query := `
		SELECT id, ride_id, from_user_id, to_user_id, role, score, 
		       COALESCE(comment, ''), tags, created_at
		FROM ratings
		WHERE ride_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, rideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		var tags pq.StringArray
		if err := rows.Scan(
			&rating.ID, &rating.RideID, &rating.FromUserID, &rating.ToUserID,
			&rating.Role, &rating.Score, &rating.Comment, &tags, &rating.CreatedAt,
		); err != nil {
			return nil, err
		}
		rating.Tags = tags
		ratings = append(ratings, rating)
	}
	return ratings, rows.Err()
}

// GetByUserID returns ratings received by a user with optional role filter
func (r *RatingRepo) GetByUserID(ctx context.Context, userID, role string, limit, offset int) ([]domain.Rating, error) {
	query := `
		SELECT id, ride_id, from_user_id, to_user_id, role, score, 
		       COALESCE(comment, ''), tags, created_at
		FROM ratings
		WHERE to_user_id = $1
	`
	args := []interface{}{userID}

	if role != "" {
		query += " AND role = $2"
		args = append(args, role)
	}

	query += " ORDER BY created_at DESC LIMIT $" + string(rune('0'+len(args)+1)) + " OFFSET $" + string(rune('0'+len(args)+2))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		var tags pq.StringArray
		if err := rows.Scan(
			&rating.ID, &rating.RideID, &rating.FromUserID, &rating.ToUserID,
			&rating.Role, &rating.Score, &rating.Comment, &tags, &rating.CreatedAt,
		); err != nil {
			return nil, err
		}
		rating.Tags = tags
		ratings = append(ratings, rating)
	}
	return ratings, rows.Err()
}

// GetUserRating returns aggregated rating for a user
func (r *RatingRepo) GetUserRating(ctx context.Context, userID, role string) (*domain.UserRating, error) {
	query := `
		SELECT user_id, role, average_score, total_ratings,
		       score_5_count, score_4_count, score_3_count, score_2_count, score_1_count
		FROM user_ratings
		WHERE user_id = $1 AND role = $2
	`
	var ur domain.UserRating
	err := r.db.QueryRowContext(ctx, query, userID, role).Scan(
		&ur.UserID, &ur.Role, &ur.AverageScore, &ur.TotalRatings,
		&ur.Score5Count, &ur.Score4Count, &ur.Score3Count, &ur.Score2Count, &ur.Score1Count,
	)
	if err == sql.ErrNoRows {
		// Return default empty rating
		return &domain.UserRating{
			UserID:       userID,
			Role:         role,
			AverageScore: 0,
			TotalRatings: 0,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &ur, nil
}

// HasRated checks if a user has already rated another user for a ride
func (r *RatingRepo) HasRated(ctx context.Context, rideID, fromUserID, toUserID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ratings WHERE ride_id = $1 AND from_user_id = $2 AND to_user_id = $3)`
	err := r.db.QueryRowContext(ctx, query, rideID, fromUserID, toUserID).Scan(&exists)
	return exists, err
}

// ListAll returns all ratings with pagination (for admin)
func (r *RatingRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.Rating, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM ratings`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, ride_id, from_user_id, to_user_id, role, score, 
		       COALESCE(comment, ''), tags, created_at
		FROM ratings
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		var tags pq.StringArray
		if err := rows.Scan(
			&rating.ID, &rating.RideID, &rating.FromUserID, &rating.ToUserID,
			&rating.Role, &rating.Score, &rating.Comment, &tags, &rating.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		rating.Tags = tags
		ratings = append(ratings, rating)
	}
	return ratings, total, rows.Err()
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
