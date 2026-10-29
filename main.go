package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/component/schedule"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/runner"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/compatibility"
	"github.com/chuccp/http2smtp/configapply"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/rest"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/webui"
	"go.uber.org/zap"
)

func createAPP() (*wf.WebFrame, error) {

	var webPort int
	var apiPort int
	var storageRoot string
	flag.IntVar(&webPort, "web_port", 0, "web port")
	flag.IntVar(&apiPort, "api_port", 0, "api port")
	flag.StringVar(&storageRoot, "storage_root", "./", "storage root")
	showVersion := false
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	// Return before loading config.ini or checking the database: asking for the version
	// must not create configuration or put an incompatible installation into page mode.
	if showVersion {
		printVersion()
		return nil, nil
	}
	webPort = util.GetEnvIntOrDefault("web_port", webPort)
	apiPort = util.GetEnvIntOrDefault("api_port", apiPort)
	storageRoot = util.GetEnvOrDefault("storage_root", storageRoot)
	fileConfig, err := config.LoadSingleFileConfig("config.ini")
	if err != nil {
		return nil, err
	}
	// Read through the typed config, not GetBoolOrDefault: an INI file yields strings,
	// which that getter refuses, so core.init/core.dbinit would always read as false here.
	coreState := model.CoreState(fileConfig)
	databaseType := fileConfig.GetString("core.dbtype")
	if strings.TrimSpace(databaseType) == "" {
		databaseType = "sqlite"
	}
	sqlitePath := ""
	if strings.EqualFold(databaseType, "sqlite") {
		sqlitePath = fileConfig.GetString("sqlite.filename")
		if strings.TrimSpace(sqlitePath) == "" {
			sqlitePath = model.DefaultSqliteConfig().Filename
		}
	}
	// This must stay before db.GetDb and AutoCreateTable. Opening an old SQLite file or
	// letting GORM create tables would modify data that this build cannot safely read.
	// An incompatible install still starts the management listener, but only to serve a
	// compatibility page; the database, models, services and public API are not attached.
	var compatibilityIssue *compatibility.IncompatibleError
	if err := compatibility.Check(compatibility.CheckOptions{
		Config:        fileConfig,
		ConfigPath:    "config.ini",
		SQLitePath:    sqlitePath,
		StorageRoot:   storageRoot,
		DatabaseType:  databaseType,
		MySQLDatabase: fileConfig.GetString("mysql.dbname"),
	}); err != nil {
		issue, ok := compatibility.AsIncompatible(err)
		if !ok {
			return nil, err
		}
		compatibilityIssue = issue
	}
	// A port given on the command line or through the environment marks the instance as
	// running in Docker: it is the container's run command (and its port mapping) that
	// fixes the ports, so the UI may not move them — RestartService reads this flag and
	// leaves them where they are.
	dockerPorts := webPort > 0 || apiPort > 0
	if dockerPorts {
		if apiPort == 0 {
			apiPort = webPort
		}
		if webPort == 0 {
			webPort = apiPort
		}
		fileConfig.Put("manage.port", webPort)
		fileConfig.Put("api.port", apiPort)
	}
	// Written on every start, either way: the flag is persisted with the rest of the file,
	// and one left behind by an earlier container run would otherwise keep claiming Docker
	// and lock the ports for good.
	fileConfig.Put("core.isdocker", strconv.FormatBool(dockerPorts))
	if len(storageRoot) > 0 {
		fileConfig.Put("core.cachepath", storageRoot)
	}
	if err := configureFileLogging(fileConfig, storageRoot); err != nil {
		return nil, err
	}
	// Without a flag or environment value the port comes from config.ini — the file the
	// setup wizard and the settings page write — and only then from the built-in default.
	// The typed config is what makes the INI's string values work here; reading the file
	// this way is also what lets a port saved in the UI survive a process restart.
	cfg, err := model.GetConfig(fileConfig)
	if err != nil {
		return nil, err
	}
	if webPort == 0 && cfg.Manage != nil && cfg.Manage.Port > 0 {
		webPort = cfg.Manage.Port
	}
	if apiPort == 0 && cfg.Api != nil && cfg.Api.Port > 0 {
		apiPort = cfg.Api.Port
	}
	builder := wf.NewBuilder(fileConfig)
	if webPort == 0 {
		webPort = model.ManagePort
	}
	// The management UI is server-rendered and embedded in the binary (see webui/), so
	// there is no web/ directory to serve and no SPA fallback to configure. The API
	// keeps its /api prefix in the route registrations instead of through ContextPath,
	// which is a server-wide prefix and would also land on the page routes that must
	// live at the site root (/login, /smtp, ...).
	// One request-body cap for both servers: the framework applies it to JSON, and
	// TokenService applies the same value to form/multipart sends.
	maxBodySize := model.DefaultMaxBodySize
	if cfg.Core != nil {
		maxBodySize = cfg.Core.MaxBodySizeValue()
	}
	if compatibilityIssue != nil {
		log.Warn("data format is incompatible; serving management-only compatibility page",
			zap.String("reason", compatibilityIssue.Reason),
		)
		compatibilityServerConfig := &web.ServerConfig{Port: webPort, MaxBodySize: maxBodySize}
		compatibilityGroup := core.NewRestGroupBuilder().
			Rest(webui.NewCompatibilityPage(compatibilityIssue, Version)).
			Port(webPort).
			ServerConfig(compatibilityServerConfig).
			Build()
		builder.RestGroup(compatibilityGroup)
		return builder.Build(), nil
	}
	restGroupBuilder := core.NewRestGroupBuilder()
	restGroupBuilder.Rest(&rest.Set{}, &rest.User{}, &rest.Token{}, &rest.Mail{}, &rest.Smtp{}, &rest.Schedule{}, &rest.Log{}, &rest.Oplog{}, webui.NewPages(Version))
	restGroupBuilder.Port(webPort)
	// Both server configs are kept: a restart re-runs these same groups, so the restart
	// endpoint rewrites the ports here to make a port saved in the UI take effect.
	manageServerConfig := &web.ServerConfig{Port: webPort, MaxBodySize: maxBodySize}
	restGroupBuilder.ServerConfig(manageServerConfig)
	// No CORS filter: the management UI is same-origin (the pages and the API share
	// this server), and the framework's default filter answers every Origin with
	// credentials allowed — which would expose the whole admin surface to any website
	// a browser can reach, with only the session cookie's SameSite standing in the way.
	restGroupBuilder.Filter(auth2.NewAuthenticationFilter[*model.User](&auth.Authentication{}))
	restGroup := restGroupBuilder.Build()

	builder.RestGroup(restGroup)

	apiRestGroupBuilder := core.NewRestGroupBuilder()
	apiRestGroupBuilder.Rest(&rest.API{})
	if apiPort == 0 {
		apiPort = model.ApiPort
	}
	apiServerConfig := &web.ServerConfig{Port: apiPort, MaxBodySize: maxBodySize}
	apiRestGroupBuilder.Port(apiPort)
	apiRestGroupBuilder.ServerConfig(apiServerConfig)
	builder.RestGroup(apiRestGroupBuilder.Build())

	manageModelGroupBuilder := core.NewModelGroupBuilder()
	manageModelGroupBuilder.AutoCreateTable(true)
	manageModelGroupBuilder.Model(
		&model.MailModel{},
		&model.SMTPModel{},
		&model.TokenModel{},
		&model.ScheduleModel{},
		&model.LogModel{},
		&model.AttachmentModel{},
		&model.UserModel{},
		&model.OperationLogModel{},
	)

	if coreState.Init || coreState.DbInit {
		connection, err := db.GetDb(fileConfig)
		if err != nil {
			return nil, err
		}
		manageModelGroupBuilder.DB(connection)
		// Hand the connection to the switch bookkeeping so a later settings save (or a
		// restart in this same process) closes this pool instead of leaking it.
		configapply.AttachInitial(connection)
	}
	restartService := &service.RestartService{}
	builder.Service(restartService, &service.TokenService{}, &service.ScheduleService{}, &service.LogService{}, &service.SmtpService{}, &service.UserService{}, &service.OperationLogService{}, &service.CaptchaService{})

	builder.Runner(schedule.NewScheduleWithSeconds(), &runner.ScheduleRunner{})

	builder.ModelGroup(manageModelGroupBuilder.Build())

	app := builder.Build()
	// Only the built app can restart itself, so the hook is handed over after Build.
	restartService.Setup(manageServerConfig, apiServerConfig, app.ReStart)
	return app, nil
}

// configureFileLogging keeps application logs with the rest of the persistent data
// when a custom storage root is used. The default working-directory storage keeps the
// existing stdout-only behavior for ordinary local runs.
func configureFileLogging(cfg config.IConfig, storageRoot string) error {
	root := strings.TrimSpace(storageRoot)
	if root == "" || root == "." || root == string(filepath.Separator) {
		return nil
	}
	cleanRoot := filepath.Clean(root)
	if cleanRoot == "." {
		return nil
	}
	logDir := filepath.Join(cleanRoot, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	cfg.Put("web.log.write", true)
	cfg.Put("web.log.path", filepath.Join(logDir, "http2smtp.log"))
	return nil
}

func main() {
	app, err := createAPP()
	if err != nil {
		log.Panic("启动失败", zap.Error(err))
		return
	}
	if app == nil {
		return
	}
	err = app.Start()
	if err != nil {
		log.Panic("启动失败", zap.Error(err))
		return
	}
}
