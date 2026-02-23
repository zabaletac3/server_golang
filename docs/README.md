# Sistema de Citas Veterinarias

Documentación general del proyecto. Aquí se define el contexto, tecnologías, arquitectura y reglas que rigen el desarrollo.

---

## Tecnologías

| Categoría | Tecnología |
|-----------|------------|
| **Lenguaje** | Go 1.24 |
| **Framework Web** | Gin |
| **Base de Datos** | MongoDB (go.mongodb.org/mongo-driver v1.17.6) |
| **Autenticación** | JWT (golang-jwt v5.3.0) |
| **Validación** | go-playground/validator v10 |
| **Documentación** | Swagger (swaggo v1.16.6) |
| **Notificaciones** | Firebase Cloud Messaging |
| **Pagos** | Stripe, Wompi |
| **Contenedores** | Docker, Docker Compose, Traefik |
| **Testing** | stretchr/testify |
| **Hot Reload** | Air |
| **Linting** | golangci-lint |

---

## Arquitectura de Módulos

Cada módulo sigue la estructura: `handler → service → repository → schema`

```
internal/modules/<modulo>/
├── handler.go    # HTTP handlers (Gin)
├── service.go    # Lógica de negocio
├── repository.go # Acceso a datos (MongoDB)
├── schema.go     # Modelos/Entidades
├── dto.go        # Data Transfer Objects
├── errors.go     # Errores específicos
└── router.go     # Registro de rutas
```

---

## Módulos del Sistema

| Módulo | Estado | Descripción |
|--------|--------|-------------|
| **appointments** | ✅ COMPLETO | Citas veterinarias |
| **auth** | ✅ COMPLETO | Autenticación |
| **mobile_auth** | ✅ COMPLETO | Autenticación móvil |
| **notifications** | ✅ COMPLETO | Notificaciones push/email |
| **owners** | ✅ COMPLETO | Dueños de mascotas |
| **patients** | ✅ COMPLETO | Mascotas |
| **payments** | ✅ COMPLETO | Pagos (Stripe, Wompi) |
| **permissions** | ✅ COMPLETO | Permisos RBAC |
| **plans** | ✅ COMPLETO | Planes |
| **resources** | ✅ COMPLETO | Recursos |
| **roles** | ✅ COMPLETO | Roles |
| **tenant** | ✅ COMPLETO | Multi-tenancy |
| **users** | ✅ COMPLETO | Usuarios/veterinarios |
| **health** | ⚠️ PARCIAL | Health checks |
| **webhooks** | ⚠️ PARCIAL | Webhooks de pago |

**Total: 15 módulos (13 COMPLETOS, 2 PARCIALES)**

---

## Reglas de Estilo de Código

Para mantener consistencia, siempre sigo estas reglas:

1. **Sin comentarios** en el código a menos que el usuario lo pida explícitamente
2. **Errores como valores de retorno** - siempre al final de la función, sin envolver en condiciones
3. **Naming conventions**:
   - `camelCase` para variables locales
   - `PascalCase` para funciones/types exportados
   - Interfaces con sufijo `er` (Repository, Service, Handler)
   - DTOs con sufijos `DTO`, `Response`, `Summary`
4. **Estructura de errores**: Usar errores centinela en `errors.go` del módulo

---

## Reglas de Pre-Ejecución

Antes de ejecutar做任何 cambio, siempre:

1. ✅ `go build ./...` - Verificar que compila
2. ✅ `make test` - Verificar que los tests pasan
3. ✅ `make lint` - Verificar sin warnings (si existe)

---

## Comandos Disponibles

| Comando | Descripción |
|---------|-------------|
| `make dev` | Desarrollo con hot reload (Air) |
| `make build` | Compilar binario |
| `make run` | Compilar y ejecutar |
| `make test` | Tests con race detector |
| `make lint` | Linting con golangci-lint |
| `make docs` | Generar Swagger |
| `make clean` | Limpiar build artifacts |
| `make docker-up` | Iniciar contenedores |
| `make docker-down` | Detener contenedores |
| `make docker-logs` | Ver logs de la API |
| `make logs-up` | Iniciar Loki + Grafana |
| `make logs-down` | Detener Loki + Grafana |

---

## Workflow para Nuevas Tareas

Al recibir una tarea, sigo este proceso:

1. **Entender el contexto** - Leer el README, specs, y código existente del módulo
2. **Analizar** - Ver tests existentes, estructura del módulo, dependencias
3. **Planificar** - Crear plan con pasos específicos antes de ejecutar
4. **Ejecutar** - Implementar con validaciones
5. **Verificar** - Ejecutar build, test, lint

---

## Convenciones de API

### Endpoints REST

- **Admin:** `/api/<recurso>`
- **Mobile:** `/api/mobile/<recurso>`
- **Auth:** `/api/auth/<acción>`

### Métodos HTTP

| Método | Uso |
|--------|-----|
| `GET` | Obtener recursos |
| `POST` | Crear recursos |
| `PUT` | Reemplazar recursos (completo) |
| `PATCH` | Actualizar parcialmente |
| `DELETE` | Eliminar recursos |

### Códigos de Respuesta

| Código | Significado |
|--------|-------------|
| `200` | OK |
| `201` | Creado |
| `400` | Error de validación |
| `401` | No autenticado |
| `403` | No autorizado |
| `404` | No encontrado |
| `500` | Error interno |

---

## Notificaciones

El sistema envía notificaciones a través de:

- **Push:** Firebase Cloud Messaging (FCM)
- **Email:** Servicio configurado

Los tipos de notificación se definen en `internal/modules/notifications/schema.go`.

---

## Multi-Tenancy

El sistema soporta múltiples clínicas (tenants):

- Cada documento en MongoDB tiene `tenant_id`
- El `tenant_id` se extrae del JWT o headers
- Middleware: `sharedMiddleware.GetTenantID(c)`

---

## Documentación Adicional

- [docs/api/citas.md](api/citas.md) - Comportamiento de la API de citas
- [docs/specs/](specs/) - Especificaciones técnicas
- [docs/plans/](plans/) - Planes de implementación
