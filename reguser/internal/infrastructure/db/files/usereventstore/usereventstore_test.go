package usereventstore

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestAppendPlay(t *testing.T) {
	f, _ := os.CreateTemp("", "*.json")
	tfn := f.Name()
	t.Log(f.Name())

	json.NewEncoder(f).Encode(Event{
		TimeStamp: time.Now(),
		Type:      EventCreate,
	})
	json.NewEncoder(f).Encode(Event{
		TimeStamp: time.Now(),
		Type:      EventDelete,
	})
	json.NewEncoder(f).Encode(Event{
		TimeStamp: time.Now(),
		Type:      EventCreate,
	})

	f.Close()

	uf, err := NewUserFile(tfn, Play)
	if err != nil {
		t.Error(err)
		return
	}
	cnt := 0
	uf.PlayEvents(func(e *Event) {
		t.Log(e.TimeStamp)
		cnt++
	})
	if cnt != 3 {
		t.Error("cnt != 3")
	}
	uf.Close()

	uf, err = NewUserFile(tfn, Append)
	if err != nil {
		t.Error(err)
		return
	}
	if err := uf.SaveEvent(Event{
		TimeStamp: time.Now(),
		Type:      EventCreate,
	}); err != nil {
		t.Error(err)
		return
	}
	uf.Close()

	uf, err = NewUserFile(tfn, Play)
	if err != nil {
		t.Error(err)
		return
	}
	cnt = 0
	uf.PlayEvents(func(e *Event) {
		t.Log(e.TimeStamp)
		cnt++
	})
	if cnt != 4 {
		t.Error("cnt != 4")
	}
	uf.Close()
	// t.Error("")
}
