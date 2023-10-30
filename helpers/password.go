package helpers

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func ComparePassword(password string, hashed string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hashed))
	if err != nil {
		return false, err
	}
	return true, nil
}