package usereventstore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type UserFile struct {
	fname string
	f     *os.File
	op    OpenOption
	enc   *json.Encoder
	dec   *json.Decoder
}

type OpenOption int

const (
	_ OpenOption = iota
	Play
	Append
)

func NewUserFile(fn string, op OpenOption) (*UserFile, error) {
	var err error
	s := &UserFile{
		fname: fn,
		op:    op,
	}
	switch op {
	case Play:
		s.f, err = os.OpenFile(fn, os.O_RDONLY, 0750)
		if err == nil {
			s.dec = json.NewDecoder(s.f)
		}
	case Append:
		s.f, err = os.OpenFile(fn, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0750)
		if err == nil {
			s.enc = json.NewEncoder(s.f)
		}
	default:
		return nil, fmt.Errorf("option invalid")
	}
	return s, err
}

func (uf *UserFile) Close() {
	if uf.f != nil {
		uf.f.Close()
	}
}

func (uf *UserFile) SaveEvent(e Event) error {
	if uf.op != Append {
		return fmt.Errorf("not for append")
	}
	if err := uf.enc.Encode(e); err != nil {
		return err
	}
	if _, err := uf.f.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func (uf *UserFile) PlayEvents(f func(*Event)) error {
	if uf.op != Play {
		return fmt.Errorf("not for playing")
	}
	for {
		e := &Event{}
		if err := uf.dec.Decode(e); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		f(e)
	}
}
