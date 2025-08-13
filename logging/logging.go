package logging

import (
	"context"
	"log/slog"
	"os"
)

var (
	Logger   *slog.Logger
	debugFile *os.File
)

// Init beállítja a console handler-t Info szinttől, és opcionálisan debug fájl handler-t.
func Init(debugMode bool) {
	// 1) Konzolos handler Info és fölötti szintekhez
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// 2) Opcionális fájl handler Debug szinthez
	var handlers []slog.Handler
	handlers = append(handlers, consoleHandler)

	if debugMode {
		var err error
		debugFile, err = os.OpenFile("logs/debug.json",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			panic("Nem lehet megnyitni a debug log fájlt: " + err.Error())
		}

		// Csak a Debug szintet engedélyezzük ebben a handlerben
		jsonHandler := slog.NewJSONHandler(debugFile, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		// wrapoljuk, hogy ténylegesen csak Debug üzenetek menjenek ide
		debugOnly := &debugFilter{inner: jsonHandler}
		handlers = append(handlers, debugOnly)
	}

	// 3) Composite handler összerakása
	root := &multiHandler{handlers: handlers}
	Logger = slog.New(root)
}

// Close bezárja a debug fájlt (ha megnyitottuk)
func Close() error {
	if debugFile != nil {
		return debugFile.Close()
	}
	return nil
}

// Fatal kiír egy hibát Error szinten, majd kilép a programból.
func Fatal(msg string, args ...any) {
	if Logger != nil {
		Logger.Error("FATAL: " + msg, args...)
	} else {
		// Biztonsági fallback, ha Logger még nincs inicializálva
		slog.Default().Error("FATAL: " + msg, args...)
	}
	os.Exit(1)
}

// debugFilter csak a Debug szintű rekordokat engedi át
type debugFilter struct {
	inner slog.Handler
}

func (f *debugFilter) Enabled(ctx context.Context, level slog.Level) bool {
	return level == slog.LevelDebug
}

func (f *debugFilter) Handle(ctx context.Context, rec slog.Record) error {
	if rec.Level == slog.LevelDebug {
		return f.inner.Handle(ctx, rec)
	}
	return nil
}

func (f *debugFilter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &debugFilter{inner: f.inner.WithAttrs(attrs)}
}

func (f *debugFilter) WithGroup(name string) slog.Handler {
	return &debugFilter{inner: f.inner.WithGroup(name)}
}