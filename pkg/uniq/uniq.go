package uniq

import "crypto/rand"

const base64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-"

// Generate returns a 60-bit random number encoded as a 10-character base64 encoded
// string.
func Uniq() string {
	bytes := make([]byte, 10)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for k, b := range bytes {
		bytes[k] = base64[b & 0b111111]
	}
	return string(bytes)
}
