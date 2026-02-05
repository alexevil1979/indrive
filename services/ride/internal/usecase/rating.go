package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ridehail/ride/internal/domain"
)

// RatingRepo defines rating repository interface
type RatingRepo interface {
	Create(ctx context.Context, rating *domain.Rating) error
	GetByRideID(ctx context.Context, rideID string) ([]domain.Rating, error)
	GetByUserID(ctx context.Context, userID, role string, limit, offset int) ([]domain.Rating, error)
	GetUserRating(ctx context.Context, userID, role string) (*domain.UserRating, error)
	HasRated(ctx context.Context, rideID, fromUserID, toUserID string) (bool, error)
	ListAll(ctx context.Context, limit, offset int) ([]domain.Rating, int, error)
}

// RatingUseCase handles rating business logic
type RatingUseCase struct {
	ratingRepo RatingRepo
	rideRepo   RideRepository // To validate ride exists and is completed
}

// NewRatingUseCase creates a new rating use case
func NewRatingUseCase(ratingRepo RatingRepo, rideRepo RideRepository) *RatingUseCase {
	return &RatingUseCase{
		ratingRepo: ratingRepo,
		rideRepo:   rideRepo,
	}
}

// SubmitRatingInput is the input for submitting a rating
type SubmitRatingInput struct {
	RideID     string
	FromUserID string
	Score      int
	Comment    string
	Tags       []string
}

// SubmitRating creates a new rating for a completed ride
func (uc *RatingUseCase) SubmitRating(ctx context.Context, input SubmitRatingInput) (*domain.Rating, error) {
	// Validate score
	if input.Score < 1 || input.Score > 5 {
		return nil, errors.New("score must be between 1 and 5")
	}

	// Get ride to validate it exists and determine participants
	ride, err := uc.rideRepo.GetByID(ctx, input.RideID)
	if err != nil {
		return nil, fmt.Errorf("ride not found: %w", err)
	}

	// Only completed rides can be rated
	if ride.Status != domain.StatusCompleted {
		return nil, errors.New("can only rate completed rides")
	}

	// Determine who is being rated
	var toUserID, role string
	if input.FromUserID == ride.PassengerID {
		// Passenger rating driver
		if ride.DriverID == "" {
			return nil, errors.New("ride has no driver")
		}
		toUserID = ride.DriverID
		role = "driver"
	} else if input.FromUserID == ride.DriverID {
		// Driver rating passenger
		toUserID = ride.PassengerID
		role = "passenger"
	} else {
		return nil, errors.New("user is not a participant of this ride")
	}

	// Check if already rated
	hasRated, err := uc.ratingRepo.HasRated(ctx, input.RideID, input.FromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing rating: %w", err)
	}
	if hasRated {
		return nil, errors.New("you have already rated this ride")
	}

	// Validate tags
	validTags := domain.DriverRatingTags
	if role == "passenger" {
		validTags = domain.PassengerRatingTags
	}
	for _, tag := range input.Tags {
		valid := false
		for _, vt := range validTags {
			if tag == vt {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid tag: %s", tag)
		}
	}

	// Create rating
	rating := &domain.Rating{
		RideID:     input.RideID,
		FromUserID: input.FromUserID,
		ToUserID:   toUserID,
		Role:       role,
		Score:      input.Score,
		Comment:    input.Comment,
		Tags:       input.Tags,
	}

	if err := uc.ratingRepo.Create(ctx, rating); err != nil {
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	return rating, nil
}

// GetRideRatings returns all ratings for a specific ride
func (uc *RatingUseCase) GetRideRatings(ctx context.Context, rideID string) ([]domain.Rating, error) {
	return uc.ratingRepo.GetByRideID(ctx, rideID)
}

// GetUserRatings returns ratings received by a user
func (uc *RatingUseCase) GetUserRatings(ctx context.Context, userID, role string, limit, offset int) ([]domain.Rating, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return uc.ratingRepo.GetByUserID(ctx, userID, role, limit, offset)
}

// GetUserRating returns aggregated rating for a user
func (uc *RatingUseCase) GetUserRating(ctx context.Context, userID, role string) (*domain.UserRating, error) {
	return uc.ratingRepo.GetUserRating(ctx, userID, role)
}

// GetRatingTags returns available tags for a role
func (uc *RatingUseCase) GetRatingTags(role string) []string {
	if role == "passenger" {
		return domain.PassengerRatingTags
	}
	return domain.DriverRatingTags
}

// ListAllRatings returns all ratings with pagination (admin)
func (uc *RatingUseCase) ListAllRatings(ctx context.Context, limit, offset int) ([]domain.Rating, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return uc.ratingRepo.ListAll(ctx, limit, offset)
}
