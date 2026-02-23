# Spec: treatments

Scope: feature

# Treatments Module Specification

## Overview
Catálogo de tratamientos con precios, duración, descripción.

## Domain Entities

### Treatment (schema.go)
- ID, TenantID, Name
- Description
- Category (consultation, procedure, surgery, therapy)
- Duration (minutes)
- Price
- RequiredProducts (inventory items)
- Instructions
- Active
- CreatedAt, UpdatedAt

---

## Business Logic

### Validaciones

#### Al crear tratamiento (CreateTreatmentDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido, único por tenant |
| Description | Opcional |
| Category | Requerido (consultation, procedure, surgery, therapy) |
| Duration | Requerido, > 0 (minutos) |
| Price | Requerido, >= 0 |
| RequiredProducts | Opcional, lista de IDs de productos |
| Instructions | Opcional |

### Reglas de Negocio

1. **Nombre único**: No puede existir otro tratamiento con el mismo nombre
2. **Productos válidos**: Cada ID en RequiredProducts debe existir
3. **Duración válida**: Entre 15 minutos y 8 horas
4. **Precio no negativo**: Puede ser 0 para tratamientos gratuitos
5. **Activo por defecto**: Al crear, Active = true

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| treatment.created | Tratamiento creado | Ninguna |
| treatment.updated | Tratamiento actualizado | Ninguna |
| treatment.deleted | Tratamiento eliminado | Ninguna |

### Transiciones de Estado

```
Active (por defecto)
    ↓ (soft delete)
Inactive
```

---

## Features

### Phase 1: Catalog
- [ ] CRUD treatments
- [ ] Categories
- [ ] Prices

### Phase 2: Integration
- [ ] Link to inventory (products needed)
- [ ] Link to appointments

### Phase 3: Advanced
- [ ] Treatment packages
- [ ] Templates

---

## Integration Points

- **inventory**: Products used, descontar stock al aplicar
- **appointments**: Treatment in appointment

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /treatments | Crear tratamiento |
| GET | /treatments | Listar tratamientos |
| GET | /treatments/:id | Obtener |
| PUT | /treatments/:id | Actualizar |
| DELETE | /treatments/:id | Eliminar (soft delete) |
| GET | /treatments/category/:category | Por categoría |

---

## DTOs (Data Transfer Objects)

### CreateTreatmentDTO
```go
type CreateTreatmentDTO struct {
    Name             string   `json:"name" binding:"required"`
    Description      string   `json:"description"`
    Category         string   `json:"category" binding:"required,oneof=consultation procedure surgery therapy"`
    Duration         int      `json:"duration" binding:"required,min=15,max=480"` // minutes
    Price            float64  `json:"price" binding:"required,min=0"`
    RequiredProducts []string `json:"required_products"` // Product IDs
    Instructions     string   `json:"instructions"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| TREATMENT_NAME_EXISTS | Ya existe un tratamiento con ese nombre |
| INVALID_CATEGORY | Categoría inválida |
| DURATION_INVALID | Duración fuera de rango (15-480 min) |
| PRODUCT_NOT_FOUND | Un producto no existe |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice único para Name + TenantID
- Validar que cada RequiredProducts[i] exista en inventory
