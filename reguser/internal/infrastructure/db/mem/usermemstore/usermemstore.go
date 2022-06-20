package usermemstore

import (
	"context"
	"database/sql"
	"goback1/lesson7/reguser/internal/entities/userentity"
	"goback1/lesson7/reguser/internal/usecases/app/repos/userrepo"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var _ userrepo.UserStore = &Users{}

type Users struct {
	sync.Mutex
	m map[uuid.UUID]userentity.User
}

func NewUsers() *Users {
	return &Users{
		m: make(map[uuid.UUID]userentity.User),
	}
}

func (us *Users) Create(ctx context.Context, u userentity.User) (*uuid.UUID, error) {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	us.m[u.ID] = u
	return &u.ID, nil
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*userentity.User, error) {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	u, ok := us.m[uid]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

// не возвращает ошибку если не нашли
func (us *Users) Delete(ctx context.Context, uid uuid.UUID) error {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(us.m, uid)
	return nil
}

func (us *Users) SearchUsers(ctx context.Context, s string) (chan userentity.User, error) {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// FIXME: переделать на дерево остатков

	chout := make(chan userentity.User, 100)

	go func() {
		defer close(chout)
		us.Lock()
		defer us.Unlock()
		for _, u := range us.m {
			if strings.Contains(u.Name, s) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					return
				case chout <- u:
				}
			}
		}
	}()

	return chout, nil
}
