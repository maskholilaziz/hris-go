package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type contextKey string
const AdminIDContextKey = contextKey("admin_user_id")

type SuperadminClaims struct {
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secretKey: []byte(secret),
	}
}

func (s *JWTService) GenerateSuperadminToken(adminID uuid.UUID) (string, error) {
	claims := SuperadminClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   adminID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "hris",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateSuperadminToken(tokenString string) (*SuperadminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SuperadminClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method tidak terduga: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SuperadminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token tidak valid")
}

func (s *JWTService) SuperadminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			util.ErrorResponse(w, http.StatusUnauthorized, "Token tidak ditemukan", "Authorization header dibutuhkan")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			util.ErrorResponse(w, http.StatusUnauthorized, "Format token salah", "Format harus 'Bearer {token}'")
		}

		tokenString := parts[1]
		claims, err := s.ValidateSuperadminToken(tokenString)
		if err != nil {
			util.ErrorResponse(w, http.StatusUnauthorized, "Token tidak valid", err.Error())
			return
		}

		adminID, err := uuid.Parse(claims.Subject)
		if err != nil {
			util.ErrorResponse(w, http.StatusUnauthorized, "Token tidak valid", "Invalid subject ID")
			return
		}

		ctx := context.WithValue(r.Context(), AdminIDContextKey, adminID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}