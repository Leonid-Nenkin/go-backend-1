package userfilemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"goback1/lesson7/reguser/internal/entities/userentity"
	"goback1/lesson7/reguser/internal/infrastructure/db/files/usereventstore"
	"goback1/lesson7/reguser/internal/infrastructure/db/files/usermemstate"
	"goback1/lesson7/reguser/internal/usecases/app/repos/userrepo"
	"os"
	"time"

	"github.com/google/uuid"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

var _ userrepo.UserStore = &Users{}

type Users struct {
	uf    *usereventstore.UserFile
	topic *pubsub.Topic
	ums   *usermemstate.Users
}

// "mem://topicA"
func NewUsers(eventfn, topicUrl string) (*Users, error) {
	topic, err := pubsub.OpenTopic(context.Background(), topicUrl)
	if err != nil {
		return nil, err
	}
	ums, err := usermemstate.NewUsers(topicUrl)
	if err != nil {
		return nil, err
	}

	uf, err := usereventstore.NewUserFile(eventfn, usereventstore.Play)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if err == nil {
		uf.PlayEvents(func(e *usereventstore.Event) {
			switch e.Type {
			case usereventstore.EventCreate:
				se := usermemstate.StateEvent{
					User: usermemstate.StateUser{
						ID:          e.User.ID,
						Name:        e.User.Name,
						Data:        e.User.Data,
						Permissions: e.User.Permissions,
					},
					Event: usermemstate.EventCreate,
				}
				sendTopic(context.Background(), topic, se)
			case usereventstore.EventDelete:
				se := usermemstate.StateEvent{
					User: usermemstate.StateUser{
						ID: e.User.ID,
					},
					Event: usermemstate.EventDelete,
				}
				sendTopic(context.Background(), topic, se)
			}
		})
		uf.Close()
	}

	uf, err = usereventstore.NewUserFile(eventfn, usereventstore.Append)
	if err != nil {
		return nil, err
	}
	s := &Users{
		uf:    uf,
		topic: topic,
		ums:   ums,
	}
	return s, nil
}

func (us *Users) Close() {
	us.topic.Shutdown(context.Background())
	us.uf.Close()
	us.ums.Close()
}

func (us *Users) Create(ctx context.Context, u userentity.User) (*uuid.UUID, error) {
	ev := usereventstore.Event{
		TimeStamp: time.Now(),
		Type:      usereventstore.EventCreate,
		User: &usereventstore.EventUser{
			ID:          u.ID,
			Name:        u.Name,
			Data:        u.Data,
			Permissions: u.Permissions,
		},
	}
	if err := us.uf.SaveEvent(ev); err != nil {
		return nil, err
	}

	se := usermemstate.StateEvent{
		User: usermemstate.StateUser{
			ID:          u.ID,
			Name:        u.Name,
			Data:        u.Data,
			Permissions: u.Permissions,
		},
		Event: usermemstate.EventCreate,
	}
	sendTopic(ctx, us.topic, se)

	uid := u.ID
	return &uid, nil
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*userentity.User, error) {
	stu, err := us.ums.Read(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &userentity.User{
		ID:          stu.ID,
		Name:        stu.Name,
		Data:        stu.Data,
		Permissions: stu.Permissions,
	}, nil
}

func (us *Users) Delete(ctx context.Context, uid uuid.UUID) error {
	ev := usereventstore.Event{
		TimeStamp: time.Now(),
		Type:      usereventstore.EventDelete,
		User: &usereventstore.EventUser{
			ID: uid,
		},
	}
	if err := us.uf.SaveEvent(ev); err != nil {
		return err
	}

	sendTopic(ctx, us.topic, usermemstate.StateEvent{
		User: usermemstate.StateUser{
			ID: uid,
		},
		Event: usermemstate.EventDelete,
	})

	return nil
}

func sendTopic(ctx context.Context, topic *pubsub.Topic, se usermemstate.StateEvent) {
	b, _ := json.Marshal(se)
	msg := &pubsub.Message{
		LoggableID: uuid.NewString(),
		Body:       b,
	}
	if err := topic.Send(ctx, msg); err != nil {
		fmt.Printf("topic send error: %v", err)
	}
}

func (us *Users) SearchUsers(ctx context.Context, s string) (chan userentity.User, error) {
	chout := make(chan userentity.User, 100)
	chin, err := us.ums.SearchUsers(ctx, s)
	if err != nil {
		return nil, err
	}
	go func() {
		for stu := range chin {
			chout <- userentity.User{
				ID:          stu.ID,
				Name:        stu.Name,
				Data:        stu.Data,
				Permissions: stu.Permissions,
			}
		}
	}()
	return chout, nil
}
