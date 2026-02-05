package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridehail/ride/internal/domain"
)

// RatingRepo handles rating persistence
type RatingRepo struct {
	pool *pgxpool.Pool
}

// NewRatingRepo creates a new rating repository
func NewRatingRepo(pool *pgxpool.Pool) *RatingRepo {
	return &RatingRepo{pool: pool}
}

// Create inserts a new rating
func (r *RatingRepo) Create(ctx context.Context, rating *domain.Rating) error {
	query := `
		INSERT INTO ratings (ride_id, from_user_id, to_user_id, role, score, comment, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		rating.RideID,
		rating.FromUserID,
		rating.ToUserID,
		rating.Role,
		rating.Score,
		nullStr(rating.Comment),
		rating.Tags,
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
	rows, err := r.pool.Query(ctx, query, rideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		if err := rows.Scan(
			&rating.ID, &rating.RideID, &rating.FromUserID, &rating.ToUserID,
			&rating.Role, &rating.Score, &rating.Comment, &rating.Tags, &rating.CreatedAt,
		); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}
	return ratings, rows.Err()
}

// GetByUserID returns ratings received by a user with optional role filter
func (r *RatingRepo) GetByUserID(ctx context.Context, userID, role string, limit, offset int) ([]domain.Rating, error) {
	var query string
	var args []interface{}

	if role != "" {
		query = `
			SELECT id, ride_id, from_user_id, to_user_id, role, score, 
			       COALESCE(comment, ''), tags, created_at
			FROM ratings
			WHERE to_user_id = $1 AND role = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{userID, role, limit, offset}
	} else {
		query = `
			SELECT id, ride_id, from_user_id, to_user_id, role, score, 
			       COALESCE(comment, ''), tags, created_at
			FROM ratings
			WHERE to_user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{userID, limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		if err := rows.Scan(
			&rating.ID, &rating.RideID, &rating.FromUserID, &rating.ToUserID,
			&rating.Role, &rating.Score, &rating.Comment, &rating.Tags, &rating.CreatedAt,
		); err != nil {
			return nil, err
		}
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
	err := r.pool.QueryRow(ctx, query, userID, role).Scan(
		&ur.UserID, &ur.Role, &ur.AverageScore, &ur.TotalRatings,
		&ur.Score5Count, &ur.Score4Count, &ur.Score3Count, &ur.Score2Count, &ur.Score1Count,
	)
	if err == pgx.ErrNoRows {
		// Return default empty rating
		return &domain.UserRating{
			UserID:       userID,
			Role:         role,
			AverageScore: 0,
			TotalRatings: 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user rating: %w", err)
	}
	return &ur, nil
}

// HasRated checks if a user has already rated another user for a ride
func (r *RatingRepo) HasRated(ctx context.Context, rideID, fromUserID, toUserID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ratings WHERE ride_id = $1 AND from_user_id = $2 AND to_user_id = $3)`
	err := r.pool.QueryRow(ctx, query, rideID, fromUserID, toUserID).Scan(&exists)
	return exists, err
}

// ListAll returns all ratings with pagination (for admin)
func (r *RatingRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.Rating, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM ratings`
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, ride_id, from_user_id, to_user_id, role, score, 
		       COALESCE(comment, ''), tags, created_at
		FROM ratings
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rating domain.Rating
		if err := rows.Scan(
			&rating.ID, &rating.RideID, &rating.FromUserID, &rating.ToUserID,
			&rating.Role, &rating.Score, &rating.Comment, &rating.Tags, &rating.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		ratings = append(ratings, rating)
	}
	return ratings, total, rows.Err()
}
