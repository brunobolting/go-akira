package entity

import (
	"crypto/rand"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	minSleepMs = 100
	maxSleepMs = 400
)

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func ComparePassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetRandomSleep() time.Duration {
	n, err := rand.Int(rand.Reader, big.NewInt(maxSleepMs-minSleepMs))
	if err != nil {
		return time.Duration(minSleepMs) * time.Millisecond
	}
	return time.Duration(n.Int64()+minSleepMs) * time.Millisecond
}
