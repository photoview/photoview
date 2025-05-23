package utils

import (
	"os"
	"strings"
)

// EnvironmentVariable represents the name of an environment variable used to configure Photoview
type EnvironmentVariable string

// General options
const (
	EnvDevelopmentMode           EnvironmentVariable = "PHOTOVIEW_DEVELOPMENT_MODE"
	EnvServeUI                   EnvironmentVariable = "PHOTOVIEW_SERVE_UI"
	EnvUIPath                    EnvironmentVariable = "PHOTOVIEW_UI_PATH"
	EnvMediaCachePath            EnvironmentVariable = "PHOTOVIEW_MEDIA_CACHE"
	EnvFaceRecognitionModelsPath EnvironmentVariable = "PHOTOVIEW_FACE_RECOGNITION_MODELS_PATH"
	EnvHTTPUserAuth              EnvironmentVariable = "PHOTOVIEW_HTTP_USER_AUTH"
)

// Network related
const (
	EnvListenIP    EnvironmentVariable = "PHOTOVIEW_LISTEN_IP"
	EnvListenPort  EnvironmentVariable = "PHOTOVIEW_LISTEN_PORT"
	EnvAPIEndpoint EnvironmentVariable = "PHOTOVIEW_API_ENDPOINT"
	EnvUIEndpoint  EnvironmentVariable = "PHOTOVIEW_UI_ENDPOINT"
)

// Database related
const (
	EnvDatabaseDriver EnvironmentVariable = "PHOTOVIEW_DATABASE_DRIVER"
	EnvMysqlURL       EnvironmentVariable = "PHOTOVIEW_MYSQL_URL"
	EnvPostgresURL    EnvironmentVariable = "PHOTOVIEW_POSTGRES_URL"
	EnvSqlitePath     EnvironmentVariable = "PHOTOVIEW_SQLITE_PATH"
)

// Feature related
const (
	EnvDisableFaceRecognition    EnvironmentVariable = "PHOTOVIEW_DISABLE_FACE_RECOGNITION"
	EnvDisableVideoEncoding      EnvironmentVariable = "PHOTOVIEW_DISABLE_VIDEO_ENCODING"
	EnvDisableRawProcessing      EnvironmentVariable = "PHOTOVIEW_DISABLE_RAW_PROCESSING"
	EnvVideoHardwareAcceleration EnvironmentVariable = "PHOTOVIEW_VIDEO_HARDWARE_ACCELERATION"
)

// GetName returns the name of the environment variable itself
func (v EnvironmentVariable) GetName() string {
	return string(v)
}

// GetValue returns the value of the environment
func (v EnvironmentVariable) GetValue() string {
	return os.Getenv(string(v))
}

// GetBool returns the environment variable as a boolean (defaults to false if not defined)
func (v EnvironmentVariable) GetBool() bool {
	value := strings.ToLower(os.Getenv(string(v)))
	trueValues := []string{"1", "true"}

	for _, x := range trueValues {
		if value == x {
			return true
		}
	}

	return false
}

// ShouldServeUI whether or not the "serve ui" option is enabled
func ShouldServeUI() bool {
	return EnvServeUI.GetBool()
}

// DevelopmentMode describes whether or not the server is running in development mode,
// and should thus print debug informations and enable other features related to developing.
func DevelopmentMode() bool {
	return EnvDevelopmentMode.GetBool()
}

// UIPath returns the value from where the static UI files are located if SERVE_UI=1
func UIPath() string {
	if path := EnvUIPath.GetValue(); path != "" {
		return path
	}

	return "./ui"
}
