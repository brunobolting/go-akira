package entity

import "errors"

var ErrBookNameInvalid = errors.New("error.book.invalid-name")

var ErrBookNameTooLong = errors.New("error.book.name-too-long")
