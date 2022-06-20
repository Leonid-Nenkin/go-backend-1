package userentity

import (
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Name        string
	Data        string
	Permissions int
}
