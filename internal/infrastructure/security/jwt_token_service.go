package security

import (
	"context"
	"fmt"
	"time"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-image-service/internal/application/abstraction"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// jwtTokenService implements TokenService using JWT
type jwtTokenService struct {
	secretKey []byte
}

// NewJWTTokenService creates a new JWT token service
func NewJWTTokenService(secretKey string) abstraction.TokenService {
	return &jwtTokenService{
		secretKey: []byte(secretKey),
	}
}

// customClaims wraps UploadTokenClaims with standard JWT claims
type customClaims struct {
	UploadTokenClaims abstraction.UploadTokenClaims `json:"upload"`
	jwt.RegisteredClaims
}

func (s *jwtTokenService) GenerateUploadToken(ctx context.Context, claims *abstraction.UploadTokenClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	jwtClaims := customClaims{
		UploadTokenClaims: *claims,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "image-service",
			Subject:   "upload",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.Get(ctx).Debug("generated upload token",
		zap.String("key", claims.Key),
		zap.String("ownerType", claims.OwnerType),
		zap.String("ownerId", claims.OwnerID),
		zap.Time("expiresAt", expiresAt),
	)

	return signedToken, nil
}

func (s *jwtTokenService) ValidateUploadToken(ctx context.Context, tokenString string) (*abstraction.UploadTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate expiration (jwt library does this automatically, but we double-check)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	logger.Get(ctx).Debug("validated upload token",
		zap.String("key", claims.UploadTokenClaims.Key),
		zap.String("ownerType", claims.UploadTokenClaims.OwnerType),
		zap.String("ownerId", claims.UploadTokenClaims.OwnerID),
	)

	return &claims.UploadTokenClaims, nil
}
