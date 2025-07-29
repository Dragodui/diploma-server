package security

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(b), err
}

func ComparePasswords(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
