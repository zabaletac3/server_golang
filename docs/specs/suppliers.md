# Spec: suppliers

Scope: feature

# Suppliers Module Specification

## Overview
Gestión de proveedores de medicamentos y productos.

## Domain Entities

### Supplier (schema.go)
- ID, TenantID, Name
- ContactName
- Email, Phone
- Address
- Notes
- Products (list of product names/codes they supply)
- Active
- CreatedAt, UpdatedAt

### PurchaseOrder (schema.go)
- ID, TenantID, SupplierID
- OrderDate
- Status (pending, ordered, received, cancelled)
- Items (product, quantity, unit price)
- Total
- ReceivedDate
- Notes

---

## Business Logic

### Validaciones

#### Al crear proveedor (CreateSupplierDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido, único por tenant |
| ContactName | Opcional |
| Email | Opcional, formato válido |
| Phone | Opcional |
| Address | Opcional |

#### Al crear orden de compra (CreatePurchaseOrderDTO)
| Campo | Regla |
|-------|-------|
| SupplierID | Requerido, debe existir y estar activo |
| Items | Requerido, al menos 1 |
| Notes | Opcional |

#### Items de orden
| Campo | Regla |
|-------|-------|
| ProductName | Requerido |
| Quantity | Requerido, > 0 |
| UnitPrice | Requerido, >= 0 |

### Reglas de Negocio

1. **Proveedor activo**: Solo se puede comprar a proveedores activos
2. **Items mínimo**: Una orden debe tener al menos 1 item
3. **Total calculado**: Es la suma de (Quantity * UnitPrice) de todos los items
4. **Estados válidos**: pending → ordered → received, o cualquier → cancelled
5. **Recibir orden**: Al recibir, se descuenta del inventario
6. **No duplicar**: No se puede recibir una orden ya recibida

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| supplier.created | Proveedor creado | Ninguna |
| supplier.deleted | Proveedor eliminado | Ninguna |
| order.created | Orden creada | Ninguna |
| order.status.changed | Cambio de estado | Ninguna |
| order.received | Orden recibida | Actualizar inventario |

### Transiciones de Estado (PurchaseOrder)

```
pending (orden creada)
    ↓ (se envía al proveedor)
ordered
    ↓ (se recibe la orden)
received
```

```
pending/ordered
    ↓ (se cancela)
cancelled
```

---

## Features

### Phase 1: Basic
- [ ] CRUD suppliers
- [ ] Contact info

### Phase 2: Orders
- [ ] Create purchase orders
- [ ] Track status
- [ ] Receive orders (update inventory)

### Phase 3: Advanced
- [ ] Order history
- [ ] Supplier reports

---

## Integration Points

- **inventory**: Link products to suppliers, update stock on receive
- **users**: Who ordered

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /suppliers | Crear proveedor |
| GET | /suppliers | Listar proveedores |
| GET | /suppliers/:id | Obtener |
| PUT | /suppliers/:id | Actualizar |
| DELETE | /suppliers/:id | Eliminar (soft delete) |
| POST | /purchase-orders | Crear orden |
| GET | /purchase-orders | Listar órdenes |
| GET | /purchase-orders/:id | Obtener orden |
| PUT | /purchase-orders/:id/status | Cambiar estado |
| POST | /purchase-orders/:id/receive | Recibir orden |
| GET | /purchase-orders/supplier/:supplier_id | Órdenes por proveedor |

---

## DTOs (Data Transfer Objects)

### CreateSupplierDTO
```go
type CreateSupplierDTO struct {
    Name        string `json:"name" binding:"required"`
    ContactName string `json:"contact_name"`
    Email       string `json:"email" binding:"omitempty,email"`
    Phone       string `json:"phone"`
    Address     string `json:"address"`
    Notes       string `json:"notes"`
    Products    []string `json:"products"`
}
```

### CreatePurchaseOrderDTO
```go
type CreatePurchaseOrderDTO struct {
    SupplierID string        `json:"supplier_id" binding:"required"`
    Items     []OrderItemDTO `json:"items" binding:"required,min=1"`
    Notes     string        `json:"notes"`
}

type OrderItemDTO struct {
    ProductName string  `json:"product_name" binding:"required"`
    Quantity    int     `json:"quantity" binding:"required,min=1"`
    UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
}
```

### UpdateOrderStatusDTO
```go
type UpdateOrderStatusDTO struct {
    Status string `json:"status" binding:"required,oneof=pending ordered received cancelled"`
    Notes  string `json:"notes"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| SUPPLIER_NOT_FOUND | El proveedor no existe |
| SUPPLIER_INACTIVE | El proveedor está inactivo |
| SUPPLIER_NAME_EXISTS | Ya existe un proveedor con ese nombre |
| INVALID_ORDER_ITEMS | La orden debe tener al menos 1 item |
| INVALID_STATUS_TRANSITION | No se puede cambiar a este estado |
| ORDER_ALREADY_RECEIVED | La orden ya fue recibida |
| PRODUCT_REQUIRED | El nombre del producto es requerido |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete para suppliers
- Crear índice único para Name + TenantID
- Al recibir orden: por cada item, hacer stock-in en inventory
- Calcular Total = suma(Quantity * UnitPrice) automáticamente
