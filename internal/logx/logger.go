package logx

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Options describes logger configuration inputs.
type Options struct {
	Level     string // debug, info, warn, error, fatal
	Format    string // json | console
	TimeFormat string // RFC3339, unix, or Go layout
}

// New returns a configured zerolog.Logger based on options.
func New(opts Options) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(opts.Level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimestampFieldName = "time"
	switch strings.ToLower(opts.TimeFormat) {
	case "rfc3339":
		zerolog.TimeFieldFormat = time.RFC3339
	case "rfc3339nano":
		zerolog.TimeFieldFormat = time.RFC3339Nano
	case "unix":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	case "unix_ms", "unixms":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	case "unix_us", "unixus":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	case "unix_ns", "unixns":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixNano
	case "":
		zerolog.TimeFieldFormat = time.RFC3339
	default:
		zerolog.TimeFieldFormat = opts.TimeFormat
	}

	var out io.Writer = os.Stdout
	if strings.ToLower(opts.Format) == "console" {
		cw := zerolog.ConsoleWriter{Out: os.Stdout}
		cw.TimeFormat = zerolog.TimeFieldFormat
		out = cw
	}

	return zerolog.New(out).Level(lvl).With().Timestamp().Logger()
}
