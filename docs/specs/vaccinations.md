# Spec: vaccinations

Scope: feature

# Vaccinations Module Specification

## Overview
Control de vacunas, recordatorios automáticos y certificados de vacunación.

## Domain Entities

### Vaccination (schema.go)
- ID, TenantID, PatientID, OwnerID
- VaccineName, Manufacturer, LotNumber
- ApplicationDate, NextDueDate
- VeterinarianID (who applied)
- CertificateNumber
- Notes
- Status (applied, due, overdue)
- CreatedAt, UpdatedAt

### Vaccine (schema.go) - Catalog
- ID, TenantID, Name, Description
- Manufacturer
- DoseNumber (1st, 2nd, booster)
- ValidityMonths
- TargetSpecies (dog, cat, etc.)
- Active

---

## Business Logic

### Validaciones

#### Al registrar vacunación (CreateVaccinationDTO)
| Campo | Regla |
|-------|-------|
| PatientID | Requerido, debe existir |
| VaccineName | Requerido |
| Manufacturer | Opcional |
| LotNumber | Opcional |
| ApplicationDate | Requerido, no puede ser futuro |
| NextDueDate | Opcional, debe ser > ApplicationDate |
| VeterinarianID | Requerido, debe existir |
| Notes | Opcional |

#### Al crear vaccine del catálogo (CreateVaccineDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido, único por tenant |
| Description | Opcional |
| Manufacturer | Opcional |
| DoseNumber | Requerido (first, second, booster) |
| ValidityMonths | Requerido, > 0 |
| TargetSpecies | Requerido |

### Reglas de Negocio

1. **Estado automático**: Se calcula automáticamente basado en fechas
   - `applied`: ApplicationDate existe y está en el pasado
   - `due`: Próxima fecha de vacunación dentro de 30 días
   - `overdue`: NextDueDate ya pasó
2. **Certificado automático**: Se genera número secuencial único por tenant
3. **No repetir疫苗**: No aplicar la misma vacuna antes de tiempo
4. **Paciente activo**: El paciente debe estar activo
5. **Especie compatible**: La vacuna debe ser para la especie del paciente

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `vaccination.created` | Se registra vacunación | Ninguna |
| `vaccination.reminder.24h` | 24h antes de NextDueDate | Notificar al owner |
| `vaccination.reminder.7d` | 7 días antes de NextDueDate | Notificar al owner |
| `vaccination.overdue` | NextDueDate pasó | Notificar al owner |
| `vaccination.due_today` | NextDueDate = hoy | Notificar al owner |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `vaccination.reminder_24h` | Owner (push + email) | 24h antes |
| `vaccination.reminder_7d` | Owner (push + email) | 7 días antes |
| `vaccination.overdue` | Owner (push + email) | Fecha pasada |
| `vaccination.due_today` | Owner (push + email) | Due date = hoy |

### Transiciones de Estado

```
No tiene fecha due
    ↓ (NextDueDate establecida)
due (dentro de 30 días)
    ↓ (pasa la fecha)
overdue
    ↓ (se registra nueva vacunación)
applied
```

---

## Features

### Phase 1: Basic CRUD
- [ ] Register vaccinations
- [ ] View vaccination history per patient
- [ ] List due/overdue vaccinations
- [ ] Generate certificate number

### Phaseinders
- [ 2: Rem ] Automatic reminders (24h, 7 days before due)
- [ ] Notifications to owners
- [ ] Mark as applied
- [ ] Scheduler para recordatorios

### Phase 3: Advanced
- [ ] Vaccine catalog management
- [ ] Certificate generation (PDF)
- [ ] Reports (compliance)
- [ ] Por especie

---

## Integration Points

- **patients**: Link to patient records, verificar especie
- **owners**: Owner notifications
- **notifications**: Send reminders
- **users**: Veterinarian attribution

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /vaccinations | Registrar vacunación |
| GET | /vaccinations | Listar vaccinations |
| GET | /vaccinations/:id | Obtener vacunación |
| PUT | /vaccinations/:id | Actualizar |
| DELETE | /vaccinations/:id | Eliminar |
| GET | /vaccinations/patient/:patient_id | Historial por paciente |
| GET | /vaccinations/due | Próximas a vencer |
| GET | /vaccinations/overdue | Vencidas |
| POST | /vaccines | Crear vaccine del catálogo |
| GET | /vaccines | Listar catálogo |
| PUT | /vaccines/:id | Actualizar vaccine |
| DELETE | /vaccines/:id | Eliminar vaccine |
| GET | /vaccines/:id/certificate | Generar certificado PDF |

---

## DTOs (Data Transfer Objects)

### CreateVaccinationDTO
```go
type CreateVaccinationDTO struct {
    PatientID       string `json:"patient_id" binding:"required"`
    VaccineName     string `json:"vaccine_name" binding:"required"`
    Manufacturer    string `json:"manufacturer"`
    LotNumber       string `json:"lot_number"`
    ApplicationDate string `json:"application_date" binding:"required"` // RFC3339
    NextDueDate     string `json:"next_due_date"` // RFC3339
    VeterinarianID  string `json:"veterinarian_id" binding:"required"`
    Notes           string `json:"notes"`
}
```

### CreateVaccineDTO
```go
type CreateVaccineDTO struct {
    Name            string   `json:"name" binding:"required"`
    Description     string   `json:"description"`
    Manufacturer    string   `json:"manufacturer"`
    DoseNumber      string   `json:"dose_number" binding:"required,oneof=first second booster"`
    ValidityMonths  int      `json:"validity_months" binding:"required,min=1"`
    TargetSpecies   []string `json:"target_species" binding:"required"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| PATIENT_NOT_FOUND | El paciente no existe |
| VETERINARIAN_NOT_FOUND | El veterinario no existe |
| VACCINE_NAME_REQUIRED | Nombre de vacuna requerido |
| NEXT_DUE_DATE_BEFORE_APPLICATION | Fecha de próxima aplicación no puede ser antes de la aplicación |
| PATIENT_INACTIVE | El paciente está inactivo |
| SPECIES_MISMATCH | La vacuna no es para la especie del paciente |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice para NextDueDate (para queries de vencimiento)
- Crear índice compuesto para PatientID + ApplicationDate
- El scheduler debe correr diariamente para:
  - Buscar vaccinations con NextDueDate = mañana → enviar recordatorio 24h
  - Buscar vaccinations con NextDueDate = hace 7 días → enviar recordatorio 7d
  - Buscar vaccinations con NextDueDate < hoy → marcar como overdue
- Generar PDF del certificado con: nombre del paciente, vacuna, fecha, veterinario, número de certificado
