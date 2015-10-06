package cloud

import (
	"errors"
	"fmt"
)

var (
	AlreadyRegistered = errors.New("userid is already registered")
)

/**
 * Registers a user.
 */
func RegisterUser(name string, email string, password string) error {
	return fmt.Errorf("not implemented: RegisterUser")
}

/**
 * Authenticates the user with the e-mail and password supplied, returns
 * a token that can be used with the activate call.
 */
func AuthenticateUser(email string, password string) (string, error) {
	return "", fmt.Errorf("not implemented: AuthenticateUser")
}

/**
 * Change the password associated with the current user.
 */
func ChangePassword(accessToken string, newPassword string) error {
	return fmt.Errorf("not implemented: ChangePassword")
}

/**
 * Activates the specified sphere.
 */
func ActivateSphere(accessToken string, nodeId string) error {
	return fmt.Errorf("not implemented: ActivateSphere")
}
