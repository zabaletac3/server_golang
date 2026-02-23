# Spec: medical-records

Scope: feature

# Medical Records Module Specification

## Overview
Historial clínico completo de pacientes: diagnósticos, tratamientos, evoluciones, alergias, antecedentes.

## Domain Entities

### MedicalRecord (schema.go)
- ID, TenantID, PatientID, VeterinarianID
- Type (consultation, emergency, surgery, checkup, vaccination)
- ChiefComplaint (reason for visit)
- Diagnosis
- Symptoms
- Weight, Temperature (vitals)
- Treatment (prescribed treatments)
- Medications (list with dosage)
- EvolutionNotes
- Attachments (file IDs)
- NextVisitDate
- CreatedAt, UpdatedAt

### Allergies (schema.go)
- ID, TenantID, PatientID
- Allergen (drug, food, other)
- Severity (mild, moderate, severe)
- Description

### MedicalHistory (schema.go) - Summary
- ID, TenantID, PatientID
- ChronicConditions
- PreviousSurgeries
- RiskFactors

---

## Business Logic

### Validaciones

#### Al crear medical record (CreateMedicalRecordDTO)
| Campo | Regla |
|-------|-------|
| PatientID | Requerido, debe existir y estar activo |
| VeterinarianID | Requerido, debe existir |
| Type | Requerido (consultation, emergency, surgery, checkup, vaccination) |
| ChiefComplaint | Requerido |
| Diagnosis | Opcional |
| Symptoms | Opcional |
| Weight | Opcional, > 0 |
| Temperature | Opcional, entre 30-45°C |
| NextVisitDate | Opcional, no puede ser pasado |

#### Al registrar alergia (CreateAllergyDTO)
| Campo | Regla |
|-------|-------|
| PatientID | Requerido, debe existir |
| Allergen | Requerido |
| Severity | Requerido (mild, moderate, severe) |
| Description | Opcional |

### Reglas de Negocio

1. **Paciente activo**: No se puede crear historial si el paciente está inactivo
2. **Veterinario requerido**: Todo registro debe tener un veterinario responsable
3. **Alergias visibles**: Las alergias del paciente se muestran en todos los registros
4. **Peso y temperatura**: Se guardan en cada visita para tracking
5. **Próxima visita**: Si se indica, se crea recordatorio
6. **Evolución**: Las notas de evolución se agregan al registro existente
7. **Adjuntos**: Solo archivos PDF e imágenes (jpg, png)
8. **Historial read-only**: No se puede editar un registro después de 24 horas (para compliance)
9. **Severidad de alergias**: Alertas visibles si severity = severe

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `medical_record.created` | Se crea registro | Ninguna |
| `medical_record.updated` | Se actualiza registro (dentro de 24h) | Ninguna |
| `medical_record.next_visit` | NextVisitDate establecida | Crear cita automaticamente |
| `allergy.created` | Nueva alergia registrada | Ninguna |
| `allergy.severe` | Alergia severity=severe | Alertar en próximas visitas |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `medical_record.next_visit_reminder` | Owner | Un día antes de NextVisitDate |

### Transiciones de Estado

```
No aplica - MedicalRecord es un registro histórico inmutable después de 24h
```

---

## Features

### Phase 1: Basic CRUD
- [ ] Create medical records
- [ ] View history per patient
- [ ] Search records
- [ ] Validar paciente activo

### Phase 2: Timeline
- [ ] Chronological view
- [ ] Filter by type, date, vet
- [ ] Weight/temperature tracking

### Phase 3: Advanced
- [ ] Attachments (images, PDFs)
- [ ] Allergies management
- [ ] Chronic conditions
- [ ] Templates for common visits
- [ ] Export to PDF
- [ ] Next visit reminders

---

## Integration Points

- **patients**: Link to patient, verificar activo
- **users**: Veterinarian attribution
- **appointments**: Link to appointments, crear cita desde NextVisitDate
- **resources**: Attachments (PDFs, imágenes)
- **notifications**: Recordatorios de próxima visita

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /medical-records | Crear registro |
| GET | /medical-records | Listar registros |
| GET | /medical-records/:id | Obtener registro |
| PUT | /medical-records/:id | Actualizar (solo dentro de 24h) |
| DELETE | /medical-records/:id | Eliminar (soft delete) |
| GET | /medical-records/patient/:patient_id | Historial por paciente |
| GET | /medical-records/patient/:patient_id/timeline | Timeline cronológico |
| POST | /allergies | Registrar alergia |
| GET | /allergies/patient/:patient_id | Alergias del paciente |
| PUT | /allergies/:id | Actualizar alergia |
| DELETE | /allergies/:id | Eliminar alergia |
| POST | /medical-history | Crear/resumen historial |
| GET | /medical-history/patient/:patient_id | Obtener resumen |

---

## DTOs (Data Transfer Objects)

### CreateMedicalRecordDTO
```go
type CreateMedicalRecordDTO struct {
    PatientID       string   `json:"patient_id" binding:"required"`
    VeterinarianID  string   `json:"veterinarian_id" binding:"required"`
    Type            string   `json:"type" binding:"required,oneof=consultation emergency surgery checkup vaccination"`
    ChiefComplaint string   `json:"chief_complaint" binding:"required"`
    Diagnosis       string   `json:"diagnosis"`
    Symptoms        string   `json:"symptoms"`
    Weight          float64  `json:"weight" binding:"omitempty,min=0"`
    Temperature     float64  `json:"temperature" binding:"omitempty,min=30,max=45"`
    Treatment       string   `json:"treatment"`
    Medications     []MedicationDTO `json:"medications"`
    EvolutionNotes  string   `json:"evolution_notes"`
    AttachmentIDs    []string `json:"attachment_ids"`
    NextVisitDate   string   `json:"next_visit_date"` // RFC3339
}

type MedicationDTO struct {
    Name     string `json:"name"`
    Dose     string `json:"dose"`
    Frequency string `json:"frequency"`
    Duration string `json:"duration"`
}
```

### CreateAllergyDTO
```go
type CreateAllergyDTO struct {
    PatientID  string `json:"patient_id" binding:"required"`
    Allergen   string `json:"allergen" binding:"required"`
    Severity   string `json:"severity" binding:"required,oneof=mild moderate severe"`
    Description string `json:"description"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| PATIENT_NOT_FOUND | El paciente no existe |
| PATIENT_INACTIVE | El paciente está inactivo |
| VETERINARIAN_NOT_FOUND | El veterinario no existe |
| RECORD_NOT_EDITABLE | El registro no se puede editar después de 24h |
| INVALID_TEMPERATURE | Temperatura fuera de rango válido (30-45°C) |
| ATTACHMENT_INVALID | Tipo de archivo no válido |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice compuesto para PatientID + CreatedAt (para timeline)
- Validar que las AttachmentIDs existan en resources
- No permitir edición después de 24 horas (guardar timestamp de creación)
- Las alergias severe deben mostrarse prominentemente en la UI
- Crear función helper para calcular edad del paciente a partir de fecha de nacimiento
