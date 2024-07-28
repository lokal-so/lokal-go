package lokal

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
)

type Lokal struct {
	baseURL   string
	basicAuth struct {
		Username string
		Password string
	}
	token string
	rest  *resty.Client
}

func NewDefault() (*Lokal, error) {

	rest := resty.New()
	rest.SetBaseURL("http://127.0.0.1:6174")
	rest.SetHeader("User-Agent", "Lokal Go - github.com/lokal-so/lokal-go")
	rest.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		// Now you have access to Client and Response instance
		// manipulate it as per your need
		version, err := semver.NewVersion(r.Header().Get("Lokal-Server-Version"))
		if err != nil {
			return errors.New("your local client might be outdated, please update")
		}

		minVersion, _ := semver.NewVersion(ServerMinVersion)

		if version.LessThan(minVersion) {
			return fmt.Errorf("your local client is outdated, please update to minimum version %v", ServerMinVersion)
		}

		return nil // if its success otherwise return error
	})

	instance := Lokal{
		baseURL: "http://127.0.0.1:6174",
		rest:    rest,
	}

	return &instance, nil
}

func (l *Lokal) SetBaseURL(url string) *Lokal {
	l.baseURL = url
	l.rest.SetBaseURL(url)
	return l
}

func (l *Lokal) SetBasicAuth(username, password string) *Lokal {
	l.basicAuth = struct {
		Username string
		Password string
	}{username, password}
	l.rest.SetBasicAuth(username, password)
	return l
}

func (l *Lokal) SetAPIToken(token string) *Lokal {
	l.token = token
	l.rest.SetHeader("X-Auth-Token", token)
	return l
}
