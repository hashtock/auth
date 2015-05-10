package conf

import (
	"fmt"
	"net/url"
	"os"

	confTool "github.com/hashtock/service-tools/conf"
)

var (
	keyAppAddress         = "APP_ADDRESS"
	keyServeAddress       = "SERVE_ADDRESS"
	keyDB                 = "DB"
	keyDBName             = "DBName"
	keySessionName        = "SESSION_KEY"
	keySessionSecret      = "SESSION_SECRET"
	keyGoogleClientID     = "GOOGLE_CLIENT_ID"
	keyGoogleClientSecret = "GOOGLE_CLIENT_SECRET"
)

func init() {
	confTool.EnvPrefix = "AUTH_"
	confTool.EnvVariableHelp = map[string]string{
		keyAppAddress:         "External address to application",
		keyServeAddress:       "Host and port for the service",
		keyDB:                 "Location of DB",
		keyDBName:             "Name of DB to use",
		keySessionName:        "Session key name",
		keySessionSecret:      "Session secret used for encrypting",
		keyGoogleClientID:     "Google app client id",
		keyGoogleClientSecret: "Google app shared secret",
	}
	confTool.Defaults = map[string]interface{}{
		keyDB:     "localhost",
		keyDBName: "auth",
	}
}

func loadConfig() {
	if cfg == nil {
		cfg = new(Config)
	}

	cfg.SessionSecret = confTool.StringValue(keySessionSecret)
	cfg.GoogleClientID = confTool.StringValue(keyGoogleClientID)
	cfg.GoogleClientSecret = confTool.StringValue(keyGoogleClientSecret)
	cfg.SessionName = confTool.StringValue(keySessionName)
	cfg.ServeAddress = confTool.StringValue(keyServeAddress)
	cfg.DB = confTool.StringValue(keyDB)
	cfg.DBName = confTool.StringValue(keyDBName)

	appAddress := confTool.StringValue(keyAppAddress)
	appURL, err := url.Parse(appAddress)
	if err != nil {
		fmt.Println("Could not parse app url: %v. Err: %v", appAddress, err)
		os.Exit(1)
	}

	cfg.AppAddress = appURL
}
