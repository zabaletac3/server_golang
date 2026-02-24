# Plan de Corrección - Vetsify

> **Objetivo**: Corregir los problemas críticos y mejoras identificadas en el análisis del código base.
> 
> **Enfoque**: Implementación por fases para minimizar riesgos y permitir validación incremental.

---

## 📋 Resumen de Fases

| Fase | Descripción | Impacto | Duración Estimada |
|------|-------------|---------|-------------------|
| **Fase 1** | Correcciones Críticas | Alto | 2-3 días |
| **Fase 2** | Mejoras de Seguridad | Medio-Alto | 1-2 días |
| **Fase 3** | Resiliencia y Performance | Medio | 2-3 días |
| **Fase 4** | Testing y Documentación | Bajo-Medio | 2-3 días |
| **Fase 5** | Limpieza y Optimización | Bajo | 1-2 días |

---

## 🔴 Fase 1: Correcciones Críticas

### 1.1 Configurar Reglas de Negocio

**Problema**: Horarios de negocio, período de trial y otros valores hardcodeados.

**Archivos a modificar**:
- `internal/config/config.go` - Agregar nuevos campos
- `.env.example` - Agregar nuevas variables
- `internal/modules/appointments/service.go` - Usar configuración
- `internal/modules/tenant/service.go` - Usar configuración

**Pasos**:
1. Agregar al struct `Config`:
   ```go
   // Business Rules
   AppointmentBusinessStartHour int `env:"APPOINTMENT_START_HOUR" envDefault:"8"`
   AppointmentBusinessEndHour   int `env:"APPOINTMENT_END_HOUR" envDefault:"18"`
   TenantTrialDays              int `env:"TENANT_TRIAL_DAYS" envDefault:"14"`
   SchedulerIntervalMinutes     int `env:"SCHEDULER_INTERVAL_MINS" envDefault:"15"`
   ```

2. Actualizar `.env.example`:
   ```env
   # Business Rules
   APPOINTMENT_START_HOUR=8
   APPOINTMENT_END_HOUR=18
   TENANT_TRIAL_DAYS=14
   SCHEDULER_INTERVAL_MINS=15
   ```

3. Modificar `appointments/service.go`:
   ```go
   // Reemplazar hardcoded 8-18
   hour := scheduledAt.Hour()
   if hour < s.cfg.AppointmentBusinessStartHour || hour >= s.cfg.AppointmentBusinessEndHour {
       return ErrInvalidAppointmentTime
   }
   ```

4. Modificar `tenant/service.go`:
   ```go
   // Reemplazar hardcoded 14 días
   trialEnd := time.Now().AddDate(0, 0, s.cfg.TenantTrialDays)
   ```

**Validación**:
- [ ] Tests unitarios pasan
- [ ] Variables de ambiente funcionan correctamente
- [ ] Valores por defecto se aplican cuando no están configurados

---

### 1.2 Agregar Índices de MongoDB

**Problema**: Consultas críticas sin índices garantizados.

**Archivos a modificar**:
- `internal/modules/tenant/repository.go`
- `internal/modules/appointments/repository.go`
- `cmd/seed/main.go` - Script de migración de índices

**Pasos**:
1. Crear archivo `internal/modules/tenant/indexes.go`:
   ```go
   func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
       indexes := []mongo.IndexModel{
           {
               Keys:    bson.D{{"subscription.external_subscription_id", 1}},
               Options: options.Index().SetUnique(true).SetSparse(true),
           },
           {
               Keys:    bson.D{{"domain", 1}},
               Options: options.Index().SetUnique(true),
           },
       }
       // Crear índices...
   }
   ```

2. Crear archivo `internal/modules/appointments/indexes.go`:
   ```go
   func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
       indexes := []mongo.IndexModel{
           {
               Keys: bson.D{
                   {"tenant_id", 1},
                   {"status", 1},
                   {"scheduled_at", 1},
               },
           },
           {
               Keys: bson.D{
                   {"veterinarian_id", 1},
                   {"scheduled_at", 1},
               },
               Options: options.Index().SetUnique(true),
           },
       }
       // Crear índices...
   }
   ```

3. Actualizar `cmd/api/main.go` para ejecutar `EnsureIndexes` al inicio

**Validación**:
- [ ] Índices se crean al iniciar la aplicación
- [ ] Queries usan los índices (explicar con `explain()`)
- [ ] No hay duplicados en campos únicos

---

### 1.3 Fixear Goroutine de Notificaciones

**Problema**: Goroutines sin timeout pueden causar memory leaks.

**Archivos a modificar**:
- `internal/modules/notifications/service.go`
- `internal/modules/notifications/handler.go`

**Pasos**:
1. Modificar `service.go`:
   ```go
   func (s *Service) sendPushAsync(ctx context.Context, notif *Notification) {
       ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
       defer cancel()
       
       go func() {
           defer cancel()
           
           // Propagar contexto al provider
           if err := s.fcmProvider.Send(ctx, notif); err != nil {
               s.logger.Error("Failed to send push", slog.String("error", err.Error()))
           }
       }()
   }
   ```

2. Actualizar todas las llamadas a `sendPushAsync` para pasar contexto

**Validación**:
- [ ] Goroutines terminan dentro del timeout
- [ ] No hay leaks en pruebas de carga
- [ ] Errores se loguean correctamente

---

### 1.4 Migración de Índices (Script Standalone)

**Problema**: Índices faltantes en producción.

**Archivos a crear**:
- `cmd/migrate-indexes/main.go`

**Pasos**:
1. Crear script de migración independiente:
   ```go
   package main
   
   func main() {
       // Conectar a MongoDB
       // Ejecutar EnsureIndexes de todos los módulos
       // Reportar resultados
   }
   ```

**Validación**:
- [ ] Script corre sin errores
- [ ] Todos los índices se crean
- [ ] Script es idempotente

---

## 🟠 Fase 2: Mejoras de Seguridad

### 2.1 Validación de Firma de Webhook

**Problema**: Firma vacía podría permitir bypass.

**Archivos a modificar**:
- `internal/modules/webhooks/handler.go`
- `internal/platform/webhook/validator.go` (nuevo)

**Pasos**:
1. Crear `internal/platform/webhook/validator.go`:
   ```go
   type SignatureValidator struct {
       secrets map[string]string
   }
   
   func (v *SignatureValidator) Validate(provider, signature, payload string) error {
       if signature == "" {
           return errors.New("missing signature header")
       }
       // Validar firma...
   }
   ```

2. Modificar `webhooks/handler.go`:
   ```go
   signature := c.GetHeader("X-Signature")
   if signature == "" {
       signature = c.GetHeader("X-Wompi-Signature")
   }
   
   if signature == "" {
       c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
       return
   }
   ```

**Validación**:
- [ ] Requests sin firma son rechazados
- [ ] Firmas inválidas son rechazadas
- [ ] Tests de seguridad pasan

---

### 2.2 Request ID en Respuestas de Error

**Problema**: Request ID generado pero no incluido en respuestas.

**Archivos a modificar**:
- `internal/shared/httpx/response.go` (nuevo)
- `internal/shared/middleware/request_id.go`
- Todos los handlers que retornan errores

**Pasos**:
1. Crear `internal/shared/httpx/response.go`:
   ```go
   type ErrorResponse struct {
       RequestID string `json:"request_id"`
       Code      string `json:"code"`
       Message   string `json:"message"`
       Details   any    `json:"details,omitempty"`
   }
   
   func Error(c *gin.Context, status int, code, message string) {
       requestID, _ := c.Get("request_id")
       c.JSON(status, ErrorResponse{
           RequestID: requestID.(string),
           Code:      code,
           Message:   message,
       })
   }
   ```

2. Actualizar middleware `request_id.go` para almacenar en contexto

3. Reemplazar todos los `c.JSON(http.StatusBadRequest, gin.H{"error": ...})` con `httpx.Error()`

**Validación**:
- [ ] Todas las respuestas de error incluyen request_id
- [ ] Request ID es consistente en logs y respuestas

---

### 2.3 Validación Consistente de Tenant

**Problema**: Algunas rutas `/api/*` no validan tenant middleware.

**Archivos a modificar**:
- `internal/app/router.go`
- Revisar todos los módulos

**Pasos**:
1. Revisar cada módulo y verificar que rutas `/api/*` usen:
   ```go
   apiGroup.Use(middleware.TenantMiddleware())
   apiGroup.Use(middleware.JWTAuth(cfg))
   apiGroup.Use(middleware.RBAC())
   ```

2. Crear test que verifique todas las rutas protegidas

**Validación**:
- [ ] Todas las rutas `/api/*` requieren tenant válido
- [ ] Tests de integración pasan

---

## 🟡 Fase 3: Resiliencia y Performance

### 3.1 Circuit Breaker para Servicios Externos

**Problema**: Sin retry/circuit breaker para Wompi y FCM.

**Archivos a modificar**:
- `internal/platform/payment/wompi.go`
- `internal/platform/fcm/provider.go`
- `go.mod` - Agregar `github.com/sony/gobreaker`

**Pasos**:
1. Agregar dependencia:
   ```bash
   go get github.com/sony/gobreaker
   ```

2. Crear wrapper `internal/platform/circuitbreaker/breaker.go`:
   ```go
   package circuitbreaker
   
   import "github.com/sony/gobreaker"
   
   var (
       PaymentBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
           Name:        "payment",
           MaxRequests: 1,
           Interval:    30 * time.Second,
           Timeout:     60 * time.Second,
           ReadyToTrip: func(counts gobreaker.Counts) bool {
               return counts.ConsecutiveFailures >= 5
           },
       })
       
       FCMBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
           Name:        "fcm",
           MaxRequests: 1,
           Interval:    10 * time.Second,
           Timeout:     30 * time.Second,
           ReadyToTrip: func(counts gobreaker.Counts) bool {
               return counts.ConsecutiveFailures >= 3
           },
       })
   )
   ```

3. Envolver llamadas externas:
   ```go
   result, err := circuitbreaker.PaymentBreaker.Execute(func() (interface{}, error) {
       return p.createPaymentLink(ctx, params)
   })
   ```

**Validación**:
- [ ] Circuit breaker se abre tras fallos consecutivos
- [ ] Se cierra automáticamente tras timeout
- [ ] Tests de resiliencia pasan

---

### 3.2 Retry con Exponential Backoff

**Problema**: Fallos transitorios no se reintentan.

**Archivos a modificar**:
- `internal/platform/retry/retry.go` (nuevo)
- `internal/platform/payment/wompi.go`
- `internal/platform/fcm/provider.go`

**Pasos**:
1. Crear `internal/platform/retry/retry.go`:
   ```go
   package retry
   
   func WithExponentialBackoff(ctx context.Context, maxRetries int, fn func() error) error {
       var lastErr error
       baseDelay := 100 * time.Millisecond
       
       for i := 0; i < maxRetries; i++ {
           if err := fn(); err == nil {
               return nil
           } else {
               lastErr = err
           }
           
           select {
           case <-time.After(baseDelay * time.Duration(1 << i)):
           case <-ctx.Done():
               return ctx.Err()
           }
       }
       return lastErr
   }
   ```

2. Integrar con circuit breaker

**Validación**:
- [ ] Retries ocurren con delays exponenciales
- [ ] Context cancellation respeta retries

---

### 3.3 Caché Redis para RBAC

**Problema**: 4 queries por cada request autorizado.

**Archivos a modificar**:
- `internal/shared/middleware/rbac.go`
- `internal/platform/cache/redis.go` (nuevo)
- `go.mod` - Agregar `github.com/redis/go-redis/v9`

**Pasos**:
1. Agregar Redis como dependencia opcional

2. Crear `internal/platform/cache/cache.go`:
   ```go
   type Cache interface {
       Get(ctx context.Context, key string) (string, error)
       Set(ctx context.Context, key string, value string, ttl time.Duration) error
       Delete(ctx context.Context, key string) error
   }
   ```

3. Implementar Redis cache

4. Modificar RBAC middleware para cachear:
   - User → Roles
   - Roles → Permissions
   - Permissions → Resources

**Validación**:
- [ ] Cache hit reduce queries a MongoDB
- [ ] Cache invalidation funciona al actualizar roles
- [ ] Fallback a MongoDB si Redis falla

---

## 🟢 Fase 4: Testing y Documentación

### 4.1 Tests de Integración con Testcontainers

**Problema**: Solo tests unitarios con mocks.

**Archivos a crear**:
- `internal/modules/appointments/integration_test.go`
- `internal/modules/tenant/integration_test.go`
- `testcontainers/helper.go`

**Pasos**:
1. Agregar dependencias:
   ```bash
   go get github.com/testcontainers/testcontainers-go
   go get github.com/testcontainers/testcontainers-go/modules/mongodb
   ```

2. Crear helper `testcontainers/helper.go`:
   ```go
   func SetupTestContainer(ctx context.Context) (*mongo.Client, func(), error) {
       // Levantar MongoDB en container
       // Retornar client y cleanup function
   }
   ```

3. Crear tests de integración para módulos críticos

**Validación**:
- [ ] Tests de integración pasan en CI
- [ ] Cobertura de integración > 60%

---

### 4.2 Audit Logging Extendido

**Problema**: Solo appointments tiene audit log de transiciones.

**Archivos a modificar**:
- `internal/modules/tenant/audit.go` (nuevo)
- `internal/modules/users/audit.go` (nuevo)
- `internal/modules/payments/audit.go` (nuevo)

**Pasos**:
1. Crear colección `audit_logs`

2. Crear servicio de audit logging:
   ```go
   type AuditService interface {
       Log(ctx context.Context, event AuditEvent) error
       Query(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
   }
   ```

3. Instrumentar operaciones críticas:
   - Crear/actualizar tenant
   - Cambios de estado de pago
   - Cambios de roles/permisos
   - Eliminaciones (soft delete)

**Validación**:
- [ ] Eventos críticos se auditan
- [ ] Query de audit logs funciona
- [ ] Performance no degradada

---

### 4.3 Rate Limiting por Tenant

**Problema**: Rate limiting solo por IP, no por tenant.

**Archivos a modificar**:
- `internal/shared/middleware/rate_limit.go`
- `internal/platform/ratelimit/limiter.go` (nuevo)

**Pasos**:
1. Crear limiter jerárquico:
   ```go
   type HierarchicalLimiter struct {
       globalLimiter ratelimit.Limiter
       tenantLimiters sync.Map // map[string]*ratelimit.Limiter
   }
   ```

2. Modificar middleware para extraer tenant_id y aplicar limiter correspondiente

**Validación**:
- [ ] Tenants no pueden exceder límite individual
- [ ] Límite global también se respeta

---

## 🔵 Fase 5: Limpieza y Optimización

### 5.1 Estandarizar Formato de Errores

**Problema**: Respuestas de error inconsistentes.

**Archivos a modificar**:
- Todos los handlers

**Pasos**:
1. Definir formato estándar en `internal/shared/httpx/errors.go`:
   ```go
   type APIError struct {
       RequestID string            `json:"request_id"`
       Timestamp time.Time         `json:"timestamp"`
       Code      string            `json:"code"`
       Message   string            `json:"message"`
       Path      string            `json:"path"`
       Details   map[string]string `json:"details,omitempty"`
   }
   ```

2. Crear códigos de error estandarizados:
   ```go
   const (
       ErrValidation      = "VALIDATION_ERROR"
       ErrNotFound        = "NOT_FOUND"
       ErrUnauthorized    = "UNAUTHORIZED"
       ErrForbidden       = "FORBIDDEN"
       ErrConflict        = "CONFLICT"
       ErrInternal        = "INTERNAL_ERROR"
       ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
   )
   ```

3. Refactorizar todos los handlers

**Validación**:
- [ ] Todas las respuestas siguen formato estándar
- [ ] Tests de handlers actualizados

---

### 5.2 Health Check Extendido

**Problema**: Health check solo verifica MongoDB.

**Archivos a modificar**:
- `internal/modules/health/handler.go`
- `internal/modules/health/service.go`

**Pasos**:
1. Agregar checks para:
   - MongoDB (ping + query simple)
   - Redis (si está configurado)
   - FCM (verificar credenciales)
   - Payment provider (verificar conexión)

2. Retornar estado detallado:
   ```json
   {
     "status": "degraded",
     "checks": {
       "mongodb": {"status": "healthy", "latency_ms": 5},
       "redis": {"status": "healthy", "latency_ms": 2},
       "fcm": {"status": "unhealthy", "error": "invalid credentials"},
       "payment": {"status": "healthy", "provider": "wompi"}
     }
   }
   ```

**Validación**:
- [ ] Health check reporta estado real
- [ ] Kubernetes/readiness probes funcionan

---

### 5.3 Completar Integración Stripe

**Problema**: Stripe provider referenciado pero no implementado.

**Archivos a modificar/crear**:
- `internal/platform/payment/stripe.go` (nuevo)
- `internal/config/config.go`
- `.env.example`

**Pasos**:
1. Agregar configuración Stripe:
   ```go
   StripeSecretKey   string `env:"STRIPE_SECRET_KEY"`
   StripePublishableKey string `env:"STRIPE_PUBLISHABLE_KEY"`
   StripeWebhookSecret  string `env:"STRIPE_WEBHOOK_SECRET"`
   ```

2. Implementar `StripeProvider` siguiendo interfaz `PaymentProvider`

3. Agregar webhook handler para Stripe

**Validación**:
- [ ] Stripe como provider alternativo
- [ ] Webhooks de Stripe procesados
- [ ] Tests de Stripe pasan

---

### 5.4 Métricas Prometheus

**Problema**: Sin métricas de negocio.

**Archivos a crear**:
- `internal/platform/metrics/prometheus.go`
- `internal/shared/middleware/metrics.go`

**Pasos**:
1. Agregar dependencias:
   ```bash
   go get github.com/prometheus/client_golang/prometheus
   ```

2. Crear métricas:
   - `appointments_total` (counter, por estado)
   - `appointments_duration_seconds` (histogram)
   - `tenants_total` (gauge)
   - `payments_total` (counter, por provider)
   - `payments_amount_total` (counter, por provider)
   - `rbac_checks_total` (counter)
   - `rbac_checks_denied_total` (counter)

3. Exponer endpoint `/metrics`

**Validación**:
- [ ] Métricas visibles en `/metrics`
- [ ] Grafana dashboards funcionan

---

## 📊 Checklist de Validación por Fase

### Fase 1 ✅
- [ ] 1.1 Configurar Reglas de Negocio
- [ ] 1.2 Agregar Índices de MongoDB
- [ ] 1.3 Fixear Goroutine de Notificaciones
- [ ] 1.4 Migración de Índices

### Fase 2 ✅
- [ ] 2.1 Validación de Firma de Webhook
- [ ] 2.2 Request ID en Respuestas de Error
- [ ] 2.3 Validación Consistente de Tenant

### Fase 3 ✅
- [ ] 3.1 Circuit Breaker para Servicios Externos
- [ ] 3.2 Retry con Exponential Backoff
- [ ] 3.3 Caché Redis para RBAC

### Fase 4 ✅
- [ ] 4.1 Tests de Integración con Testcontainers
- [ ] 4.2 Audit Logging Extendido
- [ ] 4.3 Rate Limiting por Tenant

### Fase 5 ✅
- [ ] 5.1 Estandarizar Formato de Errores
- [ ] 5.2 Health Check Extendido
- [ ] 5.3 Completar Integración Stripe
- [ ] 5.4 Métricas Prometheus

---

## 🚀 Ejecución del Plan

Para ejecutar este plan, solicitar:

```
"Aplica la Fase X del plan de corrección"
```

Cada fase será implementada con:
1. Modificaciones de código
2. Tests actualizados/creados
3. Validación de cambios
4. Documentación actualizada si corresponde

---

## 📝 Notas

- **Orden recomendado**: Seguir fases secuencialmente
- **Dependencias entre fases**: Fase 3 requiere Fase 1 completada
- **Rollback**: Cada fase debe ser reversible independientemente
- **Testing**: Ejecutar tests completos tras cada fase
- **Deploy**: Preferir deploy en staging antes de producción

---

*Documento generado tras análisis de código - Febrero 2026*
