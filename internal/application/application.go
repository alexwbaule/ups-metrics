package application

import (
	"context"
	"errors"
	"github.com/alexwbaule/ups-metrics/internal/application/config"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	Log    *logger.Logger
	Config *config.Config
}

func NewApplication() *Application {
	log := logger.NewLogger()
	log.Info("Starting application")

	cfg, err := config.NewDefaultConfig()
	if err != nil {
		log.Errorf("error opening config (%s)", err)
		os.Exit(-1)
	}
	log.SetLevel(cfg.GetLogLevel())

	return &Application{
		Log:    log,
		Config: cfg,
	}
}

func (a *Application) Run(f func(ctx context.Context) error) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		return f(ctx)
	})

	go func() {
		<-ctx.Done()
		a.Log.Info("Waiting application shutdown...")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}()

	err := wg.Wait()

	if errors.Is(err, context.Canceled) || err == nil {
		a.Log.Info("Graceful shutdown")
		return
	} else if err != nil {
		a.Log.Fatalf("Graceful shutdown error: %s", err)
	}
	os.Exit(0)
}
