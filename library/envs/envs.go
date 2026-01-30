package envs

import (
	"os"
)

const (
	PrdType   = "prd"
	StageType = "stg"
	DevType   = "dev"
)

var (
	ServerName  = getEnv("SERVER_NAME", "aitrics-vital-signs")
	ServiceType = getEnv("SERVICE_TYPE", "dev") // prd / stg / dev
	ServerPort  = getEnv("SERVER_PORT", "8080")

	LogLevel = getEnv("LOG_LEVEL", "debug") // debug | info | warn | error |

	DBHost     = getEnv("DB_HOST", "")
	DBPort     = getEnv("DB_PORT", "")
	DBName     = getEnv("DB_NAME", "")
	DBUser     = getEnv("DB_USER", "")
	DBPassword = getEnv("DB_PASSWORD", "")

	Token = getEnv("TOKEN", "")
)

func getEnv(envName, defaultVal string) string {
	envVal := os.Getenv(envName)
	if envVal == "" {
		envVal = defaultVal
	}
	return envVal
}
