package log

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewAndWrite(t *testing.T) {
	var fileName = "app.log"
	var LogPath = "logs"
	logs, err := New(Config{
		Level:      "debug",
		Format:     "json",
		LogPath:    LogPath,
		Filename:   fileName,
		MaxSize:    1,
		MaxBackups: 2,
		MaxAge:     1,
		Compress:   false,
		Console:    true,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logs.Close()
	logs.Info("hello", zap.String("key", "value"))
	logs.Error("error", zap.String("key", "value"))
	logs.Errorf("error %s", "value")
	logs.Fatal("fatal", zap.String("key", "value"))
	logs.Panic("panic", zap.String("key", "value"))

}

func TestParseLevel(t *testing.T) {
	tests := map[string]string{
		"debug": "debug",
		"INFO":  "info",
		"warn":  "warn",
	}

	for input, want := range tests {
		lv, err := parseLevel(input)
		if err != nil {
			t.Fatalf("parseLevel(%q) error = %v", input, err)
		}
		if lv.String() != want {
			t.Fatalf("parseLevel(%q) = %s, want %s", input, lv, want)
		}
	}
}
