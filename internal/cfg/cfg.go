package cfg

import "github.com/joho/godotenv"

var env map[string]string

func MustRead() {
	if env == nil {
		var err error
		env, err = godotenv.Read()
		if err != nil {
			panic("Error reading .env file: " + err.Error())
		}
	}
}

func Must(key string) string {
	if env == nil {
		MustRead()
	}
	value, ok := env[key]
	if !ok {
		panic("Environment variable not found: " + key)
	}
	return value
}
