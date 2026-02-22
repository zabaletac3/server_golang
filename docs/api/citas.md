# API de Citas Veterinarias

Documentación de comportamiento de la API de citas. Aquí se describe qué sucede en el sistema cuando realizas cada acción, incluyendo validaciones y notificaciones.

---

## Endpoints de Admin

### POST /api/appointments - Crear cita

**Propósito:** Crear una nueva cita desde el panel de administración.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `patient_id` sea un ID válido de MongoDB
   - Valida que `veterinarian_id` sea un ID válido de MongoDB

2. **Validaciones de horario**
   - La fecha no puede ser en el pasado
   - El horario debe estar entre las 8:00 AM y las 5:59 PM (antes de las 6:00 PM)
   - No se permiten citas los domingos

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - Verifica que el veterinario exista
   - Verifica que no haya conflicto de horario con otra cita del mismo veterinario
   - Si no se especifica prioridad, usa "normal" por defecto

4. **Creación**
   - Genera un ID único para la cita
   - Asigna el owner del paciente automáticamente
   - Define estado inicial: `scheduled`

5. **Notificaciones enviadas**
   - ✅ **Owner (correo + push):** "Nueva cita agendada" - Notifica la fecha y hora de la cita
   - ✅ **Veterinario (correo):** "Nueva cita asignada" - Notifica el paciente y horario

---

### GET /api/appointments - Listar citas

**Propósito:** Obtener una lista paginada de citas.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `status` - Filtrar por estado (scheduled, confirmed, in_progress, completed, cancelled, no_show)
   - `type` - Filtrar por tipo (consultation, grooming, surgery, emergency, vaccination, checkup)
   - `veterinarian_id` - Filtrar por veterinario
   - `patient_id` - Filtrar por paciente
   - `owner_id` - Filtrar por owner
   - `date_from` - Fecha inicio (formato RFC3339)
   - `date_to` - Fecha fin (formato RFC3339)
   - `priority` - Filtrar por prioridad (low, normal, high, urgent)

2. **Paginación**
   - Parámetros `page` y `limit` en query string
   - Retorna metadatos de paginación (total, página actual, etc.)

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/appointments/:id - Obtener cita

**Propósito:** Obtener los detalles de una cita específica.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea un formato válido de MongoDB

2. **Datos retornados**
   - Retorna la cita con todos sus campos
   - Por defecto no incluye datos relacionados (paciente, owner, veterinario)
   - Se puede solicitar populate para incluir datos relacionados

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/appointments/:id - Actualizar cita

**Propósito:** Actualizar los datos de una cita existente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID de la cita sea válido

2. **Validaciones de horario** (si se modifica ScheduledAt)
   - No puede ser en el pasado
   - Debe estar entre 8am-6pm
   - No puede ser domingo

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - Verifica que el veterinario exista
   - Verifica conflictos de horario (excluyendo la cita actual)

4. **Actualización**
   - Actualiza solo los campos enviados (partial update)
   - Actualiza el campo `updated_at`

5. **Notificaciones**
   - No envía notificaciones automáticamente

---

### DELETE /api/appointments/:id - Eliminar cita

**Propósito:** Eliminar (soft delete) una cita.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que la cita exista

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)
   - No elimina físicamente el documento

3. **Notificaciones**
   - No envía notificaciones

---

### PATCH /api/appointments/:id/status - Cambiar estado

**Propósito:** Cambiar el estado de una cita (confirmar, iniciar, completar, cancelar, etc.).

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Valida que la transición de estado sea válida

2. **Transiciones permitidas**
   - `scheduled` → `confirmed`
   - `confirmed` → `in_progress`
   - `in_progress` → `completed`
   - `scheduled` / `confirmed` → `cancelled`
   - `scheduled` → `no_show`
   - Estados terminales (`completed`, `cancelled`, `no_show`) no pueden cambiar

3. **Acciones por estado**
   - **Confirmed:** Registra `confirmed_at`
   - **In Progress:** Registra `started_at`
   - **Completed:** Registra `completed_at`
   - **Cancelled:** Registra `cancelled_at` y `cancel_reason`
   - **No Show:** Registra `cancelled_at`

4. **Historial**
   - Crea un registro en el historial de estados

5. **Notificaciones enviadas**
   - ✅ **Confirmed → Owner:** "Tu cita ha sido confirmada"
   - ✅ **Cancelled → Owner:** "Tu cita ha sido cancelada" + razón

---

### GET /api/appointments/calendar - Vista de calendario

**Propósito:** Obtener citas en un rango de fechas para mostrar en un calendario.

**Comportamiento:**

1. **Parámetros requeridos**
   - `date_from` - Fecha inicio (RFC3339)
   - `date_to` - Fecha fin (RFC3339)

2. **Parámetros opcionales**
   - `veterinarian_id` - Filtrar por veterinario

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/appointments/availability - Verificar disponibilidad

**Propósito:** Verificar si un veterinario tiene disponibilidad en un horario específico.

**Comportamiento:**

1. **Parámetros requeridos**
   - `veterinarian_id` - ID del veterinario
   - `scheduled_at` - Fecha y hora propuesta (RFC3339)
   - `duration` - Duración en minutos

2. **Validaciones de horario**
   - Valida horario de negocio (8am-6pm)
   - Valida que no sea domingo

3. **Retorno**
   - `{ "available": true }` - Si hay espacio
   - `{ "available": false, "conflicts": [...] }` - Si hay conflicto

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/appointments/:id/history - Historial de estados

**Propósito:** Obtener el historial de cambios de estado de una cita.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido

2. **Retorno**
   - Lista de transiciones con: estado anterior, nuevo estado, motivo, quién realizó el cambio, fecha

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints Mobile (dueños de mascotas)

### POST /api/mobile/appointments/request - Solicitar cita

**Propósito:** Un cliente desde la app móvil solicita una nueva cita.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que patient_id sea válido

2. **Validaciones de horario**
   - No puede ser en el pasado
   - Debe estar entre 8am-6pm
   - No puede ser domingo

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - **Verifica que el paciente pertenezca al owner autenticado** (seguridad)
   - Si no se especifica prioridad, usa "normal"

4. **Creación**
   - Asigna `veterinarian_id` como nil (sin asignar) - el staff elige después
   - Define duración default de 30 minutos
   - Estado inicial: `scheduled`

5. **Notificaciones enviadas**
   - ✅ **Staff (correo):** "Nueva solicitud de cita" - Notifica el nombre del cliente y paciente

---

### GET /api/mobile/appointments - Mis citas

**Propósito:** El cliente obtiene sus citas desde la app móvil.

**Comportamiento:**

1. **Filtros**
   - Muestra solo las citas del owner autenticado
   - Soporta paginación

2. **Notificaciones**
   - No envía notificaciones

---

### GET /api/mobile/appointments/:id - Detalle de mi cita

**Propósito:** El cliente ve el detalle de una cita específica.

**Comportamiento:**

1. **Validaciones**
   - Verifica que la cita pertenezca al owner autenticado (seguridad)
   - Si no pertenece, retorna error

2. **Notificaciones**
   - No envía notificaciones

---

### PATCH /api/mobile/appointments/:id/cancel - Cancelar mi cita

**Propósito:** El cliente cancela una cita desde la app móvil.

**Comportamiento:**

1. **Validaciones**
   - Verifica que la cita pertenezca al owner autenticado
   - Verifica que la transición sea válida (no se puede cancelar si ya está completada/cancelada)

2. **Cancelación**
   - Actualiza estado a `cancelled`
   - Registra `cancelled_at` y `cancel_reason`

3. **Notificaciones enviadas**
   - ✅ **Staff (correo):** "Cita cancelada por cliente" - Notifica la razón

---

## Reglas de Negocio

### Horario de atención
- **Lunes a sábado:** 8:00 AM - 6:00 PM
- **Domingo:** Cerrado (no se permiten citas)

### Duración de citas
- Mínimo: 15 minutos
- Máximo: 480 minutos (8 horas)
- Default (mobile): 30 minutos

### Estados de una cita
| Estado | Descripción |
|--------|-------------|
| `scheduled` | Programada, pendiente de confirmación |
| `confirmed` | Confirmada por el staff |
| `in_progress` | En curso |
| `completed` | Finalizada |
| `cancelled` | Cancelada |
| `no_show` | El cliente no asistió |

### Tipos de cita
| Tipo | Descripción |
|------|-------------|
| `consultation` | Consulta general |
| `grooming` | Estética |
| `surgery` | Cirugía |
| `emergency` | Emergencia |
| `vaccination` | Vacunación |
| `checkup` | Revisión |

### Prioridades
| Prioridad | Descripción |
|----------|-------------|
| `low` | Baja |
| `normal` | Normal (default) |
| `high` | Alta |
| `urgent` | Urgente |

---

## Notificaciones Resumen

| Acción | Notificación | Destinatario |
|--------|--------------|--------------|
| Crear cita (Admin) | "Nueva cita agendada" | Owner (correo + push) |
| Crear cita (Admin) | "Nueva cita asignada" | Veterinario (correo) |
| Confirmar cita | "Cita confirmada" | Owner (correo + push) |
| Cancelar cita (Admin) | "Cita cancelada" | Owner (correo + push) |
| Solicitar cita (Mobile) | "Nueva solicitud de cita" | Staff (correo) |
| Cancelar cita (Mobile) | "Cita cancelada por cliente" | Staff (correo) |
