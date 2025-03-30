package entity

import "errors"

var ErrCrawlerAlreadyRunning = errors.New("crawler is already running")

var ErrCrawlerNotRunning = errors.New("crawler is not running")

var ErrCrawlerCannotBeCancelled = errors.New("crawler cannot be cancelled")
