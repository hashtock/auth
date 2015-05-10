package conf

import (
	"fmt"
	"os"
	"sort"
	"time"
)

var (
	// EnvPrefix is optional prefix for environment variable names to prevent clashes
	EnvPrefix string

	// EnvVariableHelp is a map of expected variables and relevant help text
	EnvVariableHelp map[string]string

	// Defaults is a map of defaults for variables
	Defaults map[string]interface{}
)

// MakeKey creates full key based on name of variable and prefix if present
func MakeKey(key string) string {
	return EnvPrefix + key
}

func valueOrDefault(key string) interface{} {
	envkey := MakeKey(key)
	value := os.Getenv(envkey)

	if value != "" {
		return value
	} else if defaultValue, ok := Defaults[key]; ok {
		return defaultValue
	}

	fmt.Printf("Could not get value using environment key: %v\n\n", envkey)
	printConfHelp()
	os.Exit(1)

	return nil
}

// DurationValue gets duration value for given key or default if not present
// Exists if environment variable not set and duration is missing
func DurationValue(key string) time.Duration {
	value := valueOrDefault(key)

	if defaultValue, ok := value.(time.Duration); ok {
		return defaultValue
	}

	duration, err := time.ParseDuration(value.(string))
	if err != nil {
		fmt.Printf("Could not get time duration using environment key: %v. Error: %v\n", MakeKey(key), err)
		os.Exit(1)
	}
	return duration
}

// StringValue gets duration value for given key or default if not present
func StringValue(key string) string {
	value := valueOrDefault(key)
	return value.(string)
}

func printConfHelp() {
	keysOrder := make([]string, 0, len(EnvVariableHelp))
	for key := range EnvVariableHelp {
		keysOrder = append(keysOrder, key)
	}

	sort.StringSlice(keysOrder).Sort()

	fmt.Println("Environmental variables used in configuration")
	for _, key := range keysOrder {
		help := EnvVariableHelp[key]

		envKey := MakeKey(key)
		fmt.Println(envKey)
		value := os.Getenv(envKey)
		if value == "" {
			value = "not set"
			if duration, ok := Defaults[key]; ok {
				value = fmt.Sprintf("%v (default)", duration)
			}
		}
		fmt.Println("\tValue:", value)
		fmt.Println("\tHelp:", help)
		fmt.Println("")
	}
}
