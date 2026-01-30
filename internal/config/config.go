package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env  string
	Port string

	// HTTP Server
	ShutdownSecs          int
	ReadHeaderTimeoutSecs int
	ReadTimeoutSecs       int
	WriteTimeoutSecs      int
	IdleTimeoutSecs       int
	MaxHeaderBytes        int

	// CORS
	CORSAllowOrigins     []string
	CORSAllowMethods     []string
	CORSAllowHeaders     []string
	CORSExposeHeaders    []string
	CORSAllowCredentials bool
	CORSMaxAge           int

	// Rate Limiting
	RateLimitEnabled  bool
	RateLimitRPS      float64
	RateLimitBurst    int

	// Security Headers
	SecurityHeadersEnabled bool
	XFrameOptions          string
	ContentTypeNosniff     bool
	XSSProtection          bool

	// Compression
	CompressionEnabled bool
	CompressionLevel   int

	// Request Limits
	MaxBodySize    int64
	RequestTimeout time.Duration

	// Trusted Proxies
	TrustedProxies []string

	// MongoDB
	MongoURI      string
	MongoDatabase string
	MongoTimeout  time.Duration

	// JWT
	JWTSecret            string
	JWTExpiration        time.Duration
	JWTRefreshExpiration time.Duration

	// Payment Providers
	PaymentDefaultProvider string
	WompiPublicKey         string
	WompiPrivateKey        string
	WompiWebhookSecret     string
	StripeAPIKey           string
	StripeWebhookSecret    string
}

func Load() *Config {
	return &Config{
		Env:  getEnv("APP_ENV", "development"),
		Port: getEnv("PORT", "8080"),

		ShutdownSecs:          getEnvInt("SHUTDOWN_SECS", 10),
		ReadHeaderTimeoutSecs: getEnvInt("READ_HEADER_TIMEOUT_SECS", 5),
		ReadTimeoutSecs:       getEnvInt("READ_TIMEOUT_SECS", 15),
		WriteTimeoutSecs:      getEnvInt("WRITE_TIMEOUT_SECS", 15),
		IdleTimeoutSecs:       getEnvInt("IDLE_TIMEOUT_SECS", 60),
		MaxHeaderBytes:        getEnvInt("MAX_HEADER_BYTES", 1<<20),

		// CORS
		CORSAllowOrigins:     getEnvSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
		CORSAllowMethods:     getEnvSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		CORSAllowHeaders:     getEnvSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}),
		CORSExposeHeaders:    getEnvSlice("CORS_EXPOSE_HEADERS", []string{"X-Request-ID"}),
		CORSAllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
		CORSMaxAge:           getEnvInt("CORS_MAX_AGE", 86400),

		// Rate Limiting
		RateLimitEnabled: getEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitRPS:     getEnvFloat("RATE_LIMIT_RPS", 100),
		RateLimitBurst:   getEnvInt("RATE_LIMIT_BURST", 200),

		// Security
		SecurityHeadersEnabled: getEnvBool("SECURITY_HEADERS_ENABLED", true),
		XFrameOptions:          getEnv("X_FRAME_OPTIONS", "DENY"),
		ContentTypeNosniff:     getEnvBool("CONTENT_TYPE_NOSNIFF", true),
		XSSProtection:          getEnvBool("XSS_PROTECTION", true),

		// Compression
		CompressionEnabled: getEnvBool("COMPRESSION_ENABLED", true),
		CompressionLevel:   getEnvInt("COMPRESSION_LEVEL", 5),

		// Limits
		MaxBodySize:    getEnvInt64("MAX_BODY_SIZE", 10<<20), // 10MB
		RequestTimeout: time.Duration(getEnvInt("REQUEST_TIMEOUT_SECS", 30)) * time.Second,

		// Proxies
		TrustedProxies: getEnvSlice("TRUSTED_PROXIES", nil),

		// MongoDB
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", ""),
		MongoTimeout:  time.Duration(getEnvInt("MONGO_TIMEOUT_SECS", 10)) * time.Second,

		// JWT
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTExpiration:        time.Duration(getEnvInt("JWT_EXPIRATION_MINS", 15)) * time.Minute,
		JWTRefreshExpiration: time.Duration(getEnvInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,

		// Payment Providers
		PaymentDefaultProvider: getEnv("PAYMENT_DEFAULT_PROVIDER", "wompi"),
		WompiPublicKey:         getEnv("WOMPI_PUBLIC_KEY", ""),
		WompiPrivateKey:        getEnv("WOMPI_PRIVATE_KEY", ""),
		WompiWebhookSecret:     getEnv("WOMPI_WEBHOOK_SECRET", ""),
		StripeAPIKey:           getEnv("STRIPE_API_KEY", ""),
		StripeWebhookSecret:    getEnv("STRIPE_WEBHOOK_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func getEnvFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getEnvSlice(key string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return strings.Split(v, ",")
	}
	return def
}
