package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"nw-guardian/internal"
	"nw-guardian/internal/agent"
	"nw-guardian/web"
	"nw-guardian/workers"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	var err error

	runtime.GOMAXPROCS(4)

	log.SetFormatter(&log.TextFormatter{})

	// Load .env
	err = godotenv.Load()
	if err != nil {
		log.Error(err)
	}

	// connect to database
	database := internal.DatabaseConnection{
		URI:    os.Getenv("MONGO_URI"),
		DB:     os.Getenv("MAIN_DB"),
		Logger: log.New(),
	}

	database.Connect()

	handleSignals()

	// TODO load routes for main API (primarily front end, & agent auth?)
	r := web.NewRouter(database.MongoDB)
	r.ProbeDataChan = make(chan agent.ProbeData)
	workers.CreateProbeDataWorker(r.ProbeDataChan, r.DB)

	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods",
				"POST, PUT, PATCH, DELETE, OPTIONS")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type,*")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}
	r.App.UseRouter(crs)

	// fully load and apply routes
	r.Init()
	r.Listen(os.Getenv("LISTEN"))
}

func handleSignals() {
	// Signal Termination if using CLI
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGKILL)
	go func() {
		for range signals {
			shutdown()
		}
	}()
}

func shutdown() {
	fmt.Println()
	log.Warnf("%d threads at exit.", runtime.NumGoroutine())
	log.Warn("Shutting down NetWatcher Agent...")
	os.Exit(1)
}
