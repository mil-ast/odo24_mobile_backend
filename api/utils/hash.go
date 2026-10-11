package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/argon2"
)

const saltSize = 16

// Флаги/параметры аргонид хэша
const (
	hashFormatVersion byte   = 1
	argon2Time        uint32 = 3
	argon2Memory      uint32 = 64 * 1024 // 64 MB
	argon2Threads     uint8  = 2
	argon2KeyLen      uint32 = 32
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	return salt, err
}

// GetPasswordHash возвращает хэш пароля по argon2id с префиксом версии.
func GetPasswordHash(password []byte, salt []byte) ([]byte, error) {
	hash := argon2.IDKey(password, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	out := make([]byte, 1+len(hash))
	out[0] = hashFormatVersion
	copy(out[1:], hash)
	return out, nil
}

// IsLegacyHash определяет, является ли слепок "старым" sha256(salt||password).
func IsLegacyHash(storedHash []byte) bool {
	return len(storedHash) != 1+int(argon2KeyLen) || storedHash[0] != hashFormatVersion
}

// VerifyPassword проверяет пароль. Если слепок в старом формате (sha256),
// делает fallback. Возвращает совпадение и признак legacy-формата.
func VerifyPassword(password, salt, storedHash []byte) (bool, bool, error) {
	if IsLegacyHash(storedHash) {
		h := sha256.New()
		_, err := h.Write(salt)
		if err != nil {
			return false, false, err
		}
		_, err = h.Write(password)
		if err != nil {
			return false, false, err
		}
		return bytes.Equal(h.Sum(nil), storedHash), true, nil
	}

	hash := argon2.IDKey(password, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return bytes.Equal(storedHash[1:], hash), false, nil
}