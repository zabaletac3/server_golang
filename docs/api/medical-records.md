# API de Historias Clínicas (Medical Records)

Documentación de comportamiento de la API de historias clínicas. Aquí se describe qué sucede en el sistema cuando realizas cada acción, incluyendo validaciones y notificaciones.

---

## Endpoints de Admin

### POST /api/medical-records - Crear registro médico

**Propósito:** Crear un nuevo registro clínico para un paciente desde el panel de administración.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `patient_id` sea un ID válido de MongoDB
   - Valida que `veterinarian_id` sea un ID válido de MongoDB
   - Valida que `type` sea uno de: consultation, emergency, surgery, checkup, vaccination

2. **Validaciones de campos requeridos**
   - `patient_id` - Requerido
   - `veterinarian_id` - Requerido
   - `type` - Requerido
   - `chief_complaint` - Requerido (1-500 caracteres)

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - Verifica que el paciente esté activo
   - Verifica que el veterinario exista
   - Valida que `temperature` esté entre 30-45°C (si se proporciona)
   - Valida que `weight` sea >= 0 (si se proporciona)
   - Valida que `next_visit_date` no sea en el pasado (si se proporciona)

4. **Creación**
   - Genera un ID único para el registro
   - Asigna automáticamente el `owner_id` del paciente
   - Establece `created_at` y `updated_at`
   - Convierte los medicamentos del formato DTO al formato de base de datos

5. **Notificaciones enviadas**
   - ✅ **Owner (push):** "Nuevo registro médico" - Notifica que se creó un registro para su mascota
   - ✅ **Owner (push):** "Próxima visita programada" - Si se indicó `next_visit_date`

---

### GET /api/medical-records - Listar registros médicos

**Propósito:** Obtener una lista paginada de registros clínicos.

**Comportamiento:**

1. **Parámetros de filtros (opcionales)**
   - `patient_id` - Filtrar por paciente
   - `veterinarian_id` - Filtrar por veterinario
   - `type` - Filtrar por tipo (consultation, emergency, surgery, checkup, vaccination)
   - `date_from` - Fecha inicio (formato RFC3339)
   - `date_to` - Fecha fin (formato RFC3339)
   - `has_attachments` - Filtrar registros con archivos adjuntos (true/false)

2. **Paginación**
   - Parámetros `page` y `limit` en query string
   - Retorna metadatos de paginación (total, página actual, total_pages)

3. **Ordenamiento**
   - Por defecto ordena por `created_at` descendente (más recientes primero)

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/medical-records/:id - Obtener registro médico

**Propósito:** Obtener los detalles de un registro clínico específico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea un formato válido de MongoDB

2. **Datos retornados**
   - Retorna el registro con todos sus campos
   - Incluye medicamentos, adjuntos, y fechas de seguimiento

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/medical-records/:id - Actualizar registro médico

**Propósito:** Actualizar los datos de un registro clínico existente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que el ID del registro sea válido

2. **Validación de tiempo (REGLA CRÍTICA)**
   - **Solo se puede editar dentro de las 24 horas posteriores a la creación**
   - Después de 24 horas, el registro se considera inmutable (compliance legal)
   - Retorna error `RECORD_NOT_EDITABLE` si pasó el tiempo límite

3. **Validaciones de campos**
   - Valida que `temperature` esté entre 30-45°C (si se modifica)
   - Valida que `weight` sea >= 0 (si se modifica)
   - Valida que `next_visit_date` no sea en el pasado (si se modifica)

4. **Actualización**
   - Actualiza solo los campos enviados (partial update)
   - Actualiza el campo `updated_at`

5. **Notificaciones**
   - No envía notificaciones automáticamente

---

### DELETE /api/medical-records/:id - Eliminar registro médico

**Propósito:** Eliminar (soft delete) un registro clínico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que el registro exista

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)
   - No elimina físicamente el documento
   - Mantiene integridad referencial para auditoría

3. **Notificaciones**
   - No envía notificaciones

---

### GET /api/medical-records/patient/:patient_id - Historial por paciente

**Propósito:** Obtener todos los registros clínicos de un paciente específico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el `patient_id` sea un formato válido de MongoDB

2. **Paginación**
   - Soporta parámetros `page` y `limit`
   - Ordena por fecha de creación (más reciente primero)

3. **Datos retornados**
   - Lista de registros clínicos del paciente
   - Incluye metadatos de paginación

4. **Notificaciones**
   - No envía notificaciones

---

### GET /api/medical-records/patient/:patient_id/timeline - Timeline cronológico

**Propósito:** Obtener una vista cronológica del historial médico de un paciente.

**Comportamiento:**

1. **Parámetros requeridos**
   - `patient_id` - ID del paciente en la ruta

2. **Parámetros opcionales**
   - `date_from` - Filtrar desde fecha (RFC3339)
   - `date_to` - Filtrar hasta fecha (RFC3339)
   - `record_type` - Filtrar por tipo de registro
   - `limit` - Número de registros (default: 50)
   - `skip` - Número de registros a saltar (default: 0)

3. **Datos retornados**
   - `patient_id` - ID del paciente
   - `patient_name` - Nombre del paciente
   - `entries` - Lista cronológica con: fecha, tipo, descripción, veterinario, record_id
   - `total_count` - Total de registros

4. **Notificaciones**
   - No envía notificaciones

---

## Endpoints de Alergias

### POST /api/allergies - Registrar alergia

**Propósito:** Registrar una nueva alergia para un paciente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `patient_id` sea un ID válido de MongoDB

2. **Validaciones de campos requeridos**
   - `patient_id` - Requerido
   - `allergen` - Requerido (1-100 caracteres)
   - `severity` - Requerido, debe ser: mild, moderate, severe

3. **Validaciones de negocio**
   - Verifica que el paciente exista
   - Valida que `severity` sea válido

4. **Creación**
   - Genera un ID único para la alergia
   - Establece `created_at` y `updated_at`

5. **Notificaciones enviadas**
   - ✅ **Staff (alerta):** "Alerta: Alergia Severa Registrada" - Solo si `severity = severe`
   - La alerta incluye: nombre del paciente, alérgeno, severidad

---

### GET /api/allergies/patient/:patient_id - Alergias del paciente

**Propósito:** Obtener todas las alergias registradas de un paciente.

**Comportamiento:**

1. **Validaciones**
   - Valida que el `patient_id` sea un formato válido de MongoDB

2. **Datos retornados**
   - Lista de alergias con: alérgeno, severidad, descripción, fechas

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/allergies/:id - Actualizar alergia

**Propósito:** Actualizar los datos de una alergia existente.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID de la alergia sea válido
   - Verifica que la alergia exista

2. **Actualización**
   - Actualiza solo los campos enviados
   - Actualiza el campo `updated_at`

3. **Notificaciones**
   - No envía notificaciones automáticamente

---

### DELETE /api/allergies/:id - Eliminar alergia

**Propósito:** Eliminar (soft delete) un registro de alergia.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que la alergia exista

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints de Historial Médico (Resumen)

### POST /api/medical-history - Crear/Actualizar resumen médico

**Propósito:** Crear o actualizar el resumen del historial médico de un paciente.

**Comportamiento:**

1. **Validaciones de formato**
   - Valida que `patient_id` sea un ID válido de MongoDB

2. **Validaciones de campos**
   - `patient_id` - Requerido
   - `blood_type` - Opcional, máximo 10 caracteres

3. **Comportamiento especial**
   - **Si ya existe un resumen:** Lo actualiza (upsert)
   - **Si no existe:** Crea uno nuevo
   - Solo puede haber UN resumen por paciente

4. **Datos del resumen**
   - `chronic_conditions` - Lista de condiciones crónicas
   - `previous_surgeries` - Lista de cirugías previas
   - `risk_factors` - Lista de factores de riesgo
   - `blood_type` - Tipo de sangre

5. **Notificaciones**
   - No envía notificaciones

---

### GET /api/medical-history/patient/:patient_id - Obtener resumen médico

**Propósito:** Obtener el resumen del historial médico de un paciente.

**Comportamiento:**

1. **Validaciones**
   - Valida que el `patient_id` sea un formato válido de MongoDB

2. **Datos retornados**
   - Resumen con: condiciones crónicas, cirugías previas, factores de riesgo, tipo de sangre

3. **Notificaciones**
   - No envía notificaciones

---

### PUT /api/medical-history/:id - Actualizar resumen médico

**Propósito:** Actualizar el resumen del historial médico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID del historial sea válido
   - Verifica que el historial exista

2. **Actualización**
   - Actualiza solo los campos enviados
   - Actualiza el campo `updated_at`

3. **Notificaciones**
   - No envía notificaciones

---

### DELETE /api/medical-history/:id - Eliminar resumen médico

**Propósito:** Eliminar (soft delete) un resumen de historial médico.

**Comportamiento:**

1. **Validaciones**
   - Valida que el ID sea válido
   - Verifica que el historial exista

2. **Eliminación**
   - Realiza un soft delete (marca campo `deleted_at`)

3. **Notificaciones**
   - No envía notificaciones

---

## Endpoints Mobile (dueños de mascotas)

### GET /mobile/medical-records/patient/:patient_id - Mis registros

**Propósito:** El cliente obtiene los registros clínicos de su mascota desde la app móvil.

**Comportamiento:**

1. **Validaciones de seguridad**
   - Verifica que el paciente pertenezca al owner autenticado
   - Si no pertenece, retorna error

2. **Filtros**
   - Muestra solo los registros de las mascotas del owner
   - Soporta paginación

3. **Notificaciones**
   - No envía notificaciones

---

### GET /mobile/medical-records/patient/:patient_id/timeline - Timeline de mi mascota

**Propósito:** El cliente ve el timeline cronológico de su mascota.

**Comportamiento:**

1. **Validaciones de seguridad**
   - Verifica que el paciente pertenezca al owner autenticado

2. **Datos retornados**
   - Timeline cronológico con todos los eventos médicos

3. **Notificaciones**
   - No envía notificaciones

---

### GET /mobile/allergies/patient/:patient_id - Alergias de mi mascota

**Propósito:** El cliente ve las alergias registradas de su mascota.

**Comportamiento:**

1. **Validaciones de seguridad**
   - Verifica que el paciente pertenezca al owner autenticado

2. **Datos retornados**
   - Lista de alergias con alérgeno y severidad

3. **Notificaciones**
   - No envía notificaciones

---

### GET /mobile/medical-history/patient/:patient_id - Historial de mi mascota

**Propósito:** El cliente ve el resumen del historial médico de su mascota.

**Comportamiento:**

1. **Validaciones de seguridad**
   - Verifica que el paciente pertenezca al owner autenticado

2. **Datos retornados**
   - Resumen con condiciones crónicas, cirugías previas, factores de riesgo

3. **Notificaciones**
   - No envía notificaciones

---

## Reglas de Negocio

### Regla de las 24 horas (COMPLIANCE)
- **Los registros médicos NO se pueden editar después de 24 horas**
- Esto es para cumplir con regulaciones de integridad de historiales clínicos
- Si se necesita corregir un registro antiguo, se debe crear un nuevo registro con la corrección

### Tipos de registro médico
| Tipo | Descripción |
|------|-------------|
| `consultation` | Consulta general |
| `emergency` | Emergencia |
| `surgery` | Cirugía |
| `checkup` | Chequeo rutinario |
| `vaccination` | Vacunación |

### Severidad de alergias
| Severidad | Descripción | Acción |
|-----------|-------------|--------|
| `mild` | Leve, molestias menores | Registro normal |
| `moderate` | Moderada, requiere atención | Registro normal |
| `severe` | Severa, riesgo de vida | **Alerta automática al staff** |

### Validaciones de signos vitales
| Campo | Rango válido |
|-------|--------------|
| `temperature` | 30 - 45 °C |
| `weight` | >= 0 (cualquier valor positivo) |

### Estados de un registro médico
| Estado | Descripción |
|--------|-------------|
| `editable` | Menos de 24 horas desde creación |
| `locked` | Más de 24 horas, solo lectura |

### Adjuntos
- Se almacenan como IDs de referencia al módulo `resources`
- Tipos soportados: PDF, imágenes (JPG, PNG)
- Múltiples adjuntos por registro

---

## Notificaciones Resumen

| Acción | Notificación | Destinatario | Canal |
|--------|--------------|--------------|-------|
| Crear registro médico | "Nuevo registro médico" | Owner | Push |
| Crear registro con próxima visita | "Próxima visita programada" | Owner | Push |
| Registrar alergia severa | "Alerta: Alergia Severa Registrada" | Staff (todos) | Sistema |

---

## Errores Específicos del Módulo

| Código | HTTP | Descripción |
|--------|------|-------------|
| `PATIENT_NOT_FOUND` | 404 | El paciente no existe |
| `PATIENT_INACTIVE` | 400 | El paciente está inactivo |
| `VETERINARIAN_NOT_FOUND` | 404 | El veterinario no existe |
| `RECORD_NOT_EDITABLE` | 400 | El registro no se puede editar después de 24h |
| `INVALID_TEMPERATURE` | 400 | Temperatura fuera de rango (30-45°C) |
| `INVALID_WEIGHT` | 400 | Peso inválido (debe ser >= 0) |
| `INVALID_NEXT_VISIT_DATE` | 400 | Fecha de próxima visita no puede ser en el pasado |
| `INVALID_MEDICAL_RECORD_TYPE` | 400 | Tipo de registro inválido |
| `INVALID_ALLERGY_SEVERITY` | 400 | Severidad de alergia inválida |
| `DUPLICATE_HISTORY` | 409 | Ya existe un resumen médico para este paciente |

---

## Consideraciones de Seguridad

### Multi-tenancy
- Todos los registros están aislados por `tenant_id`
- Un tenant no puede ver registros de otro tenant

### RBAC (Roles y Permisos)
| Rol | Permisos medical-records |
|-----|-------------------------|
| `admin` | get, post, put, patch, delete |
| `veterinarian` | get, post, put, patch, delete |
| `assistant` | get (solo lectura) |
| `receptionist` | Sin acceso |
| `accountant` | Sin acceso |

### Mobile (Owners)
- Los owners solo pueden VER los registros de SUS mascotas
- No pueden crear, editar ni eliminar registros
- El acceso se valida mediante `OwnerGuardMiddleware`

### Auditoría
- Todos los registros tienen `created_at` y `updated_at`
- Soft delete mantiene los datos para auditoría
- El campo `deleted_at` indica cuándo se eliminó

---

## Ejemplos de Uso

### Crear registro de consulta
```json
POST /api/medical-records
{
  "patient_id": "507f1f77bcf86cd799439011",
  "veterinarian_id": "507f1f77bcf86cd799439022",
  "type": "consultation",
  "chief_complaint": "Vómitos y diarrea desde ayer",
  "diagnosis": "Gastroenteritis aguda",
  "symptoms": "Vómitos, diarrea, deshidratación leve",
  "weight": 15.5,
  "temperature": 39.2,
  "treatment": "Suero oral, ayuno 12h, dieta blanda",
  "medications": [
    {
      "name": "Ondansetrón",
      "dose": "4mg",
      "frequency": "Cada 8 horas",
      "duration": "3 días"
    }
  ],
  "next_visit_date": "2026-03-01T10:00:00Z"
}
```

### Registrar alergia severa
```json
POST /api/allergies
{
  "patient_id": "507f1f77bcf86cd799439011",
  "allergen": "Penicilina",
  "severity": "severe",
  "description": "Reacción anafiláctica documentada"
}
```

### Obtener timeline
```
GET /api/medical-records/patient/507f1f77bcf86cd799439011/timeline?limit=20
```
