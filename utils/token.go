package utils

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
)

// GenerateSimpleToken menggunakan random bytes + userID
func GenerateSimpleToken(nomorHP string, userID uint) string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return fmt.Sprintf("%d|%s|%d", userID, hex.EncodeToString(b), time.Now().UnixNano())
}
