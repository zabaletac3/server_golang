# Spec: promotions

Scope: feature

# Promotions Module Specification

## Overview
Gestión de descuentos, ofertas, cupones y promociones.

## Domain Entities

### Promotion (schema.go)
- ID, TenantID, Name
- Description
- Type (percentage, fixed, buy_x_get_y, coupon)
- Value (discount amount or percentage)
- ApplicableTo (all, services, products, specific_ids)
- MinPurchase (optional)
- MaxUses (optional)
- UsesCount
- StartDate, EndDate
- Code (for coupons)
- Status (active, expired, exhausted)
- CreatedAt, UpdatedAt

### CouponUsage (schema.go)
- ID, TenantID, PromotionID
- OwnerID (who used)
- AppointmentID (if used in appointment)
- UsedAt

---

## Business Logic

### Validaciones

#### Al crear promoción (CreatePromotionDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido |
| Description | Opcional |
| Type | Requerido (percentage, fixed, buy_x_get_y, coupon) |
| Value | Requerido, > 0 |
| ApplicableTo | Requerido (all, services, products, specific_ids) |
| MinPurchase | Opcional, >= 0 |
| MaxUses | Opcional, > 0 |
| StartDate | Requerido |
| EndDate | Requerido, > StartDate |
| Code | Opcional, único por tenant si se proporciona |

### Tipos de Promoción

| Tipo | Descripción | Ejemplo |
|------|-------------|---------|
| percentage | Descuento % | 10% de descuento |
| fixed | Descuento fijo | $500 de descuento |
| buy_x_get_y | Compra X lleva Y | Compra 2 lleva 1 |
| coupon | Cupón con código | SAVE20 |

### Reglas de Negocio

1. **Código único**: Si es tipo coupon, el código debe ser único por tenant
2. **Valor válido**: Percentage 1-100%, Fixed > 0
3. **Fechas válidas**: EndDate > StartDate, no pueden ser en el pasado
4. **Auto-expiración**: Cuando EndDate pasa, status = expired
5. **Auto-desactivación**: Cuando UsesCount >= MaxUses, status = exhausted
6. **Validar antes de pagar**: Siempre validar antes de completar pago
7. **Solo un cupón por cita**: Solo se puede aplicar un cupón por appointment
8. **first_visit**: Solo para clientes sin citas anteriores
9. **Mínimo de compra**: Si MinPurchase, el total debe ser >= para aplicar

### Cálculo de Descuento

```
percentage: discount = subtotal * (value / 100)
fixed: discount = min(value, subtotal) // no puede exceder el subtotal
```

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| promotion.created | Promoción creada | Ninguna |
| promotion.expired | EndDate pasado | Cambiar status |
| promotion.exhausted | MaxUses alcanzado | Cambiar status |
| promotion.used | Cupón usado | Incrementar UsesCount |

### Transiciones de Estado

```
Active (nueva promoción)
    ↓ (EndDate pasa)
Expired
```

```
Active
    ↓ (UsesCount >= MaxUses)
Exhausted
```

---

## Features

### Phase 1: Basic
- [ ] Create promotions
- [ ] Apply to appointments/payments
- [ ] Track usage

### Phase 2: Types
- [ ] Percentage discounts
- [ ] Fixed amount
- [ ] Coupon codes

### Phase 3: Advanced
- [ ] Buy X Get Y
- [ ] First visit discounts
- [ ] Referral bonuses
- [ ] Límite de usos por usuario

---

## Integration Points

- **appointments**: Apply discount to appointment
- **payments**: Apply to total, calcular descuento
- **owners**: Track usage per owner, first visit validation

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /promotions | Crear promoción |
| GET | /promotions | Listar promociones |
| GET | /promotions/:id | Obtener |
| PUT | /promotions/:id | Actualizar |
| DELETE | /promotions/:id | Eliminar (soft delete) |
| POST | /promotions/validate | Validar código de cupón |
| POST | /promotions/apply | Aplicar descuento a cita |
| GET | /promotions/:id/usage | Ver uso de promoción |

---

## DTOs (Data Transfer Objects)

### CreatePromotionDTO
```go
type CreatePromotionDTO struct {
    Name          string   `json:"name" binding:"required"`
    Description   string   `json:"description"`
    Type          string   `json:"type" binding:"required,oneof=percentage fixed buy_x_get_y coupon"`
    Value         float64  `json:"value" binding:"required,min=0"`
    ApplicableTo  string   `json:"applicable_to" binding:"required,oneof=all services products specific_ids"`
    SpecificIDs   []string `json:"specific_ids"` // IDs de servicios/productos
    MinPurchase   float64  `json:"min_purchase" binding:"min=0"`
    MaxUses       int      `json:"max_uses" binding:"min=1"`
    StartDate     string   `json:"start_date" binding:"required"` // RFC3339
    EndDate       string   `json:"end_date" binding:"required"`   // RFC3339
    Code          string   `json:"code"` // Para tipo coupon
}
```

### ValidatePromotionDTO
```go
type ValidatePromotionDTO struct {
    Code       string  `json:"code" binding:"required"`
    OwnerID    string  `json:"owner_id" binding:"required"`
    Subtotal   float64 `json:"subtotal" binding:"required,min=0"`
}
```

### ApplyPromotionDTO
```go
type ApplyPromotionDTO struct {
    Code        string `json:"code" binding:"required"`
    OwnerID     string `json:"owner_id" binding:"required"`
    AppointmentID string `json:"appointment_id" binding:"required"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| PROMOTION_NOT_FOUND | La promoción no existe |
| PROMOTION_EXPIRED | La promoción ya expiró |
| PROMOTION_EXHAUSTED | Se alcanzó el límite de usos |
| PROMOTION_NOT_ACTIVE | La promoción no está activa |
| INVALID_CODE | Código de cupón inválido |
| CODE_ALREADY_USED | Este cupón ya fue usado por el cliente |
| MIN_PURCHASE_NOT_MET | No alcanza el mínimo de compra |
| FIRST_VISIT_ONLY | Solo para primera visita |
| ALREADY_APPLIED | Ya se aplicó un cupón a esta cita |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice único para Code + TenantID (si Code no es vacío)
- Crear índice para: Status, StartDate, EndDate
- Scheduler verifica promociones expiradas diariamente
- Validar promoción ANTES de completar el pago
- Guardar CouponUsage cuando se aplica
- Para first_visit: verificar que el owner no tenga appointments anteriores
