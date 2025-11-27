package utils

import (
	"errors"
	"strconv"
	"strings"
)

// ParseToken untuk token berbentuk "<userID>|<hash>"
func ParseToken(token string) (uint, error) {
	parts := strings.Split(token, "|")
	if len(parts) < 1 {
		return 0, errors.New("format token salah")
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}
