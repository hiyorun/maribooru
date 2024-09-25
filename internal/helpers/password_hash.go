package helpers

import "golang.org/x/crypto/bcrypt"

func PasswordHash(passwd string) (string, error) {
	passwordBytes := []byte(passwd)
	hashedPasswordBytes, err := bcrypt.
		GenerateFromPassword(passwordBytes, bcrypt.MinCost)

	if err := bcrypt.CompareHashAndPassword(hashedPasswordBytes, passwordBytes); err != nil {
		return "", err
	}
	return string(hashedPasswordBytes), err
}
