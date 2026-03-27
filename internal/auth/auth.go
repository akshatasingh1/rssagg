package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetAPIKey extracts the API key from the request headers and returns it.
//example: Authorization: ApiKey {insert_api_key_here}

func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no authentication info found")
	}
	vals := strings.SplitN(val, " ", 2)

	if len(vals) != 2 {
		return "", errors.New("invalid authentication header format")
	}
	if vals[0] != "ApiKey" {
		return "", errors.New("malformed first part of authentication header")
	}
	return vals[1], nil
}
