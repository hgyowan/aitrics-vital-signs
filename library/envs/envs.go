package envs

import (
	"os"
	"strconv"
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

	VitalRiskTimeWindowHours = getEnvAsInt("VITAL_RISK_TIME_WINDOW_HOURS", 24)
)

func getEnv(envName, defaultVal string) string {
	envVal := os.Getenv(envName)
	if envVal == "" {
		envVal = defaultVal
	}
	return envVal
}

func getEnvAsInt(envName string, defaultVal int) int {
	envVal := os.Getenv(envName)
	if envVal == "" {
		return defaultVal
	}
	if intVal, err := strconv.Atoi(envVal); err == nil && intVal > 0 {
		return intVal
	}
	return defaultVal
}
