# API de Laboratorio

Documentación de comportamiento de la API de laboratorio. Aquí se describe qué sucede en el sistema cuando realizas cada acción, incluyendo validaciones y notificaciones.

---

## Endpoints de Admin

### POST /api/lab-orders - Crear orden de laboratorio

**Propósito:** Crear una nueva orden de laboratorio para un paciente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `patient_id` sea un ID válido de MongoDB
   - Valida que `veterinarian_id` sea un ID válido de MongoDB

2. **Validaciones de campos requeridos**
   - `patient_id` - Requerido
   - `veterinarian_id` - Requerido
   - `test_type` - Requerido (blood, urine, biopsy, stool, skin, ear, other)

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - Verifica que el paciente esté activo
   - Verifica que el veterinario exista
   - Valida que `test_type` sea válido

4. **Creación**
   - Genera un ID único para la orden
   - Asigna el `owner_id` del paciente automáticamente
   - Establece `order_date` = ahora
   - Define estado inicial: `pending`

5. **Notificaciones enviadas**
   - ✅ **Owner (push):** "Orden de Laboratorio Creada" - Notifica el tipo de examen

---

### GET /api/lab-orders - Listar órdenes de laboratorio

**Propósito:** Obtener una lista paginada de órdenes de laboratorio.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `patient_id` - Filtrar por paciente
   - `veterinarian_id` - Filtrar por veterinario
   - `status` - Filtrar por estado
   - `test_type` - Filtrar por tipo de examen
   - `lab_id` - Filtrar por laboratorio externo
   - `date_from` - Fecha inicio (formato RFC3339)
   - `date_to` - Fecha fin (formato RFC3339)
   - `overdue` - Filtrar órdenes vencidas (true)

2. **Paginación**
   - Parámetros `page` y `limit` en query string
   - Retorna metadatos de paginación

3. **Ordenamiento**
   - Por defecto ordena por `order_date` descendente

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/lab-orders/:id - Obtener orden de laboratorio

**Propósito:** Obtener los detalles de una orden de laboratorio específica.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea un formato válido de MongoDB

2. **Datos retornados**
   - Retorna la orden con todos sus campos
   - Incluye fechas de colección y resultado si existen

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/lab-orders/:id - Actualizar orden de laboratorio

**Propósito:** Actualizar los datos de una orden de laboratorio existente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID de la orden sea válido

2. **Validaciones de negocio**
   - Valida que `test_type` sea válido (si se modifica)

3. **Actualización**
   - Actualiza solo los campos enviados (partial update)
   - Actualiza el campo `updated_at`

4. **Notificaciones**
   - No envía notificaciones

---

### PATCH /api/lab-orders/:id/status - Actualizar estado

**Propósito:** Cambiar el estado de una orden de laboratorio.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - **Valida que la transición de estado sea válida**

2. **Transiciones permitidas**
   - `pending` → `collected`
   - `collected` → `sent`
   - `sent` → `received`
   - `received` → `processed`
   - Estados terminales (`processed`) no pueden cambiar

3. **Acciones por estado**
   - **Collected:** Registra `collection_date`
   - **Processed:** Registra `result_date` y notifica al owner

4. **Notificaciones enviadas**
   - ✅ **Owner (push):** "Resultados de Laboratorio Listos" - Cuando status = processed

---

### POST /api/lab-orders/:id/result - Subir resultado

**Propósito:** Subir el archivo de resultados de una orden de laboratorio.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Valida que `result_file_id` sea proporcionado
   - **Verifica que no se haya subido un resultado previamente**

2. **Actualización**
   - Guarda el `result_file_id` (referencia al módulo resources)
   - Actualiza el campo `updated_at`

3. **Notificaciones**
   - No envía notificaciones (usar PATCH /status para notificar)

---

### DELETE /api/lab-orders/:id - Eliminar orden de laboratorio

**Propósito:** Eliminar (soft delete) una orden de laboratorio.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que la orden exista

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/lab-orders/patient/:patient_id - Órdenes por paciente

**Propósito:** Obtener todas las órdenes de laboratorio de un paciente específico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el `patient_id` sea un formato válido de MongoDB

2. **Paginación**
   - Soporta parámetros `page` y `limit`
   - Ordena por fecha de orden (más reciente primero)

3. **Datos retornados**
   - Historial completo de órdenes de laboratorio del paciente

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/lab-orders/overdue - Órdenes vencidas

**Propósito:** Obtener órdenes de laboratorio que están vencidas.

**Comportamiento:**

1. **Filtro**
   - Retorna órdenes donde `status != processed` y pasaron los días de turnaround
   - Default turnaround time: 5 días

2. **Datos retornados**
   - Lista de órdenes vencidas

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints de Catálogo de Exámenes

### POST /api/lab-tests - Crear examen de laboratorio

**Propósito:** Crear un nuevo examen en el catálogo del tenant.

**Comportamiento:**

1. **Validaciones de campos requeridos**
   - `name` - Requerido (2-100 caracteres)
   - `category` - Requerido
   - `price` - Requerido, >= 0
   - `turnaround_time` - Requerido, >= 1 (días)

2. **Validaciones de negocio**
   - Verifica que el nombre sea único por tenant

3. **Creación**
   - Genera un ID único para el examen
   - Establece `active = true` por defecto

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/lab-tests - Listar exámenes del catálogo

**Propósito:** Obtener todos los exámenes del catálogo.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `category` - Filtrar por categoría
   - `active` - Filtrar por estado activo
   - `search` - Buscar por nombre

2. **Datos retornados**
   - Lista de exámenes con toda su información

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/lab-tests/:id - Actualizar examen

**Propósito:** Actualizar un examen del catálogo.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que el nombre sea único (si se modifica)

2. **Actualización**
   - Actualiza solo los campos enviados

3. **Notificaciones**
   - No envía notificaciones

---

### DELETE /api/lab-tests/:id - Eliminar examen

**Propósito:** Eliminar (soft delete) un examen del catálogo.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido

2. **Eliminación**
   - Realiza un soft delete

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints Mobile (dueños de mascotas)

### GET /mobile/lab-orders/patient/:patient_id - Mis órdenes de laboratorio

**Propósito:** El cliente obtiene las órdenes de laboratorio de su mascota desde la app móvil.

**Comportamiento:**

1. **Validaciones de seguridad**
   - Verifica que el paciente pertenezca al owner autenticado
   - Si no pertenece, retorna error

2. **Datos retornados**
   - Historial de órdenes de laboratorio de la mascota

3. **Notificaciones**
   - No envía notificaciones

---

## Reglas de Negocio

### Estados de una orden de laboratorio
| Estado | Descripción |
|--------|-------------|
| `pending` | Orden creada, pendiente de recolección |
| `collected` | Muestra recolectada |
| `sent` | Enviada al laboratorio externo |
| `received` | Resultados recibidos del lab |
| `processed` | Resultados procesados y disponibles |

### Tipos de examen
| Tipo | Descripción |
|------|-------------|
| `blood` | Examen de sangre |
| `urine` | Examen de orina |
| `biopsy` | Biopsia |
| `stool` | Examen de heces |
| `skin` | Examen de piel |
| `ear` | Examen de oído |
| `other` | Otro |

### Transiciones de Estado Válidas
```
pending → collected → sent → received → processed
```

**No se permite saltar estados** - Debe seguir la secuencia completa.

### Turnaround Time
- **Default**: 5 días para procesar una orden
- **Configurable**: Por examen en el catálogo
- **Alertas**: Se envían cuando una orden está vencida

### Archivo de Resultados
- Se almacena como referencia al módulo `resources`
- Solo se puede subir UN resultado por orden
- Formatos soportados: PDF

---

## Notificaciones Resumen

| Evento | Notificación | Destinatario | Trigger |
|--------|--------------|--------------|---------|
| Orden creada | "Orden de Laboratorio Creada" | Owner | Al crear orden |
| Resultados listos | "Resultados de Laboratorio Listos" | Owner | Status = processed |
| Orden vencida | "Orden de Laboratorio Vencida" | Staff | Scheduler diario |

---

## Scheduler (Tareas Automáticas)

El sistema ejecuta las siguientes tareas automáticas:

### Diariamente (cada 24 horas)
```
1. Buscar órdenes con status != processed y order_date + turnaround_days < hoy
   → Enviar alerta al staff
```

---

## Errores Específicos del Módulo

| Código | HTTP | Descripción |
|--------|------|-------------|
| `LAB_ORDER_NOT_FOUND` | 404 | La orden de laboratorio no existe |
| `LAB_TEST_NOT_FOUND` | 404 | El examen no existe |
| `LAB_TEST_NAME_EXISTS` | 409 | El nombre de examen ya existe |
| `PATIENT_NOT_FOUND` | 404 | El paciente no existe |
| `PATIENT_INACTIVE` | 400 | El paciente está inactivo |
| `VETERINARIAN_NOT_FOUND` | 404 | El veterinario no existe |
| `INVALID_STATUS` | 400 | Estado de orden inválido |
| `INVALID_TEST_TYPE` | 400 | Tipo de examen inválido |
| `INVALID_STATUS_TRANSITION` | 400 | Transición de estado inválida |
| `RESULT_ALREADY_UPLOADED` | 400 | Ya se subió un resultado |
| `RESULT_REQUIRED` | 400 | Se requiere archivo de resultado |

---

## Integración con Otros Módulos

### Medical Records
- Los resultados de laboratorio pueden adjuntarse al historial médico
- Relación mediante `patient_id`

### Resources
- Los archivos de resultados se almacenan como recursos
- `result_file_id` es una referencia al módulo resources

### Appointments
- Puede crear cita para toma de muestra
- Relación mediante `patient_id` y `owner_id`

### Notifications
- Notificaciones push/email para resultados listos
- Notificaciones al owner y al staff

---

## Ejemplos de Uso

### Crear orden de laboratorio
```json
POST /api/lab-orders
{
  "patient_id": "507f1f77bcf86cd799439011",
  "veterinarian_id": "507f1f77bcf86cd799439022",
  "lab_id": "LabVet Central",
  "test_type": "blood",
  "notes": "Hemograma completo - Ayuno 12h",
  "cost": 45.00
}
```

### Actualizar estado a collected
```json
PATCH /api/lab-orders/:id/status
{
  "status": "collected",
  "collection_date": "2026-02-24T10:00:00Z",
  "notes": "Muestra recolectada exitosamente"
}
```

### Subir resultado
```json
POST /api/lab-orders/:id/result
{
  "result_file_id": "507f1f77bcf86cd799439033",
  "notes": "Resultados dentro de rangos normales"
}
```

### Actualizar estado a processed (notifica al owner)
```json
PATCH /api/lab-orders/:id/status
{
  "status": "processed",
  "notes": "Resultados revisados por el veterinario"
}
```

### Crear examen del catálogo
```json
POST /api/lab-tests
{
  "name": "Hemograma Completo",
  "description": "Conteo completo de células sanguíneas",
  "category": "hematology",
  "price": 45.00,
  "turnaround_time": 2,
  "active": true
}
```

### Obtener órdenes vencidas
```
GET /api/lab-orders/overdue

Response:
{
  "data": [
    {
      "id": "...",
      "patient_id": "...",
      "test_type": "blood",
      "status": "sent",
      "order_date": "2026-02-10T10:00:00Z",
      "days_since_order": 14
    }
  ]
}
```

---

## Consideraciones de Seguridad

### Multi-tenancy
- Todas las órdenes están aisladas por `tenant_id`
- El catálogo de exámenes es por tenant

### RBAC (Roles y Permisos)
| Rol | Permisos lab-orders | Permisos lab-tests |
|-----|---------------------|-------------------|
| `admin` | CRUD | CRUD |
| `veterinarian` | CRUD | CRUD |
| `receptionist` | Leer, Crear, Update status | Leer |
| `assistant` | Leer | Leer |
| `accountant` | Leer | Sin acceso |

### Mobile (Owners)
- Los owners solo pueden VER las órdenes de SUS mascotas
- No pueden crear, editar ni eliminar órdenes
- El acceso se valida mediante `OwnerGuardMiddleware`

### Auditoría
- Todas las órdenes tienen `order_date`, `collection_date`, `result_date`
- Soft delete mantiene los datos para auditoría
- Trazabilidad completa del estado
