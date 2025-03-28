package entity

import "errors"

var ErrCollectionNameInvalid = errors.New("error.collection.invalid-name")

var ErrCollectionNameTooLong = errors.New("error.collection.name-too-long")
