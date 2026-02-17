package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/eren_dev/go_server/internal/app"
	"github.com/eren_dev/go_server/internal/app/lifecycle"
	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/health"
	"github.com/eren_dev/go_server/internal/platform/logger"
	"github.com/eren_dev/go_server/internal/platform/notifications/fcm"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/platform/payment/wompi"
	"github.com/eren_dev/go_server/internal/shared/database"
)

// @title           Go Server API
// @version         1.0
// @description     Template API con Go + Gin + MongoDBD
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Ingresa el token con el prefijo Bearer: Bearer <token>

func main() {
	_ = godotenv.Load(".env")

	cfg := config.Load()

	log := logger.NewSlogLogger(cfg.Env)
	logger.SetDefault(log)

	if err := cfg.Validate(); err != nil {
		logger.Default().Error(context.Background(), "invalid_configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.NewProvider(cfg)
	if err != nil {
		logger.Default().Error(context.Background(), "database_connection_failed", "error", err)
		os.Exit(1)
	}

	if db != nil {
		logger.Default().Info(context.Background(), "database_connected", "database", cfg.MongoDatabase)
		health.SetDatabase(db)
	} else {
		logger.Default().Info(context.Background(), "database_disabled", "reason", "MONGO_DATABASE not configured")
	}

	// Initialize Payment Manager
	var defaultProvider payment.ProviderType
	switch cfg.PaymentDefaultProvider {
	case "wompi":
		defaultProvider = payment.ProviderWompi
	case "stripe":
		defaultProvider = payment.ProviderStripe
	default:
		defaultProvider = payment.ProviderWompi
	}

	paymentManager := payment.NewPaymentManager(defaultProvider)

	// Register Wompi provider if configured
	if cfg.WompiPublicKey != "" && cfg.WompiPrivateKey != "" {
		wompiProvider := wompi.NewWompiProvider(
			cfg.WompiPublicKey,
			cfg.WompiPrivateKey,
			cfg.WompiWebhookSecret,
		)
		if err := paymentManager.RegisterProvider(wompiProvider); err != nil {
			logger.Default().Error(context.Background(), "failed_to_register_wompi", "error", err)
		} else {
			logger.Default().Info(context.Background(), "payment_provider_registered", "provider", "wompi")
		}
	}

	// TODO: Register Stripe provider when implemented
	// if cfg.StripeAPIKey != "" {
	//     stripeProvider := stripe.NewStripeProvider(cfg.StripeAPIKey, cfg.StripeWebhookSecret)
	//     paymentManager.RegisterProvider(stripeProvider)
	// }

	// Initialize FCM push provider
	pushProvider, err := fcm.NewProvider(ctx, cfg)
	if err != nil {
		logger.Default().Error(context.Background(), "fcm_init_failed", "error", err)
		os.Exit(1)
	}

	workers := lifecycle.NewWorkers()

	server, err := app.NewServer(cfg, db, paymentManager, pushProvider)
	if err != nil {
		logger.Default().Error(context.Background(), "server_error", "error", err)
		os.Exit(1)
	}

	logger.Default().Info(context.Background(), "server_running", "port", cfg.Port, "env", cfg.Env)

	go func() {
		if err := server.Start(); err != nil {
			logger.Default().Error(context.Background(), "server_error", "error", err)
		}
	}()

	health.SetReady(true)

	<-ctx.Done()

	if db != nil {
		if err := db.Close(context.Background()); err != nil {
			logger.Default().Error(context.Background(), "database_close_error", "error", err)
		}
	}

	shutdowner := lifecycle.NewShutdowner(server, workers, 10*time.Second)
	shutdowner.Shutdown(context.Background())
}
