// go-leaderboard
// https://github.com/topfreegames/go-leaderboard
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iris-contrib/middleware/logger"
	"github.com/iris-contrib/middleware/recovery"
	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/topfreegames/go-leaderboard/util"
	"github.com/uber-go/zap"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a go-leaderboard Application
type App struct {
	Debug       bool
	Port        int
	Host        string
	ConfigPath  string
	Errors      metrics.EWMA
	App         *iris.Framework
	Config      *viper.Viper
	Logger      zap.Logger
	RedisClient *util.RedisClient
}

// GetApp returns a new go-leaderboard Application
func GetApp(host string, port int, configPath string, debug bool) *App {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
	}
	app.Configure()
	return app
}

// Configure instantiates the required dependencies for go-leaderboard Application
func (app *App) Configure() {
	app.Logger = zap.NewJSON(zap.WarnLevel)

	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.configureApplication()
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("redis.password", "")
}

func (app *App) loadConfiguration() {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("leaderboard")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", app.ConfigPath))
	}
}

func (app *App) onErrorHandler(err interface{}, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.Object("panicText", err),
		zap.String("stack", string(stack)),
	)
}

func (app *App) configureApplication() {
	c := config.Iris{
		DisableBanner: !app.Debug,
	}

	app.App = iris.New(c)
	a := app.App

	if app.Debug {
		a.Use(logger.New(iris.Logger))
	}
	a.Use(recovery.New(os.Stderr))

	a.Use(&RecoveryMiddleware{OnError: app.onErrorHandler})
	a.Use(&VersionMiddleware{App: app})

	a.Get("/healthcheck", HealthCheckHandler(app))

	app.Errors = metrics.NewEWMA15()
	redisSettings := util.RedisSettings{
		Host:     app.Config.GetString("redis.host"),
		Port:     app.Config.GetInt("redis.port"),
		Password: app.Config.GetString("redis.password"),
	}
	app.RedisClient = util.GetRedisClient(redisSettings)

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()
}

func (app *App) addError() {
	app.Errors.Update(1)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	app.App.Listen(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
