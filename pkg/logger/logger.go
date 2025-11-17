package logger

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

func Init(debug bool) zerolog.Logger {
	var zlog zerolog.Logger
	zerolog.TimestampFieldName = "Log time"
	zerolog.LevelFieldName = "lvl"
	zerolog.CallerMarshalFunc = func(_ uintptr, filename string, line int) string {
		shortname := filename
		for i := len(filename) - 1; i > 0; i-- {
			if filename[i] == '/' {
				shortname = filename[i+1:]
				break
			}
		}
		filename = shortname
		return filename + ":" + strconv.Itoa(line)
	}
	zerolog.CallerFieldName = "call"

	if debug {
		zlog = zerolog.
			New(os.Stdout).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger().
			Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		zlog = zerolog.
			New(os.Stdout).
			Level(zerolog.InfoLevel).
			With().
			Timestamp().
			Logger()
	}
	return zlog
}
