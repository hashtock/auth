package webapp_test

import (
	"testing"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/gplus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterProviders(t *testing.T) {
	goth.ClearProviders()
	makeHandler()

	expectedProviders := []string{
		"gplus",
	}

	providers := goth.GetProviders()
	assert.Len(t, providers, len(expectedProviders))
}

func TestRegisterGoogleProviderWithCorrectCallback(t *testing.T) {
	goth.ClearProviders()
	makeHandler()

	provider, err := goth.GetProvider("gplus")
	assert.NoError(t, err)

	gplus, ok := provider.(*gplus.Provider)
	assert.True(t, ok)
	assert.Equal(t, "http://localhost:1234/login/gplus/callback", gplus.CallbackURL)
}
