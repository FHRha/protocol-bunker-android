package main

import (
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func makeMojibake(value string) string {
	encoded, err := charmap.Windows1251.NewDecoder().Bytes([]byte(value))
	if err != nil {
		return value
	}
	return string(encoded)
}

func TestSanitizeHumanText_FixesCommonMojibake(t *testing.T) {
	got := sanitizeHumanText(makeMojibake("Профессия"), "")
	if got != "Профессия" {
		t.Fatalf("expected repaired text, got %q", got)
	}

	got = sanitizeHumanText(makeMojibake("Сейчас нет голосования."), "")
	if got != "Сейчас нет голосования." {
		t.Fatalf("expected repaired sentence, got %q", got)
	}
}

func TestSanitizeHumanText_LeavesNormalText(t *testing.T) {
	input := "Обязательная категория: Профессия"
	got := sanitizeHumanText(input, "")
	if got != input {
		t.Fatalf("expected unchanged text, got %q", got)
	}
}

func TestSanitizeHumanText_FallbackWhenUnrecoverable(t *testing.T) {
	got := sanitizeHumanText("\uFFFD\uFFFD\uFFFD\uFFFD", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback text, got %q", got)
	}
}
