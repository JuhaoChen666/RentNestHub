package mysqlrepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

func (repository *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	result, err := repository.db.ExecContext(ctx, `
		INSERT INTO users (role, username, display_name, email, password_hash)
		VALUES (?, ?, ?, ?, ?)`,
		user.Role,
		user.Username,
		user.DisplayName,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		return err
	}
	user.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return repository.db.QueryRowContext(ctx,
		"SELECT created_at FROM users WHERE id = ?", user.ID,
	).Scan(&user.CreatedAt)
}

func (repository *Repository) FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error) {
	var user domain.User
	err := repository.db.QueryRowContext(ctx, `
		SELECT id, role, username, display_name, email, password_hash, created_at
		FROM users
		WHERE username = ? OR email = ?
		LIMIT 1`, identifier, identifier,
	).Scan(
		&user.ID,
		&user.Role,
		&user.Username,
		&user.DisplayName,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrNotFound
	}
	return user, err
}

func (repository *Repository) UpdateUserPassword(ctx context.Context, email, passwordHash string) error {
	result, err := repository.db.ExecContext(ctx,
		"UPDATE users SET password_hash = ? WHERE email = ?", passwordHash, email,
	)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return ErrNotFound
	}
	return nil
}

func (repository *Repository) UpdateUserEmail(ctx context.Context, userID int64, email string) error {
	result, err := repository.db.ExecContext(ctx,
		"UPDATE users SET email = ? WHERE id = ?", email, userID,
	)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return ErrNotFound
	}
	return nil
}
