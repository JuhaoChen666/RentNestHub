package domain

import (
	"context"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Role         string    `json:"role"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"displayName"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AccountRepository interface {
	CreateUser(context.Context, *User) error
	FindUserByIdentifier(context.Context, string) (User, error)
	UpdateUserEmail(context.Context, int64, string) error
	UpdateUserPassword(context.Context, string, string) error
}

type PasswordResetStore interface {
	Save(context.Context, string, string, time.Duration) error
	VerifyAndConsume(context.Context, string, string) (bool, error)
}

type PasswordResetMailer interface {
	SendPasswordReset(context.Context, string, string, string) error
}
