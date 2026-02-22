# Spec: appointments-fixes

Scope: feature

# Spec: appointments-fixes

## Scope: Feature - Corrección completa y finalización del módulo Appointments

Este spec cubre la corrección de bugs críticos, alineación arquitectónica, funcionalidad faltante, integración con notificaciones, scheduler de recordatorios, y tests para el módulo de appointments.

---

## 1. BUGS CRÍTICOS A CORREGIR

### 1.1 Tenant Isolation Roto (handler.go)

**Problema**: Todos los handlers pasan `[]primitive.ObjectID{}` vacío al service/repo. El filtro MongoDB `$in: []` no matchea nada.

**Solución**: Extraer el tenant ID del contexto usando `sharedMiddleware.GetTenantID(c)` (retorna un solo `primitive.ObjectID`).

```go
// En CADA handler admin:
tenantID := sharedMiddleware.GetTenantID(c)

// En CADA handler mobile:
tenantID := sharedMiddleware.GetTenantID(c)
```

Esto requiere también cambiar la firma de los métodos del service y repository para aceptar `tenantID primitive.ObjectID` en vez de `tenantIDs []primitive.ObjectID` (ver sección 2.1).

### 1.2 OwnerID Aleatorio en CreateAppointment (service.go)

**Problema**: `service.go:63` asigna `OwnerID: primitive.NewObjectID()` (ID random) en vez del owner real.

**Solución**: Buscar el paciente en la BD y obtener su OwnerID real:
```go
// 1. Validar que el paciente existe
patient, err := s.patientRepo.FindByID(ctx, patientOID, tenantID)
if err != nil {
    return nil, ErrPatientNotFound
}
// 2. Usar el OwnerID del paciente
appointment.OwnerID = patient.OwnerID
```

### 1.3 VeterinarianID Aleatorio en RequestAppointment (service.go)

**Problema**: `RequestAppointment` asigna `VeterinarianID: primitive.NewObjectID()` random.

**Solución**: Usar `primitive.NilObjectID` para indicar "sin veterinario asignado" y que el staff lo asigne después:
```go
appointment.VeterinarianID = primitive.NilObjectID
```
Ajustar `CheckConflicts` para saltar la validación cuando `VeterinarianID` es nil.

### 1.4 Filtros date_from/date_to Rotos (handler.go → service.go)

**Problema**: El handler pasa strings al mapa de filtros, pero `parseFilters` hace type assertion a `time.Time`, falla silenciosamente.

**Solución**: Parsear las fechas en el handler antes de pasarlas:
```go
if dateFrom := c.Query("date_from"); dateFrom != "" {
    t, err := time.Parse(time.RFC3339, dateFrom)
    if err != nil {
        return nil, fmt.Errorf("invalid date_from format, use RFC3339: %w", err)
    }
    filters["date_from"] = t  // ahora es time.Time
}
```

---

## 2. ALINEACIÓN ARQUITECTÓNICA

### 2.1 Multi-Tenancy: Cambiar a Singular

**Cambiar en todo el módulo** de `tenant_ids []primitive.ObjectID` a `tenant_id primitive.ObjectID`:

**schema.go:**
```go
type Appointment struct {
    // ANTES: TenantIds []primitive.ObjectID `bson:"tenant_ids,omitempty"`
    TenantID primitive.ObjectID `bson:"tenant_id"` // DESPUÉS
    // ... resto igual
}

type AppointmentStatusTransition struct {
    // ANTES: TenantIds []primitive.ObjectID `bson:"tenant_ids,omitempty"`
    TenantID primitive.ObjectID `bson:"tenant_id"` // DESPUÉS
    // ... resto igual
}
```

**repository.go** - Cambiar TODAS las firmas y queries:
```go
// ANTES: FindByID(ctx, id, tenantIDs []primitive.ObjectID) 
// DESPUÉS:
FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error)

// ANTES: filter["tenant_ids"] = bson.M{"$in": tenantIDs}
// DESPUÉS:
filter["tenant_id"] = tenantID
```

Aplicar a TODOS los métodos del repository: `Create`, `FindByID`, `List`, `Update`, `Delete`, `FindByDateRange`, `FindByPatient`, `FindByOwner`, `FindByVeterinarian`, `CheckConflicts`, `FindUpcoming`, `CountByStatus`.

**service.go** - Cambiar TODAS las firmas:
```go
// ANTES: func (s *Service) CreateAppointment(ctx, dto, tenantIDs []primitive.ObjectID, createdBy)
// DESPUÉS:
func (s *Service) CreateAppointment(ctx context.Context, dto CreateAppointmentDTO, tenantID primitive.ObjectID, createdBy primitive.ObjectID) (*AppointmentResponse, error)
```

**handler.go** - Usar `sharedMiddleware.GetTenantID(c)`:
```go
tenantID := sharedMiddleware.GetTenantID(c)
// Pasar tenantID (singular) a todos los métodos del service
```

### 2.2 Paginación: Usar el Patrón Compartido

**Reemplazar** la paginación manual page/limit con el sistema compartido:

**handler.go:**
```go
import "github.com/eren_dev/go_server/internal/shared/pagination"

// ANTES:
// page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
// limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

// DESPUÉS:
params := pagination.FromContext(c)  // retorna Params{Skip, Limit}
```

**service.go / repository.go:**
```go
import "github.com/eren_dev/go_server/internal/shared/pagination"

// Cambiar firmas de List, FindByPatient, FindByOwner:
// ANTES: List(ctx, filters, tenantID, page, limit int)
// DESPUÉS:
List(ctx context.Context, filters appointmentFilters, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error)

// En el repo, usar params.Skip y params.Limit directamente:
opts := options.Find().SetSkip(params.Skip).SetLimit(params.Limit).SetSort(bson.D{{"scheduled_at", 1}})
```

**dto.go** - Reemplazar `PaginatedAppointmentsResponse`:
```go
type PaginatedAppointmentsResponse struct {
    Data       []AppointmentResponse    `json:"data"`
    Pagination pagination.PaginationInfo `json:"pagination"`
}

func CreatePaginatedResponse(appointments []Appointment, params pagination.Params, total int64) *PaginatedAppointmentsResponse {
    data := make([]AppointmentResponse, len(appointments))
    for i, a := range appointments {
        data[i] = *a.ToResponse()
    }
    return &PaginatedAppointmentsResponse{
        Data:       data,
        Pagination: pagination.NewPaginationInfo(params, total),
    }
}
```

### 2.3 Acceso a Collections

**Cambiar** en repository.go:
```go
// ANTES: db.Database().Collection("appointments")
// DESPUÉS:
db.Collection("appointments")
```

### 2.4 Limpiar DTOs No Usados

Eliminar `CreateAppointmentInternalDTO`, `ToInternalDTO()` y `ToEntity()` de dto.go ya que no son utilizados por el service.

---

## 3. FUNCIONALIDAD FALTANTE

### 3.1 Validación de Entidades Referenciadas

El service debe validar que las entidades existen antes de crear/actualizar una cita.

**Dependencias nuevas del service:**
```go
type Service struct {
    repo            AppointmentRepository
    patientRepo     patients.PatientRepository
    ownerRepo       owners.OwnerRepository
    userRepo        users.UserRepository
    notificationSvc *notifications.Service
}
```

**En CreateAppointment:**
```go
// 1. Validar paciente
patient, err := s.patientRepo.FindByID(ctx, patientOID, tenantID)
if err != nil {
    return nil, ErrPatientNotFound
}

// 2. Obtener OwnerID del paciente
appointment.OwnerID = patient.OwnerID

// 3. Validar veterinario (si se proporciona)
if !dto.VeterinarianID.IsZero() {
    vet, err := s.userRepo.FindByID(ctx, vetOID)
    if err != nil {
        return nil, ErrVeterinarianNotFound
    }
    // Opcionalmente verificar que el user tiene rol veterinario
}

// 4. Verificar conflictos de horario
hasConflict, err := s.repo.CheckConflicts(ctx, vetOID, dto.ScheduledAt, dto.Duration, nil, tenantID)
if hasConflict {
    return nil, ErrAppointmentConflict
}
```

**En RequestAppointment (mobile):**
```go
// 1. Validar que el paciente pertenece al owner autenticado
patient, err := s.patientRepo.FindByID(ctx, patientOID, tenantID)
if err != nil {
    return nil, ErrPatientNotFound
}
if patient.OwnerID != ownerOID {
    return nil, sharedErrors.ErrForbidden
}
```

### 3.2 Populate de Relaciones

Implementar population real de Patient, Owner y Veterinarian cuando `populate=true`.

**Cambiar los campos en AppointmentResponse** de `interface{}` a tipos concretos:
```go
type AppointmentResponse struct {
    // ... campos existentes ...
    Patient      *PatientSummary      `json:"patient,omitempty"`
    Owner        *OwnerSummary        `json:"owner,omitempty"`
    Veterinarian *VeterinarianSummary `json:"veterinarian,omitempty"`
}

// Summaries ligeros para evitar circular imports
type PatientSummary struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type OwnerSummary struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Phone string `json:"phone"`
}

type VeterinarianSummary struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

**En service, crear método `populateAppointment`:**
```go
func (s *Service) populateAppointment(ctx context.Context, appointment *Appointment, tenantID primitive.ObjectID) (*AppointmentResponse, error) {
    resp := appointment.ToResponse()
    
    // Populate patient
    if patient, err := s.patientRepo.FindByID(ctx, appointment.PatientID, tenantID); err == nil {
        resp.Patient = &PatientSummary{ID: patient.ID.Hex(), Name: patient.Name}
    }
    
    // Populate owner
    if owner, err := s.ownerRepo.FindByID(ctx, appointment.OwnerID); err == nil {
        resp.Owner = &OwnerSummary{ID: owner.ID.Hex(), Name: owner.FirstName + " " + owner.LastName, Email: owner.Email, Phone: owner.Phone}
    }
    
    // Populate veterinarian (solo si tiene uno asignado)
    if !appointment.VeterinarianID.IsZero() {
        if vet, err := s.userRepo.FindByID(ctx, appointment.VeterinarianID); err == nil {
            resp.Veterinarian = &VeterinarianSummary{ID: vet.ID.Hex(), Name: vet.Name, Email: vet.Email}
        }
    }
    
    return resp, nil
}
```

### 3.3 Cancelación desde Mobile

Agregar endpoint para que owners puedan cancelar sus citas:

**handler.go:**
```go
// @Summary Cancel appointment (mobile)
// @Tags mobile-appointments
func (h *Handler) CancelOwnerAppointment(c *gin.Context) (any, error) {
    ownerID := sharedAuth.GetUserID(c)
    tenantID := sharedMiddleware.GetTenantID(c)
    id := c.Param("id")
    
    var dto struct {
        Reason string `json:"reason" binding:"required,max=200"`
    }
    if err := c.ShouldBindJSON(&dto); err != nil {
        return nil, validation.Validate(err)
    }
    
    // Verificar que la cita pertenece al owner
    appointment, err := h.service.GetAppointment(ctx, id, tenantID, false)
    if err != nil {
        return nil, err
    }
    ownerOID, _ := primitive.ObjectIDFromHex(ownerID)
    if appointment.OwnerID != ownerID {
        return nil, sharedErrors.ErrForbidden
    }
    
    return h.service.CancelAppointment(ctx, id, dto.Reason, tenantID, ownerOID)
}
```

**router.go (mobile):**
```go
appointments.PATCH("/:id/cancel", httpx.Router(handler.CancelOwnerAppointment))
```

### 3.4 Validación de Horario Laboral

**En service.go, agregar `validateAppointmentTime`:**
```go
func (s *Service) validateAppointmentTime(scheduledAt time.Time) error {
    // No permitir citas en el pasado
    if scheduledAt.Before(time.Now()) {
        return ErrPastAppointmentTime
    }
    
    // Validar horario laboral (configurable, por defecto 8:00-18:00)
    hour := scheduledAt.Hour()
    if hour < 8 || hour >= 18 {
        return ErrInvalidAppointmentTime
    }
    
    // No permitir domingos
    if scheduledAt.Weekday() == time.Sunday {
        return ErrInvalidAppointmentTime
    }
    
    return nil
}
```

---

## 4. INTEGRACIÓN CON NOTIFICACIONES

### 4.1 Dependencia

**router.go** - Cambiar la firma para recibir el push provider:
```go
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB, pushProvider platformNotifications.PushProvider) {
    repo := NewAppointmentRepository(db)
    patientRepo := patients.NewPatientRepository(db)
    ownerRepo := owners.NewRepository(db)
    userRepo := users.NewUserRepository(db)
    
    // Crear notification service
    notifSvc := notifications.NewService(
        notifications.NewRepository(db),
        notifications.NewStaffRepository(db),
        ownerRepo,
        pushProvider,
    )
    
    service := NewService(repo, patientRepo, ownerRepo, userRepo, notifSvc)
    handler := NewHandler(service)
    // ... rutas
}

func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB, pushProvider platformNotifications.PushProvider) {
    // Mismo patrón
}
```

**app/router.go** - Actualizar la llamada:
```go
// ANTES: appointments.RegisterAdminRoutes(privateTenant, db)
// DESPUÉS:
appointments.RegisterAdminRoutes(privateTenant, db, pushProvider)
appointments.RegisterMobileRoutes(mobileTenant, db, pushProvider)
```

### 4.2 Notificaciones por Evento

**Al crear cita (staff):**
```go
// Notificar al owner
s.notificationSvc.Send(ctx, &notifications.SendDTO{
    OwnerID:  appointment.OwnerID.Hex(),
    TenantID: tenantID.Hex(),
    Type:     notifications.TypeAppointmentReminder,
    Title:    "Nueva cita agendada",
    Body:     fmt.Sprintf("Se ha agendado una cita para %s el %s", patient.Name, appointment.ScheduledAt.Format("02/01/2006 15:04")),
    Data:     map[string]string{"appointment_id": appointment.ID.Hex(), "patient_id": appointment.PatientID.Hex()},
    SendPush: true,
})

// Notificar al veterinario
s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
    UserID:   appointment.VeterinarianID.Hex(),
    TenantID: tenantID.Hex(),
    Type:     notifications.TypeStaffNewAppointment,
    Title:    "Nueva cita asignada",
    Body:     fmt.Sprintf("Cita con %s el %s", patient.Name, appointment.ScheduledAt.Format("02/01/2006 15:04")),
    Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
})
```

**Al confirmar cita:**
```go
s.notificationSvc.Send(ctx, &notifications.SendDTO{
    OwnerID:  appointment.OwnerID.Hex(),
    TenantID: tenantID.Hex(),
    Type:     notifications.TypeAppointmentConfirmed,
    Title:    "Cita confirmada",
    Body:     fmt.Sprintf("Tu cita del %s ha sido confirmada", appointment.ScheduledAt.Format("02/01/2006 15:04")),
    Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
    SendPush: true,
})
```

**Al cancelar cita:**
```go
s.notificationSvc.Send(ctx, &notifications.SendDTO{
    OwnerID:  appointment.OwnerID.Hex(),
    TenantID: tenantID.Hex(),
    Type:     notifications.TypeAppointmentCancelled,
    Title:    "Cita cancelada",
    Body:     fmt.Sprintf("La cita del %s ha sido cancelada. Razón: %s", appointment.ScheduledAt.Format("02/01/2006 15:04"), reason),
    Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
    SendPush: true,
})
```

**Al solicitar cita (mobile):**
```go
// Notificar a staff (sin veterinario específico, notificar a todos los admins del tenant)
// Esto requiere un método adicional en notificationSvc o enviar a un userID admin genérico
s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
    UserID:   adminUserID.Hex(), // TODO: definir cómo obtener admins del tenant
    TenantID: tenantID.Hex(),
    Type:     notifications.TypeStaffNewAppointment,
    Title:    "Nueva solicitud de cita",
    Body:     fmt.Sprintf("Solicitud de cita de %s para %s", ownerName, patientName),
    Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
})
```

---

## 5. SCHEDULER DE RECORDATORIOS Y AUTO-CANCELACIÓN

### 5.1 Estructura

Crear `internal/scheduler/scheduler.go`:

```go
package scheduler

import (
    "context"
    "log/slog"
    "time"
    
    "github.com/eren_dev/go_server/internal/modules/appointments"
    "github.com/eren_dev/go_server/internal/modules/notifications"
    "github.com/eren_dev/go_server/internal/shared/database"
    "github.com/eren_dev/go_server/internal/shared/lifecycle"
)

type Scheduler struct {
    appointmentRepo appointments.AppointmentRepository
    notificationSvc *notifications.Service
    interval        time.Duration
    logger          *slog.Logger
    stopCh          chan struct{}
}

func New(db *database.MongoDB, notificationSvc *notifications.Service, logger *slog.Logger) *Scheduler {
    return &Scheduler{
        appointmentRepo: appointments.NewAppointmentRepository(db),
        notificationSvc: notificationSvc,
        interval:        15 * time.Minute, // check every 15 minutes
        logger:          logger,
        stopCh:          make(chan struct{}),
    }
}

func (s *Scheduler) Start(ctx context.Context, workers *lifecycle.Workers) {
    workers.Add(1)
    go func() {
        defer workers.Done()
        ticker := time.NewTicker(s.interval)
        defer ticker.Stop()
        
        s.logger.Info("appointment scheduler started", "interval", s.interval)
        
        for {
            select {
            case <-ticker.C:
                s.processReminders(ctx)
                s.processAutoCancellations(ctx)
            case <-s.stopCh:
                s.logger.Info("appointment scheduler stopped")
                return
            case <-ctx.Done():
                s.logger.Info("appointment scheduler context cancelled")
                return
            }
        }
    }()
}

func (s *Scheduler) Stop() {
    close(s.stopCh)
}
```

### 5.2 Recordatorios (24h y 2h antes)

```go
func (s *Scheduler) processReminders(ctx context.Context) {
    // Buscar citas próximas en las siguientes 24h que estén confirmed
    upcoming, err := s.appointmentRepo.FindUpcoming(ctx, primitive.NilObjectID, 24)
    if err != nil {
        s.logger.Error("failed to find upcoming appointments", "error", err)
        return
    }
    
    now := time.Now()
    for _, appt := range upcoming {
        if appt.Status != "confirmed" {
            continue
        }
        
        timeUntil := appt.ScheduledAt.Sub(now)
        
        // Recordatorio 24h (entre 23h30m y 24h30m)
        if timeUntil >= 23*time.Hour+30*time.Minute && timeUntil <= 24*time.Hour+30*time.Minute {
            s.sendReminder(ctx, &appt, "24 horas")
        }
        
        // Recordatorio 2h (entre 1h30m y 2h30m)
        if timeUntil >= 1*time.Hour+30*time.Minute && timeUntil <= 2*time.Hour+30*time.Minute {
            s.sendReminder(ctx, &appt, "2 horas")
        }
    }
}

func (s *Scheduler) sendReminder(ctx context.Context, appt *appointments.Appointment, timeframe string) {
    s.notificationSvc.Send(ctx, &notifications.SendDTO{
        OwnerID:  appt.OwnerID.Hex(),
        TenantID: appt.TenantID.Hex(),
        Type:     notifications.TypeAppointmentReminder,
        Title:    "Recordatorio de cita",
        Body:     fmt.Sprintf("Tu cita es en %s (%s)", timeframe, appt.ScheduledAt.Format("02/01/2006 15:04")),
        Data:     map[string]string{"appointment_id": appt.ID.Hex()},
        SendPush: true,
    })
}
```

### 5.3 Auto-Cancelación de Citas No Confirmadas

```go
func (s *Scheduler) processAutoCancellations(ctx context.Context) {
    // Buscar citas con status "scheduled" que fueron creadas hace más de 24h
    cutoff := time.Now().Add(-24 * time.Hour)
    
    // Necesita un método nuevo en el repo:
    // FindUnconfirmedBefore(ctx, cutoff) ([]Appointment, error)
    unconfirmed, err := s.appointmentRepo.FindUnconfirmedBefore(ctx, cutoff)
    if err != nil {
        s.logger.Error("failed to find unconfirmed appointments", "error", err)
        return
    }
    
    for _, appt := range unconfirmed {
        // Auto-cancelar
        now := time.Now()
        updates := bson.M{
            "$set": bson.M{
                "status":        "cancelled",
                "cancelled_at":  now,
                "cancel_reason": "Auto-cancelada: no confirmada en 24 horas",
                "updated_at":    now,
            },
        }
        if err := s.appointmentRepo.Update(ctx, appt.ID, updates, appt.TenantID); err != nil {
            s.logger.Error("failed to auto-cancel appointment", "id", appt.ID.Hex(), "error", err)
            continue
        }
        
        // Notificar al owner
        s.notificationSvc.Send(ctx, &notifications.SendDTO{
            OwnerID:  appt.OwnerID.Hex(),
            TenantID: appt.TenantID.Hex(),
            Type:     notifications.TypeAppointmentCancelled,
            Title:    "Cita cancelada automáticamente",
            Body:     fmt.Sprintf("La cita del %s fue cancelada por no ser confirmada en 24 horas", appt.ScheduledAt.Format("02/01/2006 15:04")),
            Data:     map[string]string{"appointment_id": appt.ID.Hex()},
            SendPush: true,
        })
        
        s.logger.Info("auto-cancelled unconfirmed appointment", "id", appt.ID.Hex())
    }
}
```

### 5.4 Método Nuevo en Repository

Agregar al interface `AppointmentRepository`:
```go
FindUnconfirmedBefore(ctx context.Context, before time.Time) ([]Appointment, error)
```

Implementación:
```go
func (r *appointmentRepository) FindUnconfirmedBefore(ctx context.Context, before time.Time) ([]Appointment, error) {
    filter := bson.M{
        "status":     "scheduled",
        "created_at": bson.M{"$lt": before},
        "deleted_at": nil,
    }
    
    cursor, err := r.collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var appointments []Appointment
    if err := cursor.All(ctx, &appointments); err != nil {
        return nil, err
    }
    return appointments, nil
}
```

### 5.5 Integración en main.go

```go
// En cmd/api/main.go, después de crear el server:
appointmentScheduler := scheduler.New(db, notificationSvc, logger)
appointmentScheduler.Start(ctx, workers)

// En el shutdown:
appointmentScheduler.Stop()
```

---

## 6. ÍNDICES DE MONGODB

Crear script de inicialización o método de setup que cree los índices necesarios.

**Opción recomendada**: Agregar un método `EnsureIndexes` en el repository que se llame al inicializar:

```go
func (r *appointmentRepository) EnsureIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        {Keys: bson.D{{"tenant_id", 1}, {"scheduled_at", 1}}},
        {Keys: bson.D{{"tenant_id", 1}, {"veterinarian_id", 1}, {"scheduled_at", 1}}},
        {Keys: bson.D{{"tenant_id", 1}, {"patient_id", 1}, {"scheduled_at", -1}}},
        {Keys: bson.D{{"tenant_id", 1}, {"owner_id", 1}, {"scheduled_at", -1}}},
        {Keys: bson.D{{"tenant_id", 1}, {"status", 1}, {"scheduled_at", 1}}},
        {Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
    }
    _, err := r.collection.Indexes().CreateMany(ctx, indexes)
    return err
}

// Para status transitions:
func (r *appointmentRepository) EnsureTransitionIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        {Keys: bson.D{{"appointment_id", 1}, {"created_at", -1}}},
        {Keys: bson.D{{"tenant_id", 1}, {"changed_by", 1}, {"created_at", -1}}},
    }
    _, err := r.transitionCollection.Indexes().CreateMany(ctx, indexes)
    return err
}
```

Llamar en `RegisterAdminRoutes`:
```go
repo := NewAppointmentRepository(db)
if err := repo.EnsureIndexes(context.Background()); err != nil {
    slog.Error("failed to ensure appointment indexes", "error", err)
}
```

---

## 7. TESTS

### 7.1 Tests Unitarios del Service

Crear `internal/modules/appointments/service_test.go`:
- Mock del `AppointmentRepository` interface
- Mock del `notifications.Service`
- Tests para:
  - `CreateAppointment` - happy path, paciente no encontrado, conflicto horario, horario inválido
  - `UpdateStatus` - transiciones válidas e inválidas
  - `RequestAppointment` - happy path, paciente no pertenece al owner
  - `CancelAppointment` - happy path, ya cancelada
  - `validateAppointmentTime` - pasado, fuera de horario, domingo
  - `validateStatusTransition` - todas las transiciones válidas e inválidas

### 7.2 Tests de Integración del Repository

Crear `internal/modules/appointments/repository_test.go`:
- Usar MongoDB en memoria o testcontainers
- Tests para:
  - CRUD básico con tenant isolation
  - `CheckConflicts` - sin conflicto, con conflicto, excluyendo ID
  - `FindByDateRange` - rango válido, sin resultados
  - `FindUpcoming` - citas próximas
  - `FindUnconfirmedBefore` - citas sin confirmar
  - Soft delete (verificar que deleted_at filtra correctamente)

### 7.3 Tests del Handler

Crear `internal/modules/appointments/handler_test.go`:
- Usar `httptest` con gin test context
- Tests para:
  - Validación de DTOs (campos requeridos, rangos)
  - Extracción correcta de tenant ID
  - Respuestas de error apropiadas
  - Filtros de query params

### 7.4 Ejecutar tests
```bash
make test  # go test ./... -race -count=1
```

---

## 8. SWAGGER

Verificar y actualizar las anotaciones Swagger en handler.go para reflejar:
- Nuevos endpoints (cancel mobile)
- Parámetros de paginación actualizados (skip/limit en vez de page/limit)
- Modelo de respuesta actualizado
- Regenerar con `make docs`

---

## 9. RESUMEN DE ARCHIVOS A MODIFICAR

| Archivo | Acción |
|---------|--------|
| `internal/modules/appointments/schema.go` | Cambiar `TenantIds` → `TenantID` |
| `internal/modules/appointments/dto.go` | Actualizar pagination, limpiar DTOs no usados, agregar summary types |
| `internal/modules/appointments/repository.go` | Cambiar tenancy, agregar `FindUnconfirmedBefore`, `EnsureIndexes`, fix collection access |
| `internal/modules/appointments/service.go` | Fix bugs (OwnerID, VetID), agregar validaciones, notificaciones, populate, nueva firma |
| `internal/modules/appointments/handler.go` | Fix tenant extraction, pagination, date parsing, agregar cancel mobile |
| `internal/modules/appointments/router.go` | Agregar pushProvider, ruta cancel mobile, EnsureIndexes |
| `internal/modules/appointments/errors.go` | Sin cambios mayores (ya está bien) |
| `internal/app/router.go` | Pasar pushProvider a appointments routes |
| `cmd/api/main.go` | Inicializar y arrancar el scheduler |
| `internal/scheduler/scheduler.go` | **NUEVO** - scheduler de recordatorios y auto-cancelación |
| `internal/modules/appointments/service_test.go` | **NUEVO** - tests unitarios |
| `internal/modules/appointments/repository_test.go` | **NUEVO** - tests integración |
| `internal/modules/appointments/handler_test.go` | **NUEVO** - tests handler |