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
	query += " " + orderByClause(filter.Sort) + " LIMIT ? OFFSET ?"
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

func orderByClause(sort string) string {
	switch sort {
	case "rent_asc":
		return "ORDER BY monthly_rent ASC, created_at DESC"
	case "rent_desc":
		return "ORDER BY monthly_rent DESC, created_at DESC"
	default:
		return "ORDER BY created_at DESC"
	}
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

func (repository *Repository) ListOwnedHouses(ctx context.Context, ownerID int64) ([]domain.House, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT id, landlord_id, title, description, city, district, address,
		       monthly_rent, bedrooms, bathrooms, area_sqm, amenities,
		       image_urls, status, created_at
		FROM houses
		WHERE landlord_id = ?
		ORDER BY created_at DESC, id DESC`, ownerID)
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

func (repository *Repository) UpdateHouse(ctx context.Context, house *domain.House) error {
	amenities, err := json.Marshal(house.Amenities)
	if err != nil {
		return err
	}

	result, err := repository.db.ExecContext(ctx, `
		UPDATE houses
		SET title = ?, description = ?, city = ?, district = ?, address = ?,
		    monthly_rent = ?, bedrooms = ?, bathrooms = ?, area_sqm = ?,
		    amenities = ?, status = 'draft'
		WHERE id = ?`,
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
		house.ID,
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

func (repository *Repository) UpdateHouseStatus(ctx context.Context, houseID int64, status string) error {
	result, err := repository.db.ExecContext(ctx,
		"UPDATE houses SET status = ? WHERE id = ?", status, houseID,
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

func (repository *Repository) DeleteHouse(ctx context.Context, houseID int64) error {
	result, err := repository.db.ExecContext(ctx, "DELETE FROM houses WHERE id = ?", houseID)
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

func (repository *Repository) ListPendingHouseReviews(ctx context.Context) ([]domain.HouseReview, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT h.id, h.landlord_id, h.title, h.description, h.city, h.district, h.address,
		       h.monthly_rent, h.bedrooms, h.bathrooms, h.area_sqm, h.amenities,
		       h.image_urls, h.status, h.created_at,
		       u.id, u.role, u.username, u.display_name, u.email, u.created_at
		FROM houses h
		JOIN users u ON u.id = h.landlord_id
		WHERE h.status = 'draft'
		ORDER BY h.created_at ASC, h.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]domain.HouseReview, 0)
	for rows.Next() {
		var review domain.HouseReview
		var amenitiesJSON, imagesJSON []byte
		if err := rows.Scan(
			&review.House.ID, &review.House.LandlordID, &review.House.Title, &review.House.Description,
			&review.House.City, &review.House.District, &review.House.Address, &review.House.MonthlyRent,
			&review.House.Bedrooms, &review.House.Bathrooms, &review.House.AreaSqm, &amenitiesJSON,
			&imagesJSON, &review.House.Status, &review.House.CreatedAt, &review.Publisher.ID,
			&review.Publisher.Role, &review.Publisher.Username, &review.Publisher.DisplayName,
			&review.Publisher.Email, &review.Publisher.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(amenitiesJSON, &review.House.Amenities); err != nil {
			return nil, fmt.Errorf("decode amenities: %w", err)
		}
		if err := json.Unmarshal(imagesJSON, &review.House.ImageURLs); err != nil {
			return nil, fmt.Errorf("decode image urls: %w", err)
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (repository *Repository) ReviewHouse(ctx context.Context, houseID int64, approved bool) error {
	status := "archived"
	if approved {
		status = "active"
	}
	result, err := repository.db.ExecContext(ctx,
		"UPDATE houses SET status = ? WHERE id = ? AND status = 'draft'", status, houseID,
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

func (repository *Repository) ListFavoriteHouses(ctx context.Context, tenantID int64) ([]domain.House, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT h.id, h.landlord_id, h.title, h.description, h.city, h.district, h.address,
		       h.monthly_rent, h.bedrooms, h.bathrooms, h.area_sqm, h.amenities,
		       h.image_urls, h.status, h.created_at
		FROM favorites f
		JOIN houses h ON h.id = f.house_id
		WHERE f.tenant_id = ? AND h.status = 'active'
		ORDER BY f.created_at DESC`, tenantID)
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

func (repository *Repository) CreateMessage(ctx context.Context, message *domain.Message) error {
	result, err := repository.db.ExecContext(ctx, `
		INSERT INTO messages (house_id, sender_id, recipient_id, content)
		VALUES (?, ?, ?, ?)`,
		message.HouseID,
		message.Sender.ID,
		message.Recipient.ID,
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

func (repository *Repository) ListMessages(ctx context.Context, userID int64) ([]domain.Message, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT m.id, m.house_id, h.title, m.content, m.created_at,
		       sender.id, sender.role, sender.username, sender.display_name, sender.email,
		       recipient.id, recipient.role, recipient.username, recipient.display_name, recipient.email
		FROM messages m
		JOIN houses h ON h.id = m.house_id
		JOIN users sender ON sender.id = m.sender_id
		JOIN users recipient ON recipient.id = m.recipient_id
		WHERE m.sender_id = ? OR m.recipient_id = ?
		ORDER BY m.created_at ASC, m.id ASC`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(
			&message.ID,
			&message.HouseID,
			&message.HouseTitle,
			&message.Content,
			&message.CreatedAt,
			&message.Sender.ID,
			&message.Sender.Role,
			&message.Sender.Username,
			&message.Sender.DisplayName,
			&message.Sender.Email,
			&message.Recipient.ID,
			&message.Recipient.Role,
			&message.Recipient.Username,
			&message.Recipient.DisplayName,
			&message.Recipient.Email,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (repository *Repository) ConversationExists(ctx context.Context, houseID, firstUserID, secondUserID int64) (bool, error) {
	var exists bool
	err := repository.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM messages
			WHERE house_id = ?
			  AND ((sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?))
		)`, houseID, firstUserID, secondUserID, secondUserID, firstUserID,
	).Scan(&exists)
	return exists, err
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
