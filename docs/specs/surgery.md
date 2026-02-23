# Spec: surgery

Scope: feature

# Surgery Module Specification

## Overview
Programación, seguimiento y documentación de cirugías.

## Domain Entities

### Surgery (schema.go)
- ID, TenantID, PatientID, OwnerID
- AppointmentID (optional link)
- ScheduledDate, StartTime, EndTime
- SurgeonID, AnesthetistID
- Procedure (type of surgery)
- Diagnosis
- PreOpNotes, PostOpNotes
- AnesthesiaType
- Status (scheduled, in_progress, completed, cancelled)
- Complications (if any)
- PostOpInstructions
- FollowUpDate
- CreatedAt, UpdatedAt

### SurgeryType (schema.go) - Catalog
- ID, TenantID, Name
- Description
- EstimatedDuration
- Price
- RequiresFasting
- RequiresAnesthesia
- PreOpTests (required lab tests)
- Active

---

## Business Logic

### Validaciones

#### Al programar cirugía (CreateSurgeryDTO)
| Campo | Regla |
|-------|-------|
| PatientID | Requerido, debe existir y estar activo |
| SurgeonID | Requerido, debe existir |
| ScheduledDate | Requerido, no puede ser en pasado |
| StartTime | Requerido |
| Procedure | Requerido |
| Diagnosis | Opcional |
| AnesthesiaType | Opcional |

#### Al crear tipo de cirugía (CreateSurgeryTypeDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido, único por tenant |
| Description | Opcional |
| EstimatedDuration | Requerido, > 0 (minutos) |
| Price | Requerido, >= 0 |
| RequiresFasting | Default false |
| RequiresAnesthesia | Default true |

### Reglas de Negocio

1. **Horario válido**: Solo entre 8am-6pm, no domingos
2. **Cirujano disponible**: No debe tener otra cirugía en el mismo horario
3. **Duración**: No puede exceder las 8 horas
4. **Requiere ayuno**: Si RequiresFasting, verificar que el paciente haya ayunado
5. **Requiere anesthesia**: Si RequiresAnesthesia, debe haber anesthetist
6. **Transiciones**: scheduled → in_progress → completed, o cualquier → cancelled
7. **Fecha de seguimiento**: Si se indica, crear recordatorio

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `surgery.created` | Cirugía programada | Ninguna |
| `surgery.status.changed` | Cambio de estado | Ninguna |
| `surgery.completed` | Cirugía completada | Ninguna |
| `surgery.follow_up` | FollowUpDate establecida | Crear cita de seguimiento |
| `surgery.complications` | Complicaciones registradas | Alertar al staff |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `surgery.scheduled` | Owner | Cirugía programada |
| `surgery.reminder` | Owner | 24h antes |
| `surgery.completed` | Owner | Completed + instrucciones post-op |
| `surgery.cancelled` | Owner | Cancelada |

### Transiciones de Estado

```
scheduled (programada)
    ↓ (inicia cirugía)
in_progress (en curso)
    ↓ (termina cirugía)
completed (completada)
```

```
scheduled (cualquier estado)
    ↓ (se cancela)
cancelled (cancelada)
```

---

## Features

### Phase 1: Basic
- [ ] Schedule surgeries
- [ ] Track status
- [ ] Link to patient/appointment

### Phase 2: Documentation
- [ ] Pre/post op notes
- [ ] Complications tracking
- [ ] Follow-up scheduling

### Phase 3: Advanced
- [ ] Surgery type catalog
- [ ] Pricing
- [ ] Reports (complications, volume)

---

## Integration Points

- **appointments**: Link to appointment, crear cita de seguimiento
- **patients**: Patient history, verificar especie
- **users**: Surgeon/anesthetist attribution
- **laboratory**: PreOpTests pueden requerir lab orders

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /surgeries | Programar cirugía |
| GET | /surgeries | Listar cirugías |
| GET | /surgeries/:id | Obtener |
| PUT | /surgeries/:id | Actualizar |
| DELETE | /surgeries/:id | Cancelar |
| PATCH | /surgeries/:id/status | Cambiar estado |
| GET | /surgeries/calendar | Vista calendario |
| GET | /surgeries/:id/timeline | Historial |
| POST | /surgery-types | Crear tipo |
| GET | /surgery-types | Listar tipos |
| GET | /surgery-types/:id | Obtener tipo |
| PUT | /surgery-types/:id | Actualizar tipo |
| DELETE | /surgery-types/:id | Eliminar tipo |

---

## DTOs (Data Transfer Objects)

### CreateSurgeryDTO
```go
type CreateSurgeryDTO struct {
    PatientID       string `json:"patient_id" binding:"required"`
    AppointmentID   string `json:"appointment_id"`
    SurgeonID       string `json:"surgeon_id" binding:"required"`
    AnesthetistID   string `json:"anesthetist_id"`
    ScheduledDate   string `json:"scheduled_date" binding:"required"` // RFC3339
    StartTime       string `json:"start_time" binding:"required"` // HH:mm
    Procedure       string `json:"procedure" binding:"required"`
    Diagnosis       string `json:"diagnosis"`
    AnesthesiaType  string `json:"anesthesia_type"`
    PreOpNotes      string `json:"pre_op_notes"`
}
```

### UpdateSurgeryStatusDTO
```go
type UpdateSurgeryStatusDTO struct {
    Status          string `json:"status" binding:"required,oneof=scheduled in_progress completed cancelled"`
    PostOpNotes     string `json:"post_op_notes"`
    Complications  string `json:"complications"`
    EndTime         string `json:"end_time"` // HH:mm
    FollowUpDate    string `json:"follow_up_date"` // RFC3339
    PostOpInstructions string `json:"post_op_instructions"`
}
```

### CreateSurgeryTypeDTO
```go
type CreateSurgeryTypeDTO struct {
    Name                string   `json:"name" binding:"required"`
    Description         string   `json:"description"`
    EstimatedDuration   int      `json:"estimated_duration" binding:"required,min=1"` // minutes
    Price               float64  `json:"price" binding:"required,min=0"`
    RequiresFasting     bool     `json:"requires_fasting"`
    RequiresAnesthesia  bool     `json:"requires_anesthesia"`
    PreOpTests          []string `json:"pre_op_tests"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| PATIENT_NOT_FOUND | El paciente no existe |
| PATIENT_INACTIVE | El paciente está inactivo |
| SURGEON_NOT_FOUND | El cirujano no existe |
| ANESTHETIST_REQUIRED | Se requiere anesthetist para esta cirugía |
| SCHEDULED_TIME_INVALID | Hora fuera del horario válido (8am-6pm) |
| SUNDAY_NOT_ALLOWED | No se programan cirugías los domingos |
| SURGEON_NOT_AVAILABLE | El cirujano tiene otra cirugía en ese horario |
| SURGERY_ALREADY_STARTED | La cirugía ya started, no se puede modificar |
| DURATION_EXCEEDED | La duración excede el máximo permitido |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice para ScheduledDate + Status (para calendario)
- Crear índice para SurgeonID + ScheduledDate (para disponibilidad)
- Validar que StartTime + EstimatedDuration no exceda 6pm
- Si FollowUpDate establecida, crear automáticamente cita de seguimiento
- Si RequiresFasting, guardar en PreOpNotes recordatorio
