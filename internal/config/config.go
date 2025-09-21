package config

import (
	"strings"

	"github.com/Identityplane/GoAM/internal/db/postgres_adapter"
	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/server_settings"
)

var PostgresUserDB *postgres_adapter.PostgresUserDB
var SqliteUserDB *sqlite_adapter.SQLiteUserDB

var (
	ServerSettings *server_settings.GoamServerSettings
)

func GetDbDriverName() string {

	if strings.HasPrefix(ServerSettings.DBConnString, "postgres://") {
		return "postgres"
	}
	return "sqlite"
}

func InitConfiguration(settings *server_settings.GoamServerSettings) {

	log := logger.GetGoamLogger()

	ServerSettings = settings

	log.Debug().
		Str("realm_configuration_folder", settings.RealmConfigurationFolder).
		Bool("infrastructure_as_code_mode", settings.InfrastructureAsCodeMode).
		Bool("unsafe_disable_admin_authz_check", settings.UnsafeDisableAdminAuth).
		Int("number_of_proxies", settings.ForwardingProxies).
		Bool("enable_request_timing", settings.EnableRequestTiming).
		Str("not_found_redirect_url", settings.NotFoundRedirectUrl).
		Msg("loaded server configuration")

	if settings.UnsafeDisableAdminAuth {
		log.Warn().Msg("admin auth is disabled")
	}
}
