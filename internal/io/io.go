package io

import "github.com/gopub/log"

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}
