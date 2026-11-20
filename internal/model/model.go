package model

import (
	"encoding/base64"
	"strings"
)

type Installation struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

func (i Installation) String() string   { return "telemetry installation data (redacted)" }
func (i Installation) GoString() string { return i.String() }

func (i Installation) IsValid() bool {
	isValid, id := parseToken(i.Token)
	return isValid && id == i.ID
}

func parseToken(token string) (isValid bool, id string) {
	if !strings.HasPrefix(token, "et_") {
		return false, ""
	}
	id, secret, ok := strings.Cut(strings.TrimPrefix(token, "et_"), ".")
	if !ok || !isValidEncoded(id, 16) || !isValidEncoded(secret, 32) {
		return false, ""
	}
	return true, id
}

func isValidEncoded(value string, size int) bool {
	if len(value) != base64.RawURLEncoding.EncodedLen(size) {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	return err == nil && len(decoded) == size
}
