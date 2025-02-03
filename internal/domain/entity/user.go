package entity

import (
	"golang.org/x/crypto/bcrypt"
	"goodblast/internal/application/controller/request"
)

type User struct {
	ID           int64  `bun:"id,pk,autoincrement"`
	Username     string `bun:"username,unique,notnull"`
	PasswordHash string `bun:"password_hash,notnull"`
	Coins        int64  `bun:"coins,default:1000"`
	Level        int    `bun:"level,default:1"`
	Country      string `bun:"country"`
}

func NewUserFromRequest(request request.CreateUserRequest) User {
	hashedPassword, err := hashPassword(request.Password)
	if err != nil {
		panic(err)
	}
	return User{
		Username:     request.Username,
		PasswordHash: hashedPassword,
		Coins:        1000,
		Level:        1,
		Country:      request.Country,
	}
}

func (u *User) CheckPassword(password string) error {
	if err := checkPassword(u.PasswordHash, password); err != nil {
		return err
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func checkPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (u *User) IncrementLevel(coinPerLevel int) {
	u.Level++
	u.Coins += int64(coinPerLevel)
}
