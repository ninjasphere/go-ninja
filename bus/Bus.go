package bus

import (
	"strings"

	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("bus")

type Bus interface {
	Publish(topic string, payload []byte)
	Subscribe(topic string, callback func(topic string, payload []byte)) (*subscription, error)
}

func MustConnect(host, id string) Bus {
	//return ConnectTinyBus(host, id)
	bus, err := NewPahoBus(host, id)
	if err != nil {
		log.HandleError(err, "Failed to connect to mqtt")
	}
	return bus
}

type subscription struct {
	topic    string
	callback func(topic string, payload []byte)
}

func matches(subscription string, topic string) bool {
	parts := strings.Split(topic, "/")
	subParts := strings.Split(subscription, "/")

	i := 0
	for i < len(parts) {
		// topic is longer, no match
		if i >= len(subParts) {
			return false
		}
		// matched up to here, and now the wildcard says "all others will match"
		if subParts[i] == "#" {
			return true
		}
		// text does not match, and there wasn't a + to excuse it
		if parts[i] != subParts[i] && subParts[i] != "+" {
			return false
		}
		i++
	}

	// make finance/stock/ibm/# match finance/stock/ibm
	if i == len(subParts)-1 && subParts[len(subParts)-1] == "#" {
		return true
	}

	if i == len(subParts) {
		return true
	}
	return false
}
