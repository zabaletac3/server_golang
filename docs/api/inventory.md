# API de Inventario

Documentación de comportamiento de la API de inventario. Aquí se describe qué sucede en el sistema cuando realizas cada acción, incluyendo validaciones y notificaciones.

---

## Endpoints de Admin

### POST /api/products - Crear producto

**Propósito:** Crear un nuevo producto en el inventario.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `category_id` sea un ID válido de MongoDB (si se proporciona)
   - Valida que `supplier_id` sea un ID válido de MongoDB (si se proporciona)

2. **Validaciones de campos requeridos**
   - `name` - Requerido (3-100 caracteres)
   - `sku` - Requerido (1-50 caracteres)
   - `category` - Requerido (medicine, supply, food, equipment)
   - `unit` - Requerido (tablet, ml, piece, kg, gram, box, bottle)
   - `purchase_price` - Requerido, >= 0
   - `sale_price` - Requerido, >= 0

3. **Validaciones de negocio**
   - Verifica que la categoría exista (si se proporciona)
   - Valida que `sale_price` >= `purchase_price` (margen 0+)
   - Valida que `stock` >= 0
   - Valida que `min_stock` >= 0
   - Valida que `expiration_date` sea formato RFC3339 (si se proporciona)
   - Verifica que el SKU sea único por tenant

4. **Creación**
   - Genera un ID único para el producto
   - Establece `active = true` por defecto
   - Establece `created_at` y `updated_at`

5. **Notificaciones**
   - No envía notificaciones

---

### GET /api/products - Listar productos

**Propósito:** Obtener una lista paginada de productos.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `category` - Filtrar por categoría (medicine, supply, food, equipment)
   - `active` - Filtrar por estado activo (true/false)
   - `low_stock` - Filtrar productos con stock bajo (true)
   - `expiring` - Filtrar productos próximos a vencer (true)
   - `expired` - Filtrar productos vencidos (true)
   - `search` - Buscar por nombre, SKU o barcode

2. **Paginación**
   - Parámetros `page` y `limit` en query string
   - Retorna metadatos de paginación

3. **Ordenamiento**
   - Por defecto ordena por `name` ascendente

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/products/:id - Obtener producto

**Propósito:** Obtener los detalles de un producto específico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea un formato válido de MongoDB

2. **Datos retornados**
   - Retorna el producto con todos sus campos
   - Incluye información de stock y vencimiento

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/products/:id - Actualizar producto

**Propósito:** Actualizar los datos de un producto existente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID del producto sea válido

2. **Validaciones de negocio**
   - Valida que `sale_price` >= `purchase_price` (si ambos se modifican)
   - Verifica que el SKU sea único (si se modifica)
   - Verifica que el barcode sea único (si se modifica y no está vacío)

3. **Actualización**
   - Actualiza solo los campos enviados (partial update)
   - Actualiza el campo `updated_at`

4. **Notificaciones**
   - No envía notificaciones

---

### DELETE /api/products/:id - Eliminar producto

**Propósito:** Eliminar (soft delete) un producto.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - **Verifica que el stock sea 0** (no se puede eliminar con stock > 0)

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)
   - No elimina físicamente el documento

3. **Notificaciones**
   - No envía notificaciones

---

### POST /api/products/:id/stock-in - Agregar stock

**Propósito:** Agregar stock a un producto (compra, devolución, ajuste).

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID del producto sea válido

2. **Validaciones de campos requeridos**
   - `quantity` - Requerido, > 0
   - `reason` - Requerido (purchase, return, adjustment)

3. **Validaciones de negocio**
   - Valida que el reason sea válido

4. **Actualización de stock**
   - Suma la cantidad al stock actual
   - Registra `stock_before` y `stock_after`

5. **Registro de movimiento**
   - Crea un registro en `stock_movements`
   - Incluye usuario que realizó la acción

6. **Notificaciones**
   - No envía notificaciones

---

### POST /api/products/:id/stock-out - Descontar stock

**Propósito:** Descontar stock de un producto (venta, tratamiento, vencimiento).

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID del producto sea válido

2. **Validaciones de campos requeridos**
   - `quantity` - Requerido, > 0
   - `reason` - Requerido (sale, treatment, adjustment, expired, damaged, lost)

3. **Validaciones de negocio CRÍTICAS**
   - **Verifica que haya stock suficiente**
   - Retorna `INSUFFICIENT_STOCK` si no hay suficiente

4. **Actualización de stock**
   - Resta la cantidad del stock actual
   - Registra `stock_before` y `stock_after`

5. **Registro de movimiento**
   - Crea un registro en `stock_movements`
   - Incluye usuario que realizó la acción

6. **Notificaciones**
   - No envía notificaciones

---

### GET /api/products/low-stock - Productos con stock bajo

**Propósito:** Obtener productos con stock por debajo del mínimo.

**Comportamiento:**

1. **Filtro**
   - Retorna productos donde `stock <= min_stock`
   - Solo productos activos

2. **Datos retornados**
   - Lista de productos con stock bajo

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/products/expiring - Productos por vencer

**Propósito:** Obtener productos que vencen en los próximos 30 días.

**Comportamiento:**

1. **Filtro**
   - Retorna productos donde `expiration_date` está entre hoy y +30 días
   - Solo productos activos

2. **Datos retornados**
   - Lista de productos próximos a vencer

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/products/alerts - Todas las alertas

**Propósito:** Obtener todas las alertas de productos (stock bajo, vencimiento, vencidos).

**Comportamiento:**

1. **Tipos de alertas**
   - `low_stock` - Stock por debajo del mínimo
   - `expiring` - Vence en los próximos 30 días
   - `expired` - Ya venció

2. **Datos retornados**
   - Lista unificada de alertas con:
     - `product_id`, `product_name`
     - `alert_type`
     - `current_stock`, `min_stock`
     - `stock_difference` (negativo si está por debajo)
     - `expiration_date`, `days_until_expiry`

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/stock-movements - Listar movimientos de stock

**Propósito:** Obtener el historial de movimientos de stock.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `product_id` - Filtrar por producto
   - `type` - Filtrar por tipo (in, out, adjustment, expired)
   - `reason` - Filtrar por reason
   - `date_from` - Fecha inicio (RFC3339)
   - `date_to` - Fecha fin (RFC3339)
   - `reference_id` - Filtrar por referencia (appointment, order, etc.)

2. **Paginación**
   - Parámetros `page` y `limit`
   - Ordena por `created_at` descendente

3. **Datos retornados**
   - Movimientos con: tipo, razón, cantidad, stock antes/después, usuario, notas

4. **Notificaciones**
   - No envía notificaciones

---

## Endpoints de Categorías

### POST /api/categories - Crear categoría

**Propósito:** Crear una nueva categoría de productos.

**Comportamiento:**

1. **Validaciones de campos requeridos**
   - `name` - Requerido (2-100 caracteres)

2. **Validaciones de negocio**
   - Verifica que la categoría padre exista (si se proporciona `parent_id`)
   - Verifica que el nombre sea único por tenant

3. **Creación**
   - Genera un ID único para la categoría

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/categories - Listar categorías

**Propósito:** Obtener todas las categorías del tenant.

**Comportamiento:**

1. **Datos retornados**
   - Lista de categorías con nombre, descripción, padre

2. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/categories/:id - Actualizar categoría

**Propósito:** Actualizar una categoría existente.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que el nombre sea único (si se modifica)

2. **Actualización**
   - Actualiza solo los campos enviados

3. **Notificaciones**
   - No envía notificaciones

---

### DELETE /api/categories/:id - Eliminar categoría

**Propósito:** Eliminar (soft delete) una categoría.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido

2. **Eliminación**
   - Realiza un soft delete

3. **Notificaciones**
   - No envía notificaciones

---

## Reglas de Negocio

### Categorías de productos
| Categoría | Descripción |
|-----------|-------------|
| `medicine` | Medicamentos y fármacos |
| `supply` | Insumos y materiales |
| `food` | Alimentos y nutrición |
| `equipment` | Equipos y dispositivos |

### Unidades de medida
| Unidad | Descripción |
|--------|-------------|
| `tablet` | Tabletas, pastillas |
| `ml` | Mililitros (líquidos) |
| `piece` | Piezas individuales |
| `kg` | Kilogramos |
| `gram` | Gramos |
| `box` | Cajas |
| `bottle` | Frascos/botellas |

### Tipos de movimiento de stock
| Tipo | Descripción |
|------|-------------|
| `in` | Entrada de stock |
| `out` | Salida de stock |
| `adjustment` | Ajuste de inventario |
| `expired` | Producto vencido |

### Razones de movimiento
| Razón | Descripción |
|-------|-------------|
| `purchase` | Compra a proveedor |
| `sale` | Venta a cliente |
| `treatment` | Uso en tratamiento |
| `adjustment` | Ajuste manual |
| `return` | Devolución |
| `expired` | Producto vencido |
| `damaged` | Producto dañado |
| `lost` | Producto perdido |

### Reglas de precio
- **Sale price >= Purchase price** (no se puede vender por debajo del costo)
- **Precios en centavos** para evitar problemas de redondeo

### Reglas de stock
- **Stock no puede ser negativo**
- **No se puede eliminar producto con stock > 0**
- **Low stock**: cuando `stock <= min_stock`

### Vencimientos
- **Expiring soon**: dentro de los próximos 30 días
- **Expired**: fecha de vencimiento < hoy
- **Productos vencidos no se pueden vender** (deben darse de baja)

---

## Notificaciones Resumen

| Evento | Notificación | Destinatario | Trigger |
|--------|--------------|--------------|---------|
| Stock bajo | "Stock Bajo - [producto]" | Staff (todos) | `stock <= min_stock` |
| Producto por vencer | "Producto por Vencer - [producto]" | Staff (todos) | 30 días antes de vencer |

**Nota**: Las notificaciones se envían automáticamente mediante el scheduler (ver más abajo).

---

## Errores Específicos del Módulo

| Código | HTTP | Descripción |
|--------|------|-------------|
| `PRODUCT_NOT_FOUND` | 404 | El producto no existe |
| `CATEGORY_NOT_FOUND` | 404 | La categoría no existe |
| `CATEGORY_NAME_EXISTS` | 409 | El nombre de categoría ya existe |
| `SKU_ALREADY_EXISTS` | 409 | El SKU ya existe |
| `BARCODE_ALREADY_EXISTS` | 409 | El barcode ya existe |
| `INSUFFICIENT_STOCK` | 400 | Stock insuficiente para la salida |
| `CANNOT_DELETE_WITH_STOCK` | 400 | No se puede eliminar producto con stock > 0 |
| `SALE_PRICE_TOO_LOW` | 400 | El precio de venta debe ser >= al precio de compra |
| `INVALID_TEMPERATURE` | 400 | Temperatura fuera de rango (30-45°C) |

---

## Scheduler (Tareas Automáticas)

El sistema ejecuta las siguientes tareas automáticas:

### Diariamente (cada 24 horas)
```
1. Verificar productos con stock bajo
   → Enviar notificación al staff

2. Verificar productos por vencer (30 días)
   → Enviar notificación al staff

3. Verificar productos vencidos
   → Enviar notificación al staff
   → Marcar como no disponibles
```

---

## Ejemplos de Uso

### Crear producto
```json
POST /api/products
{
  "name": "Amoxicilina 500mg",
  "sku": "AMO-500-001",
  "barcode": "7501234567890",
  "category": "medicine",
  "unit": "tablet",
  "purchase_price": 5.50,
  "sale_price": 12.00,
  "stock": 100,
  "min_stock": 20,
  "expiration_date": "2027-06-15T00:00:00Z",
  "active": true
}
```

### Agregar stock (compra)
```json
POST /api/products/:id/stock-in
{
  "quantity": 50,
  "reason": "purchase",
  "reference_id": "PO-2026-001",
  "notes": "Compra a proveedor FarmaVet"
}
```

### Descontar stock (venta)
```json
POST /api/products/:id/stock-out
{
  "quantity": 2,
  "reason": "sale",
  "reference_id": "INV-2026-001",
  "notes": "Venta en mostrador"
}
```

### Obtener alertas
```
GET /api/products/alerts

Response:
{
  "data": [
    {
      "product_id": "...",
      "product_name": "Amoxicilina 500mg",
      "alert_type": "low_stock",
      "current_stock": 15,
      "min_stock": 20,
      "stock_difference": -5
    },
    {
      "product_id": "...",
      "product_name": "Suero Fisiológico",
      "alert_type": "expiring",
      "current_stock": 50,
      "expiration_date": "2026-03-15T00:00:00Z",
      "days_until_expiry": 20
    }
  ]
}
```

---

## Integración con Otros Módulos

### Appointments
- Los tratamientos pueden consumir productos del inventario
- Al completar una cita con tratamiento → stock-out automático

### Billing
- Los productos vendidos se registran en la factura
- Cada producto vendido → stock-out con reason = "sale"

### Medical Records
- Los medicamentos recetados pueden descontarse del inventario
- Relación con `products` mediante `supplier_id`

---

## Consideraciones de Seguridad

### Multi-tenancy
- Todos los productos están aislados por `tenant_id`
- SKU único por tenant (no global)

### RBAC (Roles y Permisos)
| Rol | Permisos inventory |
|-----|-------------------|
| `admin` | get, post, put, patch, delete |
| `veterinarian` | get, post, put, patch |
| `receptionist` | get, post (stock operations) |
| `assistant` | get, post (stock-in only) |
| `accountant` | get |

### Auditoría
- Todos los movimientos de stock quedan registrados
- Incluye usuario, fecha, razón, referencia
- Stock antes y después para trazabilidad
