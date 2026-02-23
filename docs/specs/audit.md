# Spec: audit

Scope: feature

# Audit Module Specification

## Overview
Log de auditoría para compliance: registra todas las acciones de los usuarios en el sistema.

## Domain Entities

### AuditLog (schema.go)
- ID, TenantID
- UserID (who performed action)
- UserName, UserEmail
- Action (created, updated, deleted, login, logout, status_changed)
- Resource (appointment, patient, user, etc.)
- ResourceID
- Changes (JSON - before/after)
- IPAddress
- UserAgent
- Timestamp

---

## Business Logic

### Reglas de Negocio

1. **Logs inmutables**: No se pueden editar ni eliminar
2. **Captura automática**: Se genera vía middleware
3. **Retención**: TTL configurable (default 90 días)
4. **Datos capturados**:
   - Usuario que realizó la acción
   - Tipo de acción
   - Recurso afectado
   - ID del recurso
   - Cambios (antes/después) para update
   - IP y UserAgent
5. **Sin duplicados**: Un log por acción
6. **Indexación**: Por Timestamp, UserID, Resource, ResourceID

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| audit.action.performed | Cualquier acción de usuario | Crear log |

### Transiciones de Estado

```
No aplica - AuditLog es inmutable
```

---

## Features

### Phase 1: Basic Logging
- [ ] Log all CRUD operations
- [ ] Log authentication events
- [ ] Searchable logs

### Phase 2: Retention
- [ ] Configurable retention period
- [ ] Export logs (CSV/JSON)

### Phase 3: Advanced
- [ ] Dashboard of user activity
- [ ] Suspicious activity alerts
- [ ] Detección de anomalías

---

## Actions Logged

| Categoría | Acciones |
|-----------|----------|
| Auth | login, logout, login_failed |
| CRUD | created, updated, deleted |
| Estado | status_changed |
| Seguridad | role_changed, permission_changed, password_changed |
| Módulo | Depende del recurso |

---

## Integration Points

- **All modules**: Intercept via middleware
- **users**: Link to user
- **scheduler**: TTL cleanup

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /audit-logs | Listar logs (paginado) |
| GET | /audit-logs/:id | Obtener log |
| GET | /audit-logs/export | Exportar a CSV/JSON |
| GET | /audit-logs/user/:user_id | Actividad por usuario |
| GET | /audit-logs/resource/:resource | Por tipo de recurso |
| GET | /audit-logs/resource/:resource/:resource_id | Por recurso específico |
| GET | /audit-logs/search | Búsqueda avanzada |

---

## DTOs (Data Transfer Objects)

### AuditLogQueryDTO
```go
type AuditLogQueryDTO struct {
    UserID     string `form:"user_id"`
    Action     string `form:"action"`
    Resource   string `form:"resource"`
    ResourceID string `form:"resource_id"`
    DateFrom   string `form:"date_from"` // RFC3339
    DateTo     string `form:"date_to"`   // RFC3339
    Page       int    `form:"page"`
    Limit      int    `form:"limit"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| LOG_NOT_FOUND | El log no existe |
| EXPORT_TOO_LARGE | Máximo 10,000 registros para exportar |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Read-only - no update/delete
- Crear índice compuesto: TenantID + Timestamp
- Crear índice para: UserID, Resource, ResourceID
- TTL index para auto-cleanup (90 días por defecto)
- Middleware que capture todas las requests
- Guardar solo diff (campos cambiados), no el objeto completo
