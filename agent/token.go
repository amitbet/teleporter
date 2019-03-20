package agent

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/amitbet/teleporter/logger"
)

// func main() {
// 	token, _ := GenerateRandomString(32)
// 	fmt.Println(token)
// }

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) string {
	b, err := GenerateRandomBytes(s)
	if err != nil {
		logger.Error("problem in rand token gen: ", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
