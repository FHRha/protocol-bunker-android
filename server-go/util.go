package main

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func randomToken(bytesLen int) string {
	if bytesLen <= 0 {
		bytesLen = 8
	}
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "fallback-token"
	}
	return hex.EncodeToString(buf)
}

func envFlag(value string) bool {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeIdentityMode(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "dev_tab":
		return "dev_tab"
	default:
		return "prod"
	}
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func resolvePath(primary string, fallback string) string {
	if primary != "" {
		if _, err := os.Stat(primary); err == nil {
			return primary
		}
	}
	if fallback != "" {
		if _, err := os.Stat(fallback); err == nil {
			return fallback
		}
	}
	if primary != "" {
		return primary
	}
	return fallback
}

func absPath(path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return abs
}

func firstLANIPv4() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	best := ""
	bestScore := -1
	for _, iface := range interfaces {
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}
		if (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		lowerName := strings.ToLower(iface.Name)
		blockedName := strings.Contains(lowerName, "docker") ||
			strings.Contains(lowerName, "veth") ||
			strings.Contains(lowerName, "vpn") ||
			strings.Contains(lowerName, "tun") ||
			strings.Contains(lowerName, "tap") ||
			strings.Contains(lowerName, "virtual")
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			ipStr := ip.String()
			if strings.HasPrefix(ipStr, "127.") || strings.HasPrefix(ipStr, "169.254.") {
				continue
			}

			score := 5
			if isPrivateLAN(ipStr) {
				score += 100
			}
			if !blockedName {
				score += 20
			}
			if score > bestScore {
				bestScore = score
				best = ipStr
			}
		}
	}

	if best == "" {
		return "127.0.0.1"
	}
	return best
}

func isPrivateLAN(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	return parsed.IsPrivate()
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func looksLikeMojibake(text string) bool {
	value := strings.TrimSpace(text)
	if value == "" {
		return false
	}

	return mojibakeScore(value) >= 3
}

func mojibakeScore(value string) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}

	// Common markers for broken UTF-8/CP1251 round-trips observed in logs and events.
	markers := []string{
		"\u0420\u00A0",
		"\u0420\u040E",
		"\u0420\u040B",
		"\u0420\u0406\u0420\u201A",
		"\u0432\u20AC",
		"\u00D0",
		"\u00D1",
	}
	hits := 0
	for _, marker := range markers {
		hits += strings.Count(trimmed, marker)
	}

	hits += strings.Count(trimmed, "\uFFFD") * 3
	rCount := strings.Count(trimmed, "\u0420")
	sCount := strings.Count(trimmed, "\u0421")
	runeCount := utf8.RuneCountInString(trimmed)
	if runeCount > 0 {
		rsRatio := float64(rCount+sCount) / float64(runeCount)
		if rCount+sCount >= 4 && rsRatio >= 0.18 {
			hits += 2
		}
	}
	return hits
}

func tryFixMojibake(text string) string {
	value := strings.TrimSpace(text)
	if value == "" {
		return ""
	}

	encoded, err := charmap.Windows1251.NewEncoder().Bytes([]byte(value))
	if err != nil || len(encoded) == 0 || !utf8.Valid(encoded) {
		return value
	}
	fixed := strings.TrimSpace(string(encoded))
	if fixed == "" || fixed == value {
		return value
	}
	if mojibakeScore(fixed) >= mojibakeScore(value) {
		return value
	}
	return fixed
}

func sanitizeHumanText(text string, fallback string) string {
	value := strings.TrimSpace(text)
	if value == "" {
		return fallback
	}
	if looksLikeMojibake(value) {
		fixed := tryFixMojibake(value)
		if fixed != "" && !looksLikeMojibake(fixed) {
			return fixed
		}
		if fallback != "" {
			return fallback
		}
		return "Text unavailable."
	}
	// Secondary pass: attempt recovery for subtle cp1251 mojibake even if the score is low.
	fixed := tryFixMojibake(value)
	if fixed != "" && !looksLikeMojibake(fixed) && utf8.ValidString(fixed) {
		return fixed
	}
	return value
}

func defaultEventMessage(kind string) string {
	switch kind {
	case "votingStart":
		return "Voting started."
	case "elimination":
		return "A player was eliminated."
	case "playerLeftBunker":
		return "A player left the bunker."
	case "playerDisconnected":
		return "A player disconnected."
	case "playerReconnected":
		return "A player reconnected."
	default:
		return "Game state updated."
	}
}

func defaultStageMessage(round int) string {
	roundSafe := round
	if roundSafe < 1 {
		roundSafe = 1
	}
	return "Round " + strconv.Itoa(roundSafe) + ": state updated."
}
