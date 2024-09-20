package helpers

import "golang.org/x/crypto/bcrypt"

func PasswordHash(passwd string) (string, error) {
	passwordBytes := []byte(passwd)
	hashedPasswordBytes, err := bcrypt.
		GenerateFromPassword(passwordBytes, bcrypt.MinCost)
	return string(hashedPasswordBytes), err
}
