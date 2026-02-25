package models

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/church"
	"os"
	"time"
)

type ChurchModel struct {
	Db *ent.Client
}

func (m *ChurchModel) EmailExists(ctx context.Context, email string) (bool, error) {
	emailExists, err := m.Db.Church.Query().Where(church.EmailEQ(email)).Exist(ctx)
	return emailExists, err
}

func (m *ChurchModel) InviteChurch(ctx context.Context, dto *InviteDto) ModelResponse {
	emailExist, err := m.EmailExists(ctx, dto.Email)
	if err != nil {
		return ModelResponse{
			Data:  nil,
			Error: err,
		}
	}
	if emailExist {
		return ModelResponse{
			Data:  nil,
			Error: EmailAlreadyExist,
		}
	}

	token, err := m.GenerateToken(dto.Email, dto.ExpiresAt)
	if err != nil {
		return ModelResponse{
			Data:  nil,
			Error: err,
		}
	}

	userInvite, err := m.Db.Church.Create().
		SetEmail(dto.Email).
		SetName(dto.Name).
		SetAddress(dto.Address).
		SetType(church.Type(dto.Branch)).
		SetRegistrationToken(token).
		Save(ctx)
	if err != nil {
		return ModelResponse{
			Data:  nil,
			Error: CreationError,
		}
	}
	return ModelResponse{
		Data: struct {
			InviteToken string
			ChurchName  string
		}{
			InviteToken: token,
			ChurchName:  userInvite.Name,
		},
		Error: nil,
	}

}

func (m *ChurchModel) GenerateToken(email string, hours int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(hours) * time.Hour)

	claims := &jwt.RegisteredClaims{
		Issuer:    "church-admin-app",
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ExtractEmailFromToken decodes the JWT subject (email) without strict validation.
// Used only to pre-fill the register form — actual validation happens in VerifyToken.
func (m *ChurchModel) ExtractEmailFromToken(tokenString string) string {
	if tokenString == "" {
		return ""
	}
	token, _ := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if token != nil {
		if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
			return claims.Subject
		}
	}
	return ""
}

// GetByID returns a single church by ID with users edge loaded.
func (m *ChurchModel) GetByID(ctx context.Context, id int) (*ent.Church, error) {
	c, err := m.Db.Church.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return c, nil
}

// UpdateSettings updates editable church settings including congregation size.
func (m *ChurchModel) UpdateSettings(ctx context.Context, churchID int, dto *ChurchSettingsDto) error {
	return m.Db.Church.UpdateOneID(churchID).
		SetName(dto.Name).
		SetAddress(dto.Address).
		SetCity(dto.City).
		SetCountry(dto.Country).
		SetNillablePhone(nullStr(dto.Phone)).
		SetNillableWebsite(nullStr(dto.Website)).
		SetCongregationSize(dto.CongregationSize).
		Exec(ctx)
}

func (m *ChurchModel) VerifyToken(tokenString, expectedEmail string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return false, TokenValidationError
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.Subject != expectedEmail {
			return false, TokenMissMatchError
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return false, TokenExpiredError
		}

		return true, nil
	}

	return false, fmt.Errorf("invalid token claims")
}
