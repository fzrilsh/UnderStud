package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when a token fails signature/expiry/shape checks.
var ErrInvalidToken = errors.New("invalid token")

// ErrExpired is returned when a valid token has expired (dipakai untuk bedakan
// 401 TOKEN_EXPIRED dari 401 UNAUTHORIZED di middleware).
var ErrExpired = errors.New("token expired")

// Manager issues and verifies HS256 JWTs carrying the user ID as subject.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Make returns a signed token whose lifetime is the configured TTL
// (default 10 years — long-lived auth, no refresh token by design).
func (m *Manager) Make(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Claims adalah payload yang kita pakai di JWT staf.
type Claims struct {
	UserID string
	Role   string
	Kelas  string
}

// MakeClaims menerbitkan JWT dengan klaim role & kelas (untuk RBAC).
func (m *Manager) MakeClaims(userID, role, kelas string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"role":  role,
		"kelas": kelas,
		"iat":   now.Unix(),
		"exp":   now.Add(m.ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// ParseClaims memvalidasi token dan mengembalikan Claims.
func (m *Manager) ParseClaims(tokenString string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrExpired
		}
		return Claims{}, ErrInvalidToken
	}
	if !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, ErrInvalidToken
	}
	sub, _ := mc["sub"].(string)
	if sub == "" {
		return Claims{}, ErrInvalidToken
	}
	role, _ := mc["role"].(string)
	kelas, _ := mc["kelas"].(string)
	return Claims{UserID: sub, Role: role, Kelas: kelas}, nil
}

// Parse validates signature + expiry and returns the subject (user ID).
func (m *Manager) Parse(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
