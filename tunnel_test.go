package lokal

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewDefault(t *testing.T) {
	instance, err := NewDefault()
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "http://127.0.0.1:6174", instance.baseURL)
	assert.NotNil(t, instance.rest)
	assert.NotEqual(t, resty.New().BaseURL, instance.rest.BaseURL)
}

func TestSetBaseURL(t *testing.T) {
	instance, _ := NewDefault()
	newURL := "http://localhost:8080"
	instance.SetBaseURL(newURL)

	assert.Equal(t, newURL, instance.baseURL)
	assert.Equal(t, newURL, instance.rest.BaseURL)
}

func TestSetBasicAuth(t *testing.T) {
	instance, _ := NewDefault()
	username := "user"
	password := "pass"
	instance.SetBasicAuth(username, password)

	assert.Equal(t, username, instance.basicAuth.Username)
	assert.Equal(t, password, instance.basicAuth.Password)
}

func TestSetAPIToken(t *testing.T) {
	instance, _ := NewDefault()
	token := "mytoken"
	instance.SetAPIToken(token)

	assert.Equal(t, token, instance.token)
	assert.Equal(t, token, instance.rest.Header.Get("X-Auth-Token"))
}
