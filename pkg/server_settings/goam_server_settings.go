package server_settings

import (
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log = logger.GetGoamLogger()

// GoamServerSettings holds the configuration for the GoAM server
type GoamServerSettings struct {
	ListenerHttp  string `mapstructure:"listener_http"`
	ListenerHTTPS string `mapstructure:"listener_https"`
	TlsCertFile   string `mapstructure:"tls_cert_file"`
	TlsKeyFile    string `mapstructure:"tls_key_file"`
	DBConnString  string `mapstructure:"db"`

	Banner                   string `mapstructure:"banner"`
	RealmConfigurationFolder string `mapstructure:"realm_configuration_folder"`
	NotFoundRedirectUrl      string `mapstructure:"not_found_redirect_url"`

	UnsafeDisableAdminAuth   bool `mapstructure:"unsafe_disable_admin_auth"`
	EnableRequestTiming      bool `mapstructure:"enable_request_timing"`
	InfrastructureAsCodeMode bool `mapstructure:"infrastructure_as_code_mode"`
	ForwardingProxies        int  `mapstructure:"forwarding_proxies"`

	NodeSettings      map[string]string `mapstructure:"node_settings"`
	ExtensionSettings map[string]string `mapstructure:"extension_settings"`
	BaseUrlOverwrite  map[string]string `mapstructure:"realm_base_url_overwrite"`

	RunDBMigrations bool `mapstructure:"run_db_migrations"`

	// Http server
	ReadBufferSize  int `mapstructure:"read_buffer_size"`
	WriteBufferSize int `mapstructure:"write_buffer_size"`
}

// ConfigDocumentation holds documentation for each configuration option
type ConfigDocumentation struct {
	Field       string
	Description string
	Default     any
	Examples    []string
	EnvVar      string
}

func GetConfigDocumentation() []ConfigDocumentation {
	return []ConfigDocumentation{
		{
			Field:       "listener_http",
			Description: "HTTP listener address and port",
			Default:     ":8080",
			Examples:    []string{":8080", "localhost:8080", "0.0.0.0:8080"},
			EnvVar:      "GOAM_LISTENER_HTTP",
		},
		{
			Field:       "listener_https",
			Description: "HTTPS listener address and port (requires TLS certificate)",
			Default:     "",
			Examples:    []string{":8443", "localhost:8443", "0.0.0.0:8443"},
			EnvVar:      "GOAM_LISTENER_HTTPS",
		},
		{
			Field:       "tls_cert_file",
			Description: "Path to TLS certificate file in PEM format",
			Default:     "",
			Examples:    []string{"/path/to/server.crt", "./certificates/server.pem"},
			EnvVar:      "GOAM_TLS_CERT_FILE",
		},
		{
			Field:       "tls_key_file",
			Description: "Path to TLS private key file in PEM format",
			Default:     "",
			Examples:    []string{"/path/to/server.key", "./certificates/server-key.pem"},
			EnvVar:      "GOAM_TLS_KEY_FILE",
		},
		{
			Field:       "node_settings",
			Description: "Settings for nodes as key-value pairs",
			Default:     []string{},
			Examples:    []string{},
			EnvVar:      "GOAM_NODE_SETTINGS_<KEY>",
		},
		{
			Field:       "banner",
			Description: "String to be displayed in the banner",
			Default:     "GoAM",
			Examples:    []string{},
			EnvVar:      "GOAM_BANNER",
		},
		{
			Field:       "db",
			Description: "Database connection string",
			Default:     "goiam.db?_foreign_keys=on",
			Examples:    []string{"postgres://user:password@localhost:5432/goiamdb", "goiam.db?_foreign_keys=on"},
			EnvVar:      "GOAM_DB",
		},
		{
			Field:       "realm_configuration_folder",
			Description: "Folder containing realm configuration files",
			Default:     "./config",
			Examples:    []string{"./config", "./realms"},
			EnvVar:      "GOAM_REALM_CONFIGURATION_FOLDER",
		},
		{
			Field:       "not_found_redirect_url",
			Description: "URL to redirect to when a not found error occurs. if empty a 404 error will be returned",
			Default:     "",
			Examples:    []string{"https://example.com/", "http://localhost:8080/"},
			EnvVar:      "GOAM_NOT_FOUND_REDIRECT_URL",
		},
		{
			Field:       "unsafe_disable_admin_auth",
			Description: "If true, the admin auth check will be disabled. This can be used for development or if other security methods are used such as a proxy",
			Default:     "false",
			Examples:    []string{"true", "false"},
			EnvVar:      "GOAM_UNSAFE_DISABLE_ADMIN_AUTH",
		},
		{
			Field:       "enable_request_timing",
			Description: "If true, the request timing will be enabled",
			Default:     "false",
			Examples:    []string{"true", "false"},
			EnvVar:      "GOAM_ENABLE_REQUEST_TIMING",
		},
		{
			Field:       "infrastructure_as_code_mode",
			Description: "If true, GoAM overwrites the database content for realms and flows always with the local file configuration",
			Default:     "false",
			Examples:    []string{"true", "false"},
			EnvVar:      "GOAM_INFRASTRUCTURE_AS_CODE_MODE",
		},
		{
			Field:       "forwarding_proxies",
			Description: "The number of forwarding proxies to trust. This is used to trust the X-Forwarded-For header",
			Default:     "0",
			Examples:    []string{"0", "1", "2"},
			EnvVar:      "GOAM_FORWARDING_PROXIES",
		},
		{
			Field:       "extension_settings",
			Description: "Settings for extensions as key-value pairs",
			Default:     []string{},
			Examples:    []string{},
			EnvVar:      "GOAM_EXTENSION_SETTINGS_<KEY>",
		},
		{
			Field:       "run_db_migrations",
			Description: "If true, the database migrations will be run during startup",
			Default:     "false",
			Examples:    []string{"true", "false"},
			EnvVar:      "GOAM_RUN_DB_MIGRATIONS",
		},
		{
			Field:       "realm_base_url_overwrite",
			Description: "If set for a realm, the base url will be overwriten (if not set in the realm config). This is useful for settings different realm urls for different environements like dev and prod",
			Default:     []string{},
			Examples:    []string{"example.com/overwrite/"},
			EnvVar:      "GOAM_REALM_BASE_URL_OVERWRITE_<TENANT/REALM>",
		},
		{
			Field:       "read_buffer_size",
			Description: "The size of the read buffer for the http server. See: https://pkg.go.dev/github.com/valyala/fasthttp?utm_source=godoc#Server.ReadBufferSize",
			Default:     4096,
			Examples:    []string{"4096", "8192", "16384"},
			EnvVar:      "GOAM_READ_BUFFER_SIZE",
		},
		{
			Field:       "write_buffer_size",
			Description: "The size of the write buffer for the http server. See: https://pkg.go.dev/github.com/valyala/fasthttp?utm_source=godoc#Server.WriteBufferSize",
			Default:     4069,
			Examples:    []string{"4096", "8192", "16384"},
			EnvVar:      "GOAM_WRITE_BUFFER_SIZE",
		},
	}
}

func SetConfigFile(cfgFile string) {

	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgFile)
}

func SetDefaultSources() {

	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")          // current working directory
	viper.AddConfigPath("/etc/goam/") // standard linux config directory
	viper.SetConfigName("goam.yaml")
}

func BindCobraFlags(cmd *cobra.Command) error {

	for _, doc := range GetConfigDocumentation() {
		cmd.PersistentFlags().String(doc.Field, doc.GetDefaultAsString(), doc.Description)

		err := viper.BindPFlag(doc.Field, cmd.PersistentFlags().Lookup(doc.Field))
		if err != nil {
			return err
		}
	}

	return nil
}

func InitWithViper() (*GoamServerSettings, error) {

	// We always use env variables with the GOAM prefix that take precedence over the config file
	viper.SetEnvPrefix("GOAM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Viper default values
	for _, doc := range GetConfigDocumentation() {
		viper.SetDefault(doc.Field, doc.Default)
	}

	usedConfigFile := viper.ConfigFileUsed()

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil && usedConfigFile != "" {

		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Load from viper (this will be populated by cobra flags, env vars, config files)
	settings := NewGoamServerSettings()
	err := viper.Unmarshal(&settings)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// For debugging log all environement variables dirrectly from the os
	log.Trace().Interface("settings", settings).Msg("settings")

	return settings, nil
}

func NewGoamServerSettings() *GoamServerSettings {

	return &GoamServerSettings{
		NodeSettings:      make(map[string]string),
		ExtensionSettings: make(map[string]string),
		BaseUrlOverwrite:  make(map[string]string),
	}
}

func (s *ConfigDocumentation) GetDefaultAsString() string {
	return fmt.Sprintf("%v", s.Default)
}
