package bonsai_test

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/stretchr/testify/require"
)

func init() {
	initLogger()
}

func initLogger() {
	// https://github.com/golang/go/issues/62403
	// https://cs.opensource.google/go/x/exp/+/master:slog/handler.go;l=442

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				src := a.Value.Any().(*slog.Source)
				// Ruby on Rails-ish formatting
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d:in '%s'", src.File, src.Line, src.Function))
			}
			return a
		},
	})

	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}
