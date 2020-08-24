package util

import "os"

func Getenv(name, defaultValue string) string {
	output := defaultValue
	if env, ok := os.LookupEnv(name); ok {
		output = env
	}
	return output
}
