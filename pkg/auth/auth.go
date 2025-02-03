package auth

import (
	"fmt"
	"github.com/o1egl/paseto"
	"goodblast/internal/domain/entity"
	"sync"
	"time"
)

var (
	instance *Auth
	once     sync.Once
)

type IAuth interface {
	GenerateToken(user *entity.User) (string, error)
	VerifyToken(token string) (map[string]interface{}, error)
}

type Auth struct {
	secretKey []byte
	tokenTTL  time.Duration
}

func InitAuth(secretKey []byte, tokenTTL time.Duration) {
	once.Do(func() {
		instance = &Auth{
			secretKey: secretKey,
			tokenTTL:  tokenTTL,
		}
	})
}

func GetAuth() *Auth {
	if instance == nil {
		panic("auth package is not initialized. Call InitAuth first.")
	}
	return instance
}

type TokenPayload struct {
	UserID   int64  `json:"userID"`
	Username string `json:"username"`
	Exp      string `json:"exp"`
}

func (a *Auth) GenerateToken(user *entity.User) (string, error) {
	now := time.Now()
	exp := now.Add(a.tokenTTL)

	payload := TokenPayload{
		UserID:   user.ID,
		Username: user.Username,
		Exp:      exp.Format(time.RFC3339),
	}

	token, err := paseto.NewV2().Encrypt(a.secretKey, payload, nil)
	if err != nil {
		return "", fmt.Errorf("token create error: %v", err)
	}

	return token, nil
}

func (a *Auth) VerifyToken(token string) (TokenPayload, error) {
	var payload TokenPayload
	err := paseto.NewV2().Decrypt(token, a.secretKey, &payload, nil)
	if err != nil {
		return TokenPayload{}, fmt.Errorf("token decrypt error: %v", err)
	}

	exp, err := time.Parse(time.RFC3339, payload.Exp)
	if err != nil || time.Now().After(exp) {
		return TokenPayload{}, fmt.Errorf("token expired")
	}

	return payload, nil
}
