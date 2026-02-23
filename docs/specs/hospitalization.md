# Spec: hospitalization

Scope: feature

# Hospitalization Module Specification

## Overview
Gestión de pacientes internos: ocupación de jaulas/camas, seguimiento diario, altas.

## Domain Entities

### Hospitalization (schema.go)
- ID, TenantID, PatientID, OwnerID
- AdmissionDate, DischargeDate (optional)
- Reason
- InitialDiagnosis
- AssignedCage/Bed
- Status (admitted, in_progress, discharged)
- DailyNotes (array)
- AttachedTreatments
- ResponsibleVetID
- CreatedAt, UpdatedAt

### Cage (schema.go)
- ID, TenantID, Name/Number
- Type (dog, cat, isolation)
- Capacity
- IsAvailable
- Location

### DailyProgress (schema.go)
- ID, HospitalizationID
- Date
- Notes
- MedicationsGiven
- FoodProvided
- VetID

---

## Business Logic

### Validaciones

#### Al admitir paciente (CreateHospitalizationDTO)
| Campo | Regla |
|-------|-------|
| PatientID | Requerido, debe existir y estar activo |
| Reason | Requerido |
| InitialDiagnosis | Opcional |
| CageID | Requerido, jaula debe estar disponible |
| ResponsibleVetID | Requerido, debe existir |

#### Al crear jaula (CreateCageDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido, único por tenant |
| Type | Requerido (dog, cat, isolation) |
| Capacity | Requerido, >= 1 |
| Location | Opcional |

#### Al agregar nota diaria (CreateDailyProgressDTO)
| Campo | Regla |
|-------|-------|
| Notes | Opcional |
| MedicationsGiven | Opcional |
| FoodProvided | Opcional |

### Reglas de Negocio

1. **Jaula disponible**: Solo se puede admitir si hay jaula disponible del tipo correcto
2. **No duplicar admisión**: Un paciente no puede tener más de una admisión activa
3. **Alta con fecha**: Al dar de alta, guardar DischargeDate
4. **Solo un paciente por jaula**: Capacity = 1 por defecto
5. **Jaula isolation**: Para pacientes contagiosos
6. **Notas diarias**: Obligatorio al menos una nota por día
7. **Veterinario responsable**: Debe estar asignado todo el tiempo

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `hospitalization.admitted` | Paciente admitido | Ninguna |
| `hospitalization.discharged` | Paciente dado de alta | Notificar al owner |
| `hospitalization.daily_note` | Nota diaria agregada | Ninguna |
| `cage.occupied` | Jaula ocupada | Marcar como no disponible |
| `cage.available` | Jaula liberada | Marcar como disponible |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `hospitalization.admitted` | Owner (opcional) | Paciente admitido |
| `hospitalization.discharged` | Owner (push + email) | Paciente dado de alta |

### Transiciones de Estado

```
admitted (paciente admitido)
    ↓ (paciente en observación)
in_progress (en curso)
    ↓ (alta médica)
discharged (alta)
```

---

## Features

### Phase 1: Basic
- [ ] Admit/discharge patients
- [ ] Assign cage
- [ ] Track status

### Phase 2: Daily Care
- [ ] Daily progress notes
- [ ] Medication tracking
- [ ] Visit scheduling

### Phase 3: Advanced
- [ ] Cage management
- [ ] Occupancy reports
- [ ] Calendar view

---

## Integration Points

- **patients**: Link to patient, verificar activo
- **users**: Veterinarian attribution
- **appointments**: Link to appointment
- **notifications**: Notificar admisión/alta

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /hospitalizations | Admitir paciente |
| GET | /hospitalizations | Listar hospitalizaciones |
| GET | /hospitalizations/:id | Obtener |
| PUT | /hospitalizations/:id | Actualizar |
| DELETE | /hospitalizations/:id | Dar de alta |
| POST | /hospitalizations/:id/notes | Agregar nota diaria |
| GET | /hospitalizations/:id/notes | Obtener notas |
| GET | /hospitalizations/active | Actualmente admitidos |
| GET | /hospitalizations/patient/:patient_id | Historial por paciente |
| POST | /cages | Crear jaula |
| GET | /cages | Listar jaulas |
| GET | /cages/available | Jaulas disponibles |
| GET | /cages/:id | Obtener jaula |
| PUT | /cages/:id | Actualizar jaula |
| DELETE | /cages/:id | Eliminar jaula |

---

## DTOs (Data Transfer Objects)

### CreateHospitalizationDTO
```go
type CreateHospitalizationDTO struct {
    PatientID         string `json:"patient_id" binding:"required"`
    Reason            string `json:"reason" binding:"required"`
    InitialDiagnosis  string `json:"initial_diagnosis"`
    CageID            string `json:"cage_id" binding:"required"`
    ResponsibleVetID  string `json:"responsible_vet_id" binding:"required"`
}
```

### CreateCageDTO
```go
type CreateCageDTO struct {
    Name     string `json:"name" binding:"required"`
    Type     string `json:"type" binding:"required,oneof=dog cat isolation"`
    Capacity int    `json:"capacity" binding:"required,min=1"`
    Location string `json:"location"`
}
```

### CreateDailyProgressDTO
```go
type CreateDailyProgressDTO struct {
    Notes           string `json:"notes"`
    MedicationsGiven string `json:"medications_given"`
    FoodProvided    string `json:"food_provided"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| PATIENT_NOT_FOUND | El paciente no existe |
| PATIENT_INACTIVE | El paciente está inactivo |
| CAGE_NOT_FOUND | La jaula no existe |
| CAGE_NOT_AVAILABLE | La jaula no está disponible |
| CAGE_TYPE_MISMATCH | Tipo de jaula no coincide con especie |
| ALREADY_ADMITTED | El paciente ya está admitido |
| HOSPITALIZATION_NOT_FOUND | Hospitalización no encontrada |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice para Cage + IsAvailable (para query de disponibles)
- Crear índice para Hospitalization + Status
- Al admitir: marcar cage como no disponible
- Al dar de alta: marcar cage como disponible
- Verificar que el tipo de jaula coincida con la especie del paciente
