package ha

import (
	"errors"
	"strings"
)

func MakeProvider(provider string) (HighAvailability, error) {
	provider = strings.ToLower(provider)
	switch provider {
	case "maxscale":
		return MaxScale{}, nil
	case "patroni":
		return Patroni{}, nil
	default:
		return nil, errors.New("Unknown HA provider: " + provider)
	}
}
