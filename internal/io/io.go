package io

import "github.com/gopub/log/v2"

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}
