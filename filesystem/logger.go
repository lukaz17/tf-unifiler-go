package filesystem

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger = log.Logger

func SetLogger(l zerolog.Logger) {
	logger = l
}
