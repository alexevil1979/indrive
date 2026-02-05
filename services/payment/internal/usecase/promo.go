package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/alexevil1979/indrive/services/payment/internal/domain"
)

// PromoRepo defines promo repository interface
type PromoRepo interface {
	Create(ctx context.Context, promo *domain.Promo) error
	GetByID(ctx context.Context, id string) (*domain.Promo, error)
	GetByCode(ctx context.Context, code string) (*domain.Promo, error)
	Update(ctx context.Context, promo *domain.Promo) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int, activeOnly bool) ([]domain.Promo, int, error)
	GetUserPromoCount(ctx context.Context, userID, promoID string) (int, error)
	RecordUsage(ctx context.Context, usage *domain.UserPromo) error
	GetUserPromos(ctx context.Context, userID string, limit int) ([]domain.UserPromo, error)
	CheckPromoUsedForRide(ctx context.Context, rideID string) (bool, error)
}

// PromoUseCase handles promo business logic
type PromoUseCase struct {
	repo PromoRepo
}

// NewPromoUseCase creates a new promo use case
func NewPromoUseCase(repo PromoRepo) *PromoUseCase {
	return &PromoUseCase{repo: repo}
}

// ValidatePromo checks if a promo code is valid for a user and order
func (uc *PromoUseCase) ValidatePromo(ctx context.Context, code, userID string, orderAmount float64) (*domain.PromoResult, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return &domain.PromoResult{Valid: false, Error: "Промокод не указан"}, nil
	}

	promo, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.PromoResult{Valid: false, Error: "Промокод не найден"}, nil
		}
		return nil, err
	}

	// Check if promo is valid
	if !promo.IsValid() {
		return &domain.PromoResult{Valid: false, Error: "Промокод недействителен или истёк"}, nil
	}

	// Check minimum order value
	if orderAmount < promo.MinOrderValue {
		return &domain.PromoResult{
			Valid: false,
			Error: "Минимальная сумма заказа для этого промокода: " + formatPrice(promo.MinOrderValue),
		}, nil
	}

	// Check per-user limit
	if promo.PerUserLimit > 0 {
		count, err := uc.repo.GetUserPromoCount(ctx, userID, promo.ID)
		if err != nil {
			return nil, err
		}
		if count >= promo.PerUserLimit {
			return &domain.PromoResult{Valid: false, Error: "Вы уже использовали этот промокод"}, nil
		}
	}

	// Calculate discount
	discount := promo.CalculateDiscount(orderAmount)
	finalPrice := orderAmount - discount

	return &domain.PromoResult{
		Valid:      true,
		Promo:      promo,
		Discount:   discount,
		FinalPrice: finalPrice,
	}, nil
}

// ApplyPromo validates and applies a promo code to a ride
func (uc *PromoUseCase) ApplyPromo(ctx context.Context, code, userID, rideID string, orderAmount float64) (*domain.PromoResult, error) {
	// First validate
	result, err := uc.ValidatePromo(ctx, code, userID, orderAmount)
	if err != nil {
		return nil, err
	}

	if !result.Valid {
		return result, nil
	}

	// Check if promo already used for this ride
	if rideID != "" {
		used, err := uc.repo.CheckPromoUsedForRide(ctx, rideID)
		if err != nil {
			return nil, err
		}
		if used {
			return &domain.PromoResult{Valid: false, Error: "Промокод уже применён к этой поездке"}, nil
		}
	}

	// Record usage
	usage := &domain.UserPromo{
		UserID:   userID,
		PromoID:  result.Promo.ID,
		RideID:   rideID,
		Discount: result.Discount,
	}

	if err := uc.repo.RecordUsage(ctx, usage); err != nil {
		return nil, err
	}

	return result, nil
}

// CreatePromo creates a new promo code (admin)
func (uc *PromoUseCase) CreatePromo(ctx context.Context, promo *domain.Promo) error {
	promo.Code = strings.TrimSpace(strings.ToUpper(promo.Code))
	if promo.Code == "" {
		return errors.New("code is required")
	}
	if promo.Value <= 0 {
		return errors.New("value must be positive")
	}
	if promo.Type != domain.PromoTypePercent && promo.Type != domain.PromoTypeFixed {
		return errors.New("type must be 'percent' or 'fixed'")
	}
	if promo.Type == domain.PromoTypePercent && promo.Value > 100 {
		return errors.New("percent discount cannot exceed 100")
	}

	return uc.repo.Create(ctx, promo)
}

// UpdatePromo updates a promo code (admin)
func (uc *PromoUseCase) UpdatePromo(ctx context.Context, promo *domain.Promo) error {
	return uc.repo.Update(ctx, promo)
}

// DeletePromo deletes a promo code (admin)
func (uc *PromoUseCase) DeletePromo(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// GetPromo returns a promo by ID
func (uc *PromoUseCase) GetPromo(ctx context.Context, id string) (*domain.Promo, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListPromos returns all promos with pagination
func (uc *PromoUseCase) ListPromos(ctx context.Context, limit, offset int, activeOnly bool) ([]domain.Promo, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return uc.repo.List(ctx, limit, offset, activeOnly)
}

// GetUserPromos returns user's promo usage history
func (uc *PromoUseCase) GetUserPromos(ctx context.Context, userID string, limit int) ([]domain.UserPromo, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.GetUserPromos(ctx, userID, limit)
}

// helper
func formatPrice(amount float64) string {
	if amount == float64(int(amount)) {
		return string(rune(int(amount))) + " ₽"
	}
	return string(rune(int(amount*100)/100)) + " ₽"
}
