package usermemstate

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

type StateUser struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name,omitempty"`
	Data        string    `json:"data,omitempty"`
	Permissions int       `json:"perms,omitempty"`
}

type EventType string

const (
	EventCreate EventType = "create"
	EventDelete EventType = "delete"
)

type StateEvent struct {
	User  StateUser `json:"user"`
	Event EventType `json:"eventType"`
}

type Users struct {
	sync.Mutex
	m    map[uuid.UUID]StateUser
	subs *pubsub.Subscription
	cf   context.CancelFunc
}

// "mem://topicA"
func NewUsers(topicUrl string) (*Users, error) {
	subs, err := pubsub.OpenSubscription(context.Background(), topicUrl)
	if err != nil {
		return nil, err
	}

	us := &Users{
		m:    make(map[uuid.UUID]StateUser),
		subs: subs,
	}

	ctx, cf := context.WithCancel(context.Background())

	us.cf = cf
	go us.listen(ctx)

	return us, nil
}

func (us *Users) Close() {
	us.cf()
	us.subs.Shutdown(context.Background())
}

func (us *Users) listen(ctx context.Context) {
	for {
		msg, err := us.subs.Receive(ctx)
		if err != nil {
			// Errors from Receive indicate that Receive will no longer succeed.
			log.Printf("Receiving message error: %v", err)
			break
		}
		// Do work based on the message, for example:
		fmt.Printf("Got message: %q\n", msg.Body)

		se := &StateEvent{}
		if err := json.Unmarshal(msg.Body, se); err != nil {
			log.Printf("StateEvent message unmarshal error: %v", err)
		} else {
			us.Lock()

			switch se.Event {
			case EventCreate:
				us.m[se.User.ID] = se.User
			case EventDelete:
				delete(us.m, se.User.ID)
			}

			us.Unlock()
		}
		// Messages must always be acknowledged with Ack.
		msg.Ack()
	}
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*StateUser, error) {
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

func (us *Users) SearchUsers(ctx context.Context, s string) (chan StateUser, error) {
	us.Lock()
	defer us.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	chout := make(chan StateUser, 100)

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
