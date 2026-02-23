# Spec: inventory

Scope: feature

# Inventory Module Specification

## Overview
Control de inventario de medicamentos, productos veterinarios y suministros médicos.

## Domain Entities

### Product (schema.go)
- ID, TenantID, Name, Description
- SKU, Barcode
- Category (medicine, supply, food, equipment)
- Unit (tablet, ml, piece, kg)
- PurchasePrice, SalePrice
- Stock (current quantity)
- MinStock (alert threshold)
- ExpirationDate
- SupplierID (optional)
- Active (bool)
- CreatedAt, UpdatedAt

### Category (schema.go)
- ID, TenantID, Name, Description
- ParentID (for subcategories)

---

## Business Logic

### Validaciones

#### Al crear producto (CreateProductDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido, 3-100 caracteres |
| SKU | Requerido, único por tenant |
| Barcode | Opcional, único por tenant |
| Category | Requerido, debe existir |
| Unit | Requerido |
| PurchasePrice | Requerido, >= 0 |
| SalePrice | Requerido, >= 0 |
| Stock | >= 0 |
| MinStock | >= 0, si no se envía, default 0 |

#### Al actualizar stock (StockInDTO / StockOutDTO)
| Campo | Regla |
|-------|-------|
| Quantity | Requerido, > 0 |
| Reason | Requerido (compra, ajuste, venta, tratamiento, etc.) |
| ReferenceID | Opcional (ID de orden de compra, cita, etc.) |

### Reglas de Negocio

1. **Precio de venta debe ser mayor o igual al precio de compra** (margen 0+)
2. **Stock no puede ser negativo**
3. **No se puede eliminar producto con stock > 0** (soft delete con flag active=false)
4. **Código SKU debe ser único por tenant**
5. **Código de barras debe ser único por tenant** (si se proporciona)
6. **Alerta de stock bajo**: cuando Stock <= MinStock
7. **Producto próximo a vencer**: cuando ExpirationDate <= 30 días

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `product.created` | Se crea un producto | Ninguna |
| `product.updated` | Se actualiza un producto | Ninguna |
| `product.stock.low` | Stock <= MinStock | Enviar notificación al staff |
| `product.expiring` | Expira en 30 días | Enviar notificación al staff |
| `product.expired` | Fecha de expiración pasada | Marcar como no disponible |
| `product.deleted` | Soft delete | Ninguna |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `inventory.low_stock` | Staff (todos) | Stock <= MinStock |
| `inventory.expiring` | Staff (todos) | 30 días antes de expirar |
| `inventory.expired` | Staff (todos) | Producto expirado |

### Transiciones de Estado

```
Active (por defecto)
    ↓ (soft delete)
Inactive (active=false)
    ↓ (stock > 0)
No se puede eliminar si hay stock
```

---

## Features

### Phase 1: Basic CRUD
- [ ] Create, Read, Update, Delete products
- [ ] List with filters (category, active, low stock)
- [ ] Search by name, SKU, barcode
- [ ] Unique SKU/Barcode validation

### Phase 2: Stock Management
- [ ] Stock in (add inventory) - con reason
- [ ] Stock out (deduct for treatments/sales) - con reason
- [ ] Stock adjustment with reason
- [ ] Low stock alerts (notifications)
- [ ] Expiration tracking

### Phase 3: Advanced
- [ ] Product categories with hierarchy
- [ ] Supplier management link
- [ ] Cost/price tracking
- [ ] Reports (expiring soon, low stock, valuation)

---

## Integration Points

- **appointments**: Consumir productos en tratamientos
- **patients**: Productos por mascota
- **payments**: Productos en facturas
- **notifications**: Enviar alertas de stock bajo y expiración

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /products | Crear producto |
| GET | /products | Listar productos (con filtros) |
| GET | /products/:id | Obtener producto |
| PUT | /products/:id | Actualizar producto |
| DELETE | /products/:id | Eliminar (soft delete) |
| POST | /products/:id/stock-in | Agregar stock |
| POST | /products/:id/stock-out | Descontar stock |
| GET | /products/low-stock | Productos con stock bajo |
| GET | /products/expiring | Productos próximos a vencer |
| POST | /categories | Crear categoría |
| GET | /categories | Listar categorías |
| PUT | /categories/:id | Actualizar categoría |
| DELETE | /categories/:id | Eliminar categoría |

### Mobile (future)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /mobile/products | Listar productos disponibles |

---

## DTOs (Data Transfer Objects)

### CreateProductDTO
```go
type CreateProductDTO struct {
    Name          string  `json:"name" binding:"required,min=3,max=100"`
    Description   string  `json:"description"`
    SKU           string  `json:"sku" binding:"required"`
    Barcode       string  `json:"barcode"`
    Category      string  `json:"category" binding:"required"`
    Unit          string  `json:"unit" binding:"required"`
    PurchasePrice float64 `json:"purchase_price" binding:"required,min=0"`
    SalePrice     float64 `json:"sale_price" binding:"required,min=0"`
    Stock         int     `json:"stock" binding:"min=0"`
    MinStock      int     `json:"min_stock" binding:"min=0"`
    ExpirationDate string `json:"expiration_date"`
    SupplierID    string  `json:"supplier_id"`
}
```

### StockInDTO
```go
type StockInDTO struct {
    Quantity    int    `json:"quantity" binding:"required,min=1"`
    Reason      string `json:"reason" binding:"required"` // compra, ajuste, devolución
    ReferenceID string `json:"reference_id"` // ID de orden de compra
}
```

### StockOutDTO
```go
type StockOutDTO struct {
    Quantity    int    `json:"quantity" binding:"required,min=1"`
    Reason      string `json:"reason" binding:"required"` // venta, tratamiento, muestra, caducado
    ReferenceID string `json:"reference_id"` // ID de cita, factura
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| SKU_ALREADY_EXISTS | El SKU ya existe |
| BARCODE_ALREADY_EXISTS | El código de barras ya existe |
| INSUFFICIENT_STOCK | Stock insuficiente para descontar |
| CATEGORY_NOT_FOUND | La categoría no existe |
| CANNOT_DELETE_WITH_STOCK | No se puede eliminar producto con stock |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Use existing validation, errors, dto patterns
- Multi-tenant isolation required
- Soft delete pattern
- Crear índice único para SKU y Barcode
- Crear índice para ExpirationDate (para queries de productos por vencer)
- El scheduler debe verificar productos con stock bajo y por vencer diariamente
