package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"nw-guardian/internal"
	"nw-guardian/web"
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

	r.App.Use(logger.New())

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
