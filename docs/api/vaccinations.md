# API de Vacunación

Documentación de comportamiento de la API de vacunación. Aquí se describe qué sucede en el sistema cuando realizas cada acción, incluyendo validaciones y notificaciones.

---

## Endpoints de Admin

### POST /api/vaccinations - Registrar vacunación

**Propósito:** Registrar una nueva vacunación para un paciente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `patient_id` sea un ID válido de MongoDB
   - Valida que `veterinarian_id` sea un ID válido de MongoDB

2. **Validaciones de campos requeridos**
   - `patient_id` - Requerido
   - `veterinarian_id` - Requerido
   - `vaccine_name` - Requerido (1-100 caracteres)
   - `application_date` - Requerido (formato RFC3339)

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - Verifica que el paciente esté activo
   - Verifica que el veterinario exista
   - **La fecha de aplicación no puede ser en el futuro**
   - Valida que `next_due_date` sea posterior a `application_date` (si se proporciona)

4. **Determinación automática del estado**
   - `applied` - Fecha de próxima vacunación no vencida
   - `due` - Próxima vacunación dentro de los próximos 30 días
   - `overdue` - Fecha de próxima vacunación ya pasó

5. **Generación de certificado**
   - Genera un `certificate_number` único: `VAC-{tenant_id}-{fecha}`

6. **Creación**
   - Genera un ID único para la vacunación
   - Asigna el `owner_id` del paciente automáticamente

7. **Notificaciones enviadas**
   - ✅ **Owner (push):** "Vacunación Registrada" - Notifica la vacuna aplicada

---

### GET /api/vaccinations - Listar vacunaciones

**Propósito:** Obtener una lista paginada de vacunaciones.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `patient_id` - Filtrar por paciente
   - `veterinarian_id` - Filtrar por veterinario
   - `status` - Filtrar por estado (applied, due, overdue)
   - `vaccine_name` - Filtrar por nombre de vacuna
   - `date_from` - Fecha inicio (formato RFC3339)
   - `date_to` - Fecha fin (formato RFC3339)
   - `due_soon` - Filtrar vacunaciones próximas a vencer (30 días)
   - `overdue` - Filtrar vacunaciones vencidas

2. **Paginación**
   - Parámetros `page` y `limit` en query string
   - Retorna metadatos de paginación

3. **Ordenamiento**
   - Por defecto ordena por `application_date` descendente

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/vaccinations/:id - Obtener vacunación

**Propósito:** Obtener los detalles de una vacunación específica.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea un formato válido de MongoDB

2. **Datos retornados**
   - Retorna la vacunación con todos sus campos
   - Incluye número de certificado

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/vaccinations/:id - Actualizar vacunación

**Propósito:** Actualizar los datos de una vacunación existente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID de la vacunación sea válido

2. **Validaciones de negocio**
   - Valida que `next_due_date` sea posterior a `application_date` (si se modifica)

3. **Actualización automática del estado**
   - Si se modifica `next_due_date`, recalcula el estado automáticamente

4. **Actualización**
   - Actualiza solo los campos enviados (partial update)
   - Actualiza el campo `updated_at`

5. **Notificaciones**
   - No envía notificaciones

---

### PATCH /api/vaccinations/:id/status - Actualizar estado

**Propósito:** Cambiar el estado de una vacunación.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Valida que el estado sea válido (applied, due, overdue)

2. **Actualización**
   - Actualiza el estado de la vacunación

3. **Notificaciones**
   - No envía notificaciones

---

### DELETE /api/vaccinations/:id - Eliminar vacunación

**Propósito:** Eliminar (soft delete) una vacunación.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que la vacunación exista

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/vaccinations/patient/:patient_id - Vacunaciones por paciente

**Propósito:** Obtener todas las vacunaciones de un paciente específico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el `patient_id` sea un formato válido de MongoDB

2. **Paginación**
   - Soporta parámetros `page` y `limit`
   - Ordena por fecha de aplicación (más reciente primero)

3. **Datos retornados**
   - Historial completo de vacunaciones del paciente

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/vaccinations/due - Vacunaciones próximas a vencer

**Propósito:** Obtener vacunaciones que vencen en los próximos 30 días.

**Comportamiento:**

1. **Filtro**
   - Retorna vacunaciones donde `next_due_date` está entre hoy y +30 días
   - Solo vacunaciones activas (no eliminadas)

2. **Datos retornados**
   - Lista de vacunaciones próximas a vencer

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/vaccinations/overdue - Vacunaciones vencidas

**Propósito:** Obtener vacunaciones que ya vencieron.

**Comportamiento:**

1. **Filtro**
   - Retorna vacunaciones donde `next_due_date` < hoy
   - Estado = `overdue`

2. **Datos retornados**
   - Lista de vacunaciones vencidas

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints de Catálogo de Vacunas

### POST /api/vaccines - Crear vacuna del catálogo

**Propósito:** Crear una nueva vacuna en el catálogo del tenant.

**Comportamiento:**

1. **Validaciones de campos requeridos**
   - `name` - Requerido (2-100 caracteres)
   - `dose_number` - Requerido (first, second, booster)
   - `validity_months` - Requerido, >= 1
   - `target_species` - Requerido (lista de especies)

2. **Validaciones de negocio**
   - Verifica que el nombre sea único por tenant
   - Valida que `dose_number` sea válido

3. **Creación**
   - Genera un ID único para la vacuna
   - Establece `active = true` por defecto

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/vaccines - Listar vacunas del catálogo

**Propósito:** Obtener todas las vacunas del catálogo.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `dose_number` - Filtrar por tipo de dosis
   - `target_species` - Filtrar por especie objetivo
   - `active` - Filtrar por estado activo
   - `search` - Buscar por nombre o fabricante

2. **Datos retornados**
   - Lista de vacunas con toda su información

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/vaccines/:id - Actualizar vacuna

**Propósito:** Actualizar una vacuna del catálogo.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que el nombre sea único (si se modifica)

2. **Actualización**
   - Actualiza solo los campos enviados

3. **Notificaciones**
   - No envía notificaciones

---

### DELETE /api/vaccines/:id - Eliminar vacuna

**Propósito:** Eliminar (soft delete) una vacuna del catálogo.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido

2. **Eliminación**
   - Realiza un soft delete

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints Mobile (dueños de mascotas)

### GET /mobile/vaccinations/patient/:patient_id - Mis vacunaciones

**Propósito:** El cliente obtiene las vacunaciones de su mascota desde la app móvil.

**Comportamiento:**

1. **Validaciones de seguridad**
   - Verifica que el paciente pertenezca al owner autenticado
   - Si no pertenece, retorna error

2. **Datos retornados**
   - Historial de vacunaciones de la mascota

3. **Notificaciones**
   - No envía notificaciones

---

## Reglas de Negocio

### Estados de una vacunación
| Estado | Descripción |
|--------|-------------|
| `applied` | Vacuna aplicada, no vencida |
| `due` | Próxima a vencer (dentro de 30 días) |
| `overdue` | Vencida |

### Tipos de dosis
| Tipo | Descripción |
|------|-------------|
| `first` | Primera dosis |
| `second` | Segunda dosis |
| `booster` | Refuerzo |

### Validaciones de fechas
- **Application date**: No puede ser en el futuro
- **Next due date**: Debe ser posterior a application date

### Certificado de vacunación
- Formato: `VAC-{tenant_id}-{YYYYMMDD}`
- Ejemplo: `VAC-507f-20260224`
- Se genera automáticamente al crear la vacunación

### Cálculo automático de próxima fecha
Si no se proporciona `next_due_date`, se puede calcular automáticamente basado en:
- `validity_months` del catálogo de vacunas
- `application_date` + `validity_months`

---

## Notificaciones Resumen

| Evento | Notificación | Destinatario | Trigger |
|--------|--------------|--------------|---------|
| Vacunación registrada | "Vacunación Registrada" | Owner | Al crear vacunación |
| Recordatorio 7 días | "Vacuna Vence en 7 días" | Owner | Scheduler diario |
| Recordatorio 24h | "Vacuna Vence Mañana" | Owner | Scheduler diario |
| Vacuna vencida | "Vacuna Vencida" | Owner | Scheduler diario |

---

## Scheduler (Tareas Automáticas)

El sistema ejecuta las siguientes tareas automáticas:

### Diariamente (cada 24 horas a las 8:00 AM)
```
1. Buscar vacunaciones con next_due_date = hoy + 7 días
   → Enviar recordatorio 7 días

2. Buscar vacunaciones con next_due_date = mañana
   → Enviar recordatorio 24 horas

3. Buscar vacunaciones con next_due_date < hoy
   → Marcar como overdue (si no lo están)
   → Enviar notificación de vacuna vencida
```

---

## Errores Específicos del Módulo

| Código | HTTP | Descripción |
|--------|------|-------------|
| `VACCINATION_NOT_FOUND` | 404 | La vacunación no existe |
| `VACCINE_NOT_FOUND` | 404 | La vacuna no existe |
| `VACCINE_NAME_EXISTS` | 409 | El nombre de vacuna ya existe |
| `PATIENT_NOT_FOUND` | 404 | El paciente no existe |
| `PATIENT_INACTIVE` | 400 | El paciente está inactivo |
| `VETERINARIAN_NOT_FOUND` | 404 | El veterinario no existe |
| `INVALID_APPLICATION_DATE` | 400 | La fecha de aplicación no puede ser en el futuro |
| `INVALID_NEXT_DUE_DATE` | 400 | La fecha próxima debe ser después de la aplicación |
| `INVALID_STATUS` | 400 | Estado de vacunación inválido |
| `INVALID_DOSE_TYPE` | 400 | Tipo de dosis inválido |
| `SPECIES_MISMATCH` | 400 | La vacuna no es para esta especie |

---

## Integración con Otros Módulos

### Medical Records
- Las vacunaciones pueden registrarse como tipo de registro médico
- Relación mediante `patient_id`

### Appointments
- Recordatorios de vacunación pueden generar citas automáticas
- Relación mediante `patient_id` y `owner_id`

### Inventory
- Las vacunas aplicadas pueden descontar stock del inventario
- Relación mediante `vaccine_name` → `products`

### Notifications
- Notificaciones push/email para recordatorios
- Notificaciones al owner y al staff

---

## Ejemplos de Uso

### Registrar vacunación
```json
POST /api/vaccinations
{
  "patient_id": "507f1f77bcf86cd799439011",
  "veterinarian_id": "507f1f77bcf86cd799439022",
  "vaccine_name": "Rabia",
  "manufacturer": "Intervet",
  "lot_number": "RAB-2026-001",
  "application_date": "2026-02-24T10:00:00Z",
  "next_due_date": "2027-02-24T10:00:00Z",
  "notes": "Primera vacuna de rabia"
}
```

### Crear vacuna del catálogo
```json
POST /api/vaccines
{
  "name": "Rabia",
  "description": "Vacuna contra la rabia",
  "manufacturer": "Intervet",
  "dose_number": "first",
  "validity_months": 12,
  "target_species": ["dog", "cat"],
  "active": true
}
```

### Obtener vacunaciones próximas a vencer
```
GET /api/vaccinations/due

Response:
{
  "data": [
    {
      "id": "...",
      "patient_id": "...",
      "vaccine_name": "Rabia",
      "application_date": "2025-02-24T10:00:00Z",
      "next_due_date": "2026-02-24T10:00:00Z",
      "status": "due",
      "certificate_number": "VAC-507f-20250224"
    }
  ]
}
```

### Obtener historial por paciente
```
GET /api/vaccinations/patient/507f1f77bcf86cd799439011?page=1&limit=20
```

---

## Consideraciones de Seguridad

### Multi-tenancy
- Todas las vacunaciones están aisladas por `tenant_id`
- El catálogo de vacunas es por tenant

### RBAC (Roles y Permisos)
| Rol | Permisos vaccinations | Permisos vaccines |
|-----|----------------------|-------------------|
| `admin` | CRUD | CRUD |
| `veterinarian` | CRUD | CRUD |
| `receptionist` | Leer, Crear | Leer |
| `assistant` | Leer | Leer |
| `accountant` | Sin acceso | Sin acceso |

### Mobile (Owners)
- Los owners solo pueden VER las vacunaciones de SUS mascotas
- No pueden crear, editar ni eliminar vacunaciones
- El acceso se valida mediante `OwnerGuardMiddleware`

### Auditoría
- Todas las vacunaciones tienen `created_at` y `updated_at`
- Soft delete mantiene los datos para auditoría
- Número de certificado único para trazabilidad
