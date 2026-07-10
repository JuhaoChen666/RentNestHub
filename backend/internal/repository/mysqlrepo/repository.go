package mysqlrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

var ErrNotFound = errors.New("record not found")

type Repository struct {
	db *sql.DB
}

func New(dsn string) (*Repository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (repository *Repository) Close() error {
	return repository.db.Close()
}

func (repository *Repository) ListHouses(
	ctx context.Context,
	filter domain.HouseFilter,
) ([]domain.House, error) {
	query := `
		SELECT id, landlord_id, title, description, city, district, address,
		       monthly_rent, bedrooms, bathrooms, area_sqm, amenities,
		       image_urls, status, created_at
		FROM houses
		WHERE 1 = 1`
	args := make([]any, 0, 8)

	if filter.OnlyActive {
		query += " AND status = 'active'"
	}
	if filter.City != "" {
		query += " AND city = ?"
		args = append(args, filter.City)
	}
	if filter.District != "" {
		query += " AND district = ?"
		args = append(args, filter.District)
	}
	if filter.MinRent > 0 {
		query += " AND monthly_rent >= ?"
		args = append(args, filter.MinRent)
	}
	if filter.MaxRent > 0 {
		query += " AND monthly_rent <= ?"
		args = append(args, filter.MaxRent)
	}
	if filter.Bedrooms > 0 {
		query += " AND bedrooms >= ?"
		args = append(args, filter.Bedrooms)
	}
	if filter.Keyword != "" {
		keyword := "%" + escapeLike(filter.Keyword) + "%"
		query += ` AND (
			title LIKE ? ESCAPE '\\'
			OR description LIKE ? ESCAPE '\\'
			OR address LIKE ? ESCAPE '\\'
			OR amenities LIKE ? ESCAPE '\\'
		)`
		args = append(args, keyword, keyword, keyword, keyword)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 24
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, filter.Offset)

	rows, err := repository.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	houses := make([]domain.House, 0)
	for rows.Next() {
		house, err := scanHouse(rows)
		if err != nil {
			return nil, err
		}
		houses = append(houses, house)
	}
	return houses, rows.Err()
}

func (repository *Repository) GetHouse(ctx context.Context, id int64) (domain.House, error) {
	row := repository.db.QueryRowContext(ctx, `
		SELECT id, landlord_id, title, description, city, district, address,
		       monthly_rent, bedrooms, bathrooms, area_sqm, amenities,
		       image_urls, status, created_at
		FROM houses
		WHERE id = ?`, id)

	house, err := scanHouse(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.House{}, ErrNotFound
	}
	return house, err
}

func (repository *Repository) CreateHouse(ctx context.Context, house *domain.House) error {
	amenities, err := json.Marshal(house.Amenities)
	if err != nil {
		return err
	}
	images, err := json.Marshal(house.ImageURLs)
	if err != nil {
		return err
	}
	if house.Status == "" {
		house.Status = "active"
	}

	result, err := repository.db.ExecContext(ctx, `
		INSERT INTO houses (
			landlord_id, title, description, city, district, address,
			monthly_rent, bedrooms, bathrooms, area_sqm, amenities,
			image_urls, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		house.LandlordID,
		house.Title,
		house.Description,
		house.City,
		house.District,
		house.Address,
		house.MonthlyRent,
		house.Bedrooms,
		house.Bathrooms,
		house.AreaSqm,
		amenities,
		images,
		house.Status,
	)
	if err != nil {
		return err
	}
	house.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	house.CreatedAt = time.Now()
	return nil
}

func (repository *Repository) AddFavorite(ctx context.Context, favorite domain.Favorite) error {
	_, err := repository.db.ExecContext(ctx, `
		INSERT INTO favorites (tenant_id, house_id)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE created_at = created_at`,
		favorite.TenantID,
		favorite.HouseID,
	)
	return err
}

func (repository *Repository) RemoveFavorite(ctx context.Context, favorite domain.Favorite) error {
	_, err := repository.db.ExecContext(ctx,
		"DELETE FROM favorites WHERE tenant_id = ? AND house_id = ?",
		favorite.TenantID,
		favorite.HouseID,
	)
	return err
}

func (repository *Repository) CreateMessage(ctx context.Context, message *domain.Message) error {
	result, err := repository.db.ExecContext(ctx, `
		INSERT INTO messages (house_id, sender_id, content)
		VALUES (?, ?, ?)`,
		message.HouseID,
		message.SenderID,
		message.Content,
	)
	if err != nil {
		return err
	}
	message.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	message.CreatedAt = time.Now()
	return nil
}

type scanner interface {
	Scan(...any) error
}

func scanHouse(row scanner) (domain.House, error) {
	var house domain.House
	var amenitiesJSON []byte
	var imagesJSON []byte
	err := row.Scan(
		&house.ID,
		&house.LandlordID,
		&house.Title,
		&house.Description,
		&house.City,
		&house.District,
		&house.Address,
		&house.MonthlyRent,
		&house.Bedrooms,
		&house.Bathrooms,
		&house.AreaSqm,
		&amenitiesJSON,
		&imagesJSON,
		&house.Status,
		&house.CreatedAt,
	)
	if err != nil {
		return domain.House{}, err
	}
	if err := json.Unmarshal(amenitiesJSON, &house.Amenities); err != nil {
		return domain.House{}, fmt.Errorf("decode amenities: %w", err)
	}
	if err := json.Unmarshal(imagesJSON, &house.ImageURLs); err != nil {
		return domain.House{}, fmt.Errorf("decode image urls: %w", err)
	}
	return house, nil
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
