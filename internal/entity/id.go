package entity

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func NewID() string {
	return generateULID()
}

func generateULID() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
