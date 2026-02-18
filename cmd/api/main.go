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

// @title           Vetsify API
// @version         1.0
// @description     API para plataforma veterinaria multi-tenant. Soporta panel de administración (staff) y app móvil (owners).
// @description
// @description     ## Autenticación
// @description     Todas las rutas protegidas requieren el header `Authorization: Bearer <token>`.
// @description     - **Staff**: Obtén tokens via `POST /api/auth/login`
// @description     - **Owners (mobile)**: Obtén tokens via `POST /mobile/auth/login`
// @description
// @description     ## Multi-tenancy
// @description     Las rutas que operan sobre datos de un tenant específico (patients, species, etc.) requieren el header `X-Tenant-ID` con el ObjectID del tenant.
// @description     Este header debe enviarse en **todas** las peticiones a rutas marcadas con el parámetro `X-Tenant-ID`.
// @description     El frontend puede configurarlo una vez en el interceptor HTTP (axios/fetch).
// @description
// @description     ## Roles (RBAC)
// @description     Las rutas del panel admin (`/api/*`) están protegidas por RBAC.
// @description     Roles disponibles: `admin`, `veterinarian`, `receptionist`, `assistant`, `accountant`.
// @description
// @description     ## Paginación
// @description     Las rutas que retornan listas soportan `skip` y `limit` como query params.
// @description     Respuesta incluye `pagination: { skip, limit, total, total_pages }`.

// @contact.name   Vetsify Support
// @contact.email  support@vetsify.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

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
			cfg.WompiRedirectURL,
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
