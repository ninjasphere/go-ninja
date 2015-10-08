package cloud

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("cloud")

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
	idPrefix  string
	apiPrefix string
	clientId  string
}

var cloudInstance = cloud{
	idPrefix:  config.String("https://id.sphere.ninja", "cloud.id"),
	apiPrefix: config.String("https://api.sphere.ninja", "cloud.api"),
	clientId:  config.String("0u2jota2o1dlou72hot4", "cloud.client_id"),
}

func CloudAPI() Cloud {
	return &cloudInstance
}

func (c *cloud) RegisterUser(name string, email string, password string) error {

	data := map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
	}

	if buffer, err := json.Marshal(data); err != nil {
		return err
	} else {

		if req, err := http.NewRequest("POST", c.idPrefix+"/auth/register", bytes.NewBuffer(buffer)); err != nil {
			return err
		} else {
			req.Header["Content-Type"] = []string{"application/json"}

			resp, err := getClient().Do(req)
			if err == nil {
				data, err := decodeData(resp)
				if err != nil {
					return err
				}
				if ok, present := data["ok"].(bool); present {
					if !ok {
						return fmt.Errorf("failed - %v", data["why"])
					} else {
						return nil
					}
				} else {
					return fmt.Errorf("empty response")
				}
			}
			return err
		}
	}
}

func (c *cloud) AuthenticateUser(email string, password string) (string, error) {

	data := map[string]string{
		"grant_type": "password",
		"username":   email,
		"password":   password,
		"client_id":  c.clientId,
	}

	if buffer, err := json.Marshal(data); err != nil {
		return "", err
	} else {

		if req, err := http.NewRequest("POST", c.idPrefix+"/oauth/token", bytes.NewBuffer(buffer)); err != nil {
			return "", err
		} else {
			req.Header["Content-Type"] = []string{"application/json"}

			if resp, err := getClient().Do(req); err != nil {
				return "", err
			} else {
				data, err := decodeData(resp)
				if err != nil {
					return "", err
				}
				if token, ok := data["access_token"].(string); !ok {
					e := data["error"]
					d := data["error_description"]
					return "", fmt.Errorf("%s(\"%s\")", e, d)
				} else {
					return token, nil
				}
			}
		}
	}
}

func (c *cloud) ActivateSphere(accessToken string, nodeId string) error {
	data := map[string]interface{}{
		"nodeId": nodeId,
	}

	if buffer, err := json.Marshal(data); err != nil {
		return err
	} else {

		if req, err := http.NewRequest("POST", c.apiPrefix+"/rest/v1/node", bytes.NewBuffer(buffer)); err != nil {
			return err
		} else {
			req.Header["Content-Type"] = []string{"application/json"}
			req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", accessToken)}

			if resp, err := getClient().Do(req); err != nil {
				return err
			} else {
				data, err := decodeData(resp)
				if err != nil {
					return err
				}
				if e, ok := data["type"].(string); ok && e == "error" {
					if data, ok := data["data"].(map[string]interface{}); ok {
						return fmt.Errorf("%s", data["message"])
					} else {
						return fmt.Errorf("failed unknown message: %+v", data)
					}
				}

				return nil
			}
		}
	}
}

func getClient() *http.Client {
	client := &http.Client{}

	if config.Bool(false, "cloud", "allowSelfSigned") {
		log.Warningf("Allowing self-signed cerificate (should only be used to connect to development cloud)")
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}

func decodeData(resp *http.Response) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	copy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	err = json.NewDecoder(bytes.NewBuffer(copy)).Decode(&data)
	if err != nil {
		return data, fmt.Errorf("failed to decode response: %s", string(copy))
	} else {
		return data, nil
	}
}
