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

/**
 * Answers the userid used by RegisterSphere.
 *
 * Will typically be a string of the form sphere+{nodeId}@ninjablocks.com but
 * this can be varied by configuration.
 */
func SphereAutoRegisterUserId(nodeId string) string {
	return fmt.Sprintf("sphere+%s@ninjablocks.com", nodeId)
}

/**
 * Registers the current sphere with a unique userid for the sphere
 * which is generated from the sphere serial id.
 *
 * The password is only required on a second attempt to auto-register the sphere.
 *
 * This performs the equivalent of:
 *
 * - RegisterUser("Default User", SphereAutoRegisterUserId("serial"), "{random-password}")
 * - token = AuthenticateUser(SphereAutoRegisterUserId("serial"), "{random-password}")
 * - ActivateSphere(token, "{serial}")
 */
func RegisterSphere(password *string) error {
	return fmt.Errorf("not implemented: RegisterSphere")
}
