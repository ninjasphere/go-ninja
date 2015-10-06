package cloud

import (
	"errors"
	"fmt"
)

var (
	AlreadyRegistered = errors.New("userid is already registered")
)

type Cloud interface {
	/**
	 * Registers a user.
	 */
	RegisterUser(name string, email string, password string) error

	/**
	 * Authenticates the user with the e-mail and password supplied, returns
	 * a token that can be used with the activate call.
	 */
	AuthenticateUser(email string, password string) (string, error)

	/**
	 * Activates the specified sphere.
	 */
	ActivateSphere(accessToken string, nodeId string) error
}

type cloud struct {
}

var cloudInstance cloud

func CloudAPI() Cloud {
	return &cloudInstance
}

func (c *cloud) RegisterUser(name string, email string, password string) error {
	return fmt.Errorf("not implemented: RegisterUser")
}

func (c *cloud) AuthenticateUser(email string, password string) (string, error) {
	return "", fmt.Errorf("not implemented: AuthenticateUser")
}

func (c *cloud) ActivateSphere(accessToken string, nodeId string) error {
	return fmt.Errorf("not implemented: ActivateSphere")
}
