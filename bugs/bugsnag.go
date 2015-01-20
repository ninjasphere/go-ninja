// Package bugs provides a wrapper around bugsnag so additional ninja specific stuff
// can be added to the configuration without duplicating it in every driver.
package bugs

import (
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/wolfeidau/bugsnag-go"
)

// Configure our bug tracker using the environment and key supplied
func Configure(env, key string) {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       key,
		ReleaseStage: env,
		Logger:       logger.GetBugsnagLogger("bugsnag"),
	})

}
