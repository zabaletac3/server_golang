# Spec: laboratory

Scope: feature

# Laboratory Module Specification

## Overview
Gestión de resultados de exámenes de laboratorio externos, solicitudes y seguimiento.

## Domain Entities

### LabOrder (schema.go)
- ID, TenantID, PatientID, OwnerID
- OrderDate, CollectionDate, ResultDate
- VeterinarianID (requested by)
- LabID (external lab name)
- TestType (blood, urine, biopsy, etc.)
- Status (pending, collected, sent, received, processed)
- ResultFile (attachment)
- Notes
- Cost
- CreatedAt, UpdatedAt

### LabTest (schema.go) - Catalog
- ID, TenantID, Name
- Description
- Category (hematology, biochemistry, urinalysis, etc.)
- Price
- TurnaroundTime (days)
- Active

---

## Business Logic

### Validaciones

#### Al crear orden (CreateLabOrderDTO)
| Campo | Regla |
|-------|-------|
| PatientID | Requerido, debe existir y estar activo |
| VeterinarianID | Requerido, debe existir |
| TestType | Requerido |
| LabID | Opcional |
| Notes | Opcional |

#### Al subir resultado (UploadResultDTO)
| Campo | Regla |
|-------|-------|
| ResultFile | Requerido, archivo PDF |

### Reglas de Negocio

1. **Estados válidos**: pending → collected → sent → received → processed
2. **No saltarse estados**: Debe seguir la secuencia
3. **Resultado solo si collected**: No se puede subir resultado sin haber collected
4. **Costo automático**: Se obtiene del catálogo si existe
5. **Fecha de resultado**: Se guarda automáticamente al subir resultado
6. **Paciente activo**: El paciente debe estar activo
7. **Notificación owner**: Cuando status = processed, notificar al owner

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `lab_order.created` | Se crea orden | Ninguna |
| `lab_order.status.changed` | Cambio de estado | Ninguna |
| `lab_order.result_ready` | Status = processed | Notificar al owner |
| `lab_order.overdue` | Pasó TurnaroundTime | Notificar al staff |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `lab.result_ready` | Owner (push + email) | Resultados listos |
| `lab.overdue` | Staff | Pasó tiempo estimado |

### Transiciones de Estado

```
pending (orden creada)
    ↓ (se recolecta muestra)
collected
    ↓ (se envía al lab externo)
sent
    ↓ (el lab devuelve resultados)
received
    ↓ (se procesa el resultado)
processed (resultado disponible)
```

---

## Features

### Phase 1: Basic
- [ ] Create lab orders
- [ ] Link to patient
- [ ] Upload results

### Phase 2: Tracking
- [ ] Status tracking (pending → collected → sent → received)
- [ ] Notifications when results ready
- [ ] Alerts for overdue

### Phase 3: Integration
- [ ] Lab test catalog
- [ ] Price list
- [ ] Reports

---

## Integration Points

- **patients**: Link to patient, verificar activo
- **owners**: Notify owner of results
- **resources**: File attachments (PDFs)
- **payments**: Lab costs en factura

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | /lab-orders | Crear orden |
| GET | /lab-orders | Listar órdenes |
| GET | /lab-orders/:id | Obtener orden |
| PUT | /lab-orders/:id | Actualizar orden |
| PATCH | /lab-orders/:id/status | Cambiar estado |
| DELETE | /lab-orders/:id | Eliminar |
| GET | /lab-orders/patient/:patient_id | Órdenes por paciente |
| POST | /lab-orders/:id/result | Subir resultado |
| GET | /lab-orders/:id/result | Descargar resultado |
| POST | /lab-tests | Crear prueba catálogo |
| GET | /lab-tests | Listar catálogo |
| PUT | /lab-tests/:id | Actualizar prueba |
| DELETE | /lab-tests/:id | Eliminar prueba |

---

## DTOs (Data Transfer Objects)

### CreateLabOrderDTO
```go
type CreateLabOrderDTO struct {
    PatientID      string `json:"patient_id" binding:"required"`
    VeterinarianID string `json:"veterinarian_id" binding:"required"`
    TestType       string `json:"test_type" binding:"required"`
    LabID          string `json:"lab_id"`
    Notes          string `json:"notes"`
}
```

### UpdateLabOrderStatusDTO
```go
type UpdateLabOrderStatusDTO struct {
    Status string `json:"status" binding:"required,oneof=pending collected sent received processed"`
    Notes  string `json:"notes"`
}
```

### UploadResultDTO
```go
type UploadResultDTO struct {
    ResultFile string `json:"result_file" binding:"required"` // File ID from resources
}
```

### CreateLabTestDTO
```go
type CreateLabTestDTO struct {
    Name           string `json:"name" binding:"required"`
    Description    string `json:"description"`
    Category       string `json:"category" binding:"required"`
    Price          float64 `json:"price" binding:"required,min=0"`
    TurnaroundTime int    `json:"turnaround_time" binding:"required,min=1"` // days
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| PATIENT_NOT_FOUND | El paciente no existe |
| PATIENT_INACTIVE | El paciente está inactivo |
| VETERINARIAN_NOT_FOUND | El veterinarian no existe |
| INVALID_STATUS_TRANSITION | No se puede cambiar directamente a este estado |
| RESULT_WITHOUT_COLLECTION | No se puede subir resultado sin recolectar muestra |
| TEST_NOT_FOUND | La prueba no existe en el catálogo |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Soft delete
- Crear índice para PatientID + Status
- Crear índice para OrderDate (para queries)
- El scheduler debe verificar órdenes que pasaron el TurnaroundTime
- Validar que el ResultFile exista en resources y sea PDF
