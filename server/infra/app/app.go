package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"owlet/init/db"
	"owlet/server/infra/assemble"
	"owlet/server/infra/fail"
	"owlet/server/infra/localize"
	"owlet/server/infra/persistence"
	"owlet/server/infra/tracing"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

var (
	GracefulShutdownTimeout time.Duration = 3 * time.Second
	HttpPort                              = 80
)

// Bootstrap
//    database migration (with distribution lock)
//    database connection pool
//    http serve start and become healthy
func Bootstrap() {
	logrus.Infoln("bootstrap...")

	// tracer
	logrus.Infoln("tracing setting...")
	tracer, closer, err := tracing.NewTracer()
	if err != nil {
		logrus.Fatalf("tracing setting: %v\n", err)
	}
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	logrus.Infoln("tracing setting success")

	// database setting up
	if os.Getenv(persistence.EnvDatabaseURL) == "" {
		os.Setenv(persistence.EnvDatabaseURL, "mysql://root:root@(127.0.0.1:3306)/owlet-go?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s")
	}
	logrus.Infoln("database setting...")
	dsn, err := persistence.ParseDatabaseConfigFromEnv()
	if err != nil {
		logrus.Fatalf("database setting: %v\n", err)
	}
	if err := db.PrepareMysqlDatabase(dsn); err != nil {
		logrus.Fatalf("database setting: prepare database: %v\n", err)
	}
	gormDB, err := persistence.StartGormDB(dsn)
	if err != nil {
		logrus.Fatalf("database setting: open db: %v\n", err)
	}
	defer persistence.StopGormDB(gormDB)

	if err := gormDB.AutoMigrate(assemble.AutoMigrations...); err != nil {
		logrus.Fatalf("database setting: auto migration: %v\n", err)
	}

	persistence.ActiveGormDB = gormDB
	logrus.Infoln("database setting success")

	// http server
	engine := gin.New()

	engine.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/"}}),
		localize.LocalizeMiddleware("./i18n"),
		tracing.TracingRestAPI(),
		fail.ErrorHandling(),
		// gin.Recovery(),
	)

	for _, registerEntry := range assemble.RestAPIRegistry {
		registerEntry.Register(engine, registerEntry.MiddleWares...)
	}

	StartHTTPServer(engine)
}

// StartHTTPServer running http server
func StartHTTPServer(engine *gin.Engine) {
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(HttpPort),
		Handler: engine,
	}

	// run http server async
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err) // exit
		}
	}()

	// watch terminate signal
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Infoln("[SHUTDOWN] shutdown signal has been received, the service will exit in 3 seconds.")

	ctx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
	defer cancel()

	// graceful shutdown http.Server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("[SHUTDOWN] http server shutdown:%v\n", err)
	}
	logrus.Infoln("[SHUTDOWN] http server is shutdowning gracefully, new request will be rejected.")

	// waiting for ctx.Done(). timeout of 3 seconds.
	<-ctx.Done()

	logrus.Infoln("[SHUTDOWN] service exiting")
}
