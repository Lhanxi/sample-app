package item

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("item not found")
	ErrConflict = errors.New("item already exists")
)

type Repository interface {
	List(ctx context.Context) ([]Item, error)
	GetByID(ctx context.Context, id string) (Item, error)
	Create(ctx context.Context, input CreateItemRequest) (Item, error)
	Update(ctx context.Context, id string, input UpdateItemRequest) (Item, error)
	Delete(ctx context.Context, id string) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context) ([]Item, error) {
	const query = `
		SELECT id, name, description, created_at, updated_at
		FROM items
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (Item, error) {
	const query = `
		SELECT id, name, description, created_at, updated_at
		FROM items
		WHERE id = $1`

	item, err := scanItem(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("get item: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	input CreateItemRequest,
) (Item, error) {
	const query = `
		INSERT INTO items (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_at, updated_at`

	item, err := scanItem(r.db.QueryRow(
		ctx,
		query,
		uuid.NewString(),
		input.Name,
		input.Description,
	))
	if isUniqueViolation(err) {
		return Item{}, ErrConflict
	}
	if err != nil {
		return Item{}, fmt.Errorf("create item: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	id string,
	input UpdateItemRequest,
) (Item, error) {
	const query = `
		UPDATE items
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at`

	item, err := scanItem(r.db.QueryRow(ctx, query, id, input.Name, input.Description))
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return Item{}, ErrConflict
	}
	if err != nil {
		return Item{}, fmt.Errorf("update item: %w", err)
	}

	return item, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM items WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanItem(row rowScanner) (Item, error) {
	var item Item
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
