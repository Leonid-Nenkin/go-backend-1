package pgstore

import (
	"context"
	"database/sql"
	"goback1/lesson7/reguser/internal/entities/userentity"
	"goback1/lesson7/reguser/internal/usecases/app/repos/userrepo"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // Postgresql driver
)

var _ userrepo.UserStore = &Users{}

type DBPgUser struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Name        string
	Data        string
	Permissions int
}

type Users struct {
	db *sql.DB
}

func NewUsers(dsn string) (*Users, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id uuid NOT NULL,
		created_at timestamptz NOT NULL,
		updated_at timestamptz NOT NULL,
		deleted_at timestamptz NULL,
		name varchar NOT NULL,
		"data" varchar NULL,
		perms int2 NULL,
		CONSTRAINT users_pk PRIMARY KEY (id)
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	us := &Users{
		db: db,
	}
	return us, nil
}

func (us *Users) Close() {
	us.db.Close()
}

func (us *Users) Create(ctx context.Context, u userentity.User) (*uuid.UUID, error) {
	dbu := &DBPgUser{
		ID:          u.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:        u.Name,
		Data:        u.Data,
		Permissions: u.Permissions,
	}

	_, err := us.db.ExecContext(ctx, `INSERT INTO users 
	(id, created_at, updated_at, deleted_at, name, data, perms)
	values ($1, $2, $3, $4, $5, $6, $7)`,
		dbu.ID,
		dbu.CreatedAt,
		dbu.UpdatedAt,
		nil,
		dbu.Name,
		dbu.Data,
		dbu.Permissions,
	)
	if err != nil {
		return nil, err
	}

	return &u.ID, nil
}

func (us *Users) Delete(ctx context.Context, uid uuid.UUID) error {
	_, err := us.db.ExecContext(ctx, `UPDATE users SET deleted_at = $2 WHERE id = $1`,
		uid, time.Now(),
	)
	return err
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*userentity.User, error) {
	dbu := &DBPgUser{}
	rows, err := us.db.QueryContext(ctx,
		`SELECT id, created_at, updated_at, deleted_at, name, data, perms 
	FROM users WHERE id = $1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&dbu.ID,
			&dbu.CreatedAt,
			&dbu.UpdatedAt,
			&dbu.DeletedAt,
			&dbu.Name,
			&dbu.Data,
			&dbu.Permissions,
		); err != nil {
			return nil, err
		}
	}

	return &userentity.User{
		ID:          dbu.ID,
		Name:        dbu.Name,
		Data:        dbu.Data,
		Permissions: dbu.Permissions,
	}, nil
}

func (us *Users) SearchUsers(ctx context.Context, s string) (chan userentity.User, error) {
	chout := make(chan userentity.User, 100)

	go func() {
		defer close(chout)
		dbu := &DBPgUser{}

		rows, err := us.db.QueryContext(ctx, `
		SELECT id, created_at, updated_at, deleted_at, name, data, perms 
		FROM users WHERE name LIKE $1 and deleted_at is null`, "%"+s+"%")
		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(
				&dbu.ID,
				&dbu.CreatedAt,
				&dbu.UpdatedAt,
				&dbu.DeletedAt,
				&dbu.Name,
				&dbu.Data,
				&dbu.Permissions,
			); err != nil {
				log.Println(err)
				return
			}

			chout <- userentity.User{
				ID:          dbu.ID,
				Name:        dbu.Name,
				Data:        dbu.Data,
				Permissions: dbu.Permissions,
			}
		}
	}()

	return chout, nil
}
