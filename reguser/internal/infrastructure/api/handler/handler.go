package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"goback1/lesson7/reguser/internal/entities/userentity"
	"goback1/lesson7/reguser/internal/usecases/app/repos/userrepo"

	"github.com/google/uuid"
)

type Handlers struct {
	us *userrepo.Users
}

func NewHandlers(us *userrepo.Users) *Handlers {
	r := &Handlers{
		us: us,
	}
	return r
}

type User struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Data       string    `json:"data"`
	Permission int       `json:"perms"`
}

func (rt *Handlers) CreateUser(ctx context.Context, u User) (User, error) {
	bu := userentity.User{
		Name: u.Name,
		Data: u.Data,
	}

	nbu, err := rt.us.Create(ctx, bu)
	if err != nil {
		return User{}, fmt.Errorf("error when creating: %w", err)
	}

	return User{
		ID:         nbu.ID,
		Name:       nbu.Name,
		Data:       nbu.Data,
		Permission: nbu.Permissions,
	}, nil
}

var ErrUserNotFound = errors.New("user not found")

// read?uid=...
func (rt *Handlers) ReadUser(ctx context.Context, uid uuid.UUID) (User, error) {
	if (uid == uuid.UUID{}) {
		return User{}, fmt.Errorf("bad request: uid is empty")
	}

	nbu, err := rt.us.Read(ctx, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("error when reading: %w", err)
	}

	return User{
		ID:         nbu.ID,
		Name:       nbu.Name,
		Data:       nbu.Data,
		Permission: nbu.Permissions,
	}, nil
}

func (rt *Handlers) DeleteUser(ctx context.Context, uid uuid.UUID) (User, error) {
	if (uid == uuid.UUID{}) {
		return User{}, fmt.Errorf("bad request: uid is empty")
	}

	nbu, err := rt.us.Delete(ctx, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("error when reading: %w", err)
	}

	return User{
		ID:         nbu.ID,
		Name:       nbu.Name,
		Data:       nbu.Data,
		Permission: nbu.Permissions,
	}, nil
}

// /search?q=...
func (rt *Handlers) SearchUser(ctx context.Context, q string, f func(User) error) error {
	ch, err := rt.us.SearchUsers(ctx, q)
	if err != nil {
		return fmt.Errorf("error when reading: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case u, ok := <-ch:
			if !ok {
				return nil
			}
			if err := f(User{
				ID:         u.ID,
				Name:       u.Name,
				Data:       u.Data,
				Permission: u.Permissions,
			}); err != nil {
				return err
			}
		}
	}
}
