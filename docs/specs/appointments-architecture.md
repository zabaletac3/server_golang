# Spec: appointments-architecture

Scope: feature

# Appointments System - Feature Specification

## Architecture Overview
Following the established module pattern with 7 core files: dto.go, schema.go, service.go, handler.go, repository.go, router.go, errors.go

## File Structure: `internal/modules/appointments/`

### 1. `schema.go` - Domain Entity
```go
type Appointment struct {
    ID          primitive.ObjectID   `bson:"_id,omitempty"`
    TenantIds   []primitive.ObjectID `bson:"tenant_ids,omitempty"`
    
    // Core appointment data
    PatientID    primitive.ObjectID `bson:"patient_id"`
    OwnerID      primitive.ObjectID `bson:"owner_id"`
    VeterinarianID primitive.ObjectID `bson:"veterinarian_id"`
    
    // Scheduling
    ScheduledAt  time.Time `bson:"scheduled_at"`
    Duration     int       `bson:"duration"` // minutes
    
    // Appointment details
    Type        string `bson:"type"`        // consultation, surgery, vaccination, etc.
    Status      string `bson:"status"`      // scheduled, confirmed, in_progress, completed, cancelled, no_show
    Priority    string `bson:"priority"`    // low, normal, high, emergency
    
    // Notes and observations
    Reason      string `bson:"reason"`      // Reason for visit
    Notes       string `bson:"notes,omitempty"`       // Staff notes
    OwnerNotes  string `bson:"owner_notes,omitempty"` // Owner notes
    
    // Status tracking
    ConfirmedAt *time.Time `bson:"confirmed_at,omitempty"`
    StartedAt   *time.Time `bson:"started_at,omitempty"`
    CompletedAt *time.Time `bson:"completed_at,omitempty"`
    CancelledAt *time.Time `bson:"cancelled_at,omitempty"`
    CancelReason string    `bson:"cancel_reason,omitempty"`
    
    // Standard fields
    CreatedAt time.Time  `bson:"created_at"`
    UpdatedAt time.Time  `bson:"updated_at"`
    DeletedAt *time.Time `bson:"deleted_at,omitempty"`
    
    // Populated fields (not stored in DB)
    Patient      *patients.Patient `bson:"-"`
    Owner        *owners.Owner     `bson:"-"`
    Veterinarian *users.User       `bson:"-"`
}

type AppointmentStatusTransition struct {
    ID           primitive.ObjectID `bson:"_id,omitempty"`
    TenantIds    []primitive.ObjectID `bson:"tenant_ids,omitempty"`
    AppointmentID primitive.ObjectID `bson:"appointment_id"`
    FromStatus   string             `bson:"from_status"`
    ToStatus     string             `bson:"to_status"`
    ChangedBy    primitive.ObjectID `bson:"changed_by"`
    Reason       string             `bson:"reason,omitempty"`
    CreatedAt    time.Time          `bson:"created_at"`
}
```

### 2. `dto.go` - Data Transfer Objects
```go
// Input DTOs
type CreateAppointmentDTO struct {
    PatientID      string    `json:"patient_id" binding:"required" example:"507f1f77bcf86cd799439011"`
    VeterinarianID string    `json:"veterinarian_id" binding:"required" example:"507f1f77bcf86cd799439012"`
    ScheduledAt    time.Time `json:"scheduled_at" binding:"required" example:"2024-01-15T10:30:00Z"`
    Duration       int       `json:"duration" binding:"required,min=15,max=480" example:"30"`
    Type           string    `json:"type" binding:"required" example:"consultation"`
    Priority       string    `json:"priority" binding:"omitempty,oneof=low normal high emergency" example:"normal"`
    Reason         string    `json:"reason" binding:"required,max=500" example:"Annual checkup"`
    Notes          string    `json:"notes" binding:"omitempty,max=1000" example:"First visit for this patient"`
}

type UpdateAppointmentDTO struct {
    ScheduledAt *time.Time `json:"scheduled_at" binding:"omitempty" example:"2024-01-15T11:00:00Z"`
    Duration    *int       `json:"duration" binding:"omitempty,min=15,max=480" example:"45"`
    Type        *string    `json:"type" binding:"omitempty" example:"surgery"`
    Priority    *string    `json:"priority" binding:"omitempty,oneof=low normal high emergency" example:"high"`
    Reason      *string    `json:"reason" binding:"omitempty,max=500" example:"Updated reason"`
    Notes       *string    `json:"notes" binding:"omitempty,max=1000" example:"Updated notes"`
}

type UpdateStatusDTO struct {
    Status string `json:"status" binding:"required,oneof=confirmed in_progress completed cancelled no_show" example:"confirmed"`
    Reason string `json:"reason" binding:"omitempty,max=200" example:"Patient confirmed by phone"`
}

type MobileAppointmentRequestDTO struct {
    PatientID   string    `json:"patient_id" binding:"required" example:"507f1f77bcf86cd799439011"`
    ScheduledAt time.Time `json:"scheduled_at" binding:"required" example:"2024-01-15T10:30:00Z"`
    Type        string    `json:"type" binding:"required" example:"consultation"`
    Reason      string    `json:"reason" binding:"required,max=500" example:"My pet is not feeling well"`
    OwnerNotes  string    `json:"owner_notes" binding:"omitempty,max=1000" example:"Additional information"`
}

// Response DTOs
type AppointmentResponse struct {
    ID             string    `json:"id" example:"507f1f77bcf86cd799439011"`
    PatientID      string    `json:"patient_id" example:"507f1f77bcf86cd799439012"`
    OwnerID        string    `json:"owner_id" example:"507f1f77bcf86cd799439013"`
    VeterinarianID string    `json:"veterinarian_id" example:"507f1f77bcf86cd799439014"`
    ScheduledAt    time.Time `json:"scheduled_at" example:"2024-01-15T10:30:00Z"`
    Duration       int       `json:"duration" example:"30"`
    Type           string    `json:"type" example:"consultation"`
    Status         string    `json:"status" example:"scheduled"`
    Priority       string    `json:"priority" example:"normal"`
    Reason         string    `json:"reason" example:"Annual checkup"`
    Notes          string    `json:"notes,omitempty" example:"First visit"`
    OwnerNotes     string    `json:"owner_notes,omitempty" example:"Patient anxious"`
    ConfirmedAt    *time.Time `json:"confirmed_at,omitempty" example:"2024-01-14T15:00:00Z"`
    StartedAt      *time.Time `json:"started_at,omitempty" example:"2024-01-15T10:35:00Z"`
    CompletedAt    *time.Time `json:"completed_at,omitempty" example:"2024-01-15T11:05:00Z"`
    CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
    CancelReason   string     `json:"cancel_reason,omitempty"`
    CreatedAt      time.Time  `json:"created_at" example:"2024-01-10T14:20:00Z"`
    UpdatedAt      time.Time  `json:"updated_at" example:"2024-01-14T15:00:00Z"`
    
    // Populated data
    Patient      *patients.PatientResponse `json:"patient,omitempty"`
    Owner        *owners.OwnerResponse     `json:"owner,omitempty"`
    Veterinarian *users.UserResponse       `json:"veterinarian,omitempty"`
}

type PaginatedAppointmentsResponse struct {
    Data       []AppointmentResponse `json:"data"`
    Pagination struct {
        Page       int   `json:"page"`
        Limit      int   `json:"limit"`
        Total      int64 `json:"total"`
        TotalPages int   `json:"total_pages"`
    } `json:"pagination"`
}

// Internal DTOs
type appointmentFilters struct {
    Status         []string
    Type           []string
    VeterinarianID *primitive.ObjectID
    PatientID      *primitive.ObjectID
    OwnerID        *primitive.ObjectID
    DateFrom       *time.Time
    DateTo         *time.Time
    Priority       *string
}

// Conversion functions
func (a *Appointment) ToResponse() *AppointmentResponse {
    response := &AppointmentResponse{
        ID:             a.ID.Hex(),
        PatientID:      a.PatientID.Hex(),
        OwnerID:        a.OwnerID.Hex(),
        VeterinarianID: a.VeterinarianID.Hex(),
        ScheduledAt:    a.ScheduledAt,
        Duration:       a.Duration,
        Type:           a.Type,
        Status:         a.Status,
        Priority:       a.Priority,
        Reason:         a.Reason,
        Notes:          a.Notes,
        OwnerNotes:     a.OwnerNotes,
        ConfirmedAt:    a.ConfirmedAt,
        StartedAt:      a.StartedAt,
        CompletedAt:    a.CompletedAt,
        CancelledAt:    a.CancelledAt,
        CancelReason:   a.CancelReason,
        CreatedAt:      a.CreatedAt,
        UpdatedAt:      a.UpdatedAt,
    }
    
    if a.Patient != nil {
        response.Patient = a.Patient.ToResponse()
    }
    if a.Owner != nil {
        response.Owner = a.Owner.ToResponse()
    }
    if a.Veterinarian != nil {
        response.Veterinarian = a.Veterinarian.ToResponse()
    }
    
    return response
}
```

### 3. `repository.go` - Data Access Layer
```go
type AppointmentRepository interface {
    Create(ctx context.Context, appointment *Appointment) error
    FindByID(ctx context.Context, id primitive.ObjectID, tenantIDs []primitive.ObjectID) (*Appointment, error)
    List(ctx context.Context, filters appointmentFilters, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error)
    Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantIDs []primitive.ObjectID) error
    Delete(ctx context.Context, id primitive.ObjectID, tenantIDs []primitive.ObjectID) error
    
    // Business-specific methods
    FindByDateRange(ctx context.Context, from, to time.Time, tenantIDs []primitive.ObjectID) ([]Appointment, error)
    FindByPatient(ctx context.Context, patientID primitive.ObjectID, tenantIDs []primitive.ObjectID) ([]Appointment, error)
    FindByOwner(ctx context.Context, ownerID primitive.ObjectID, tenantIDs []primitive.ObjectID) ([]Appointment, error)
    FindByVeterinarian(ctx context.Context, vetID primitive.ObjectID, from, to time.Time, tenantIDs []primitive.ObjectID) ([]Appointment, error)
    CheckConflicts(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantIDs []primitive.ObjectID) (bool, error)
    
    // Status transitions
    CreateStatusTransition(ctx context.Context, transition *AppointmentStatusTransition) error
    GetStatusHistory(ctx context.Context, appointmentID primitive.ObjectID) ([]AppointmentStatusTransition, error)
}
```

### 4. `service.go` - Business Logic
```go
type Service struct {
    repo        AppointmentRepository
    patientSvc  patients.Service
    ownerSvc    owners.Service
    userSvc     users.Service
    notificationSvc notifications.Service
}

func NewService(repo AppointmentRepository, patientSvc patients.Service, ownerSvc owners.Service, userSvc users.Service, notificationSvc notifications.Service) *Service {
    return &Service{
        repo: repo,
        patientSvc: patientSvc,
        ownerSvc: ownerSvc,
        userSvc: userSvc,
        notificationSvc: notificationSvc,
    }
}

// Core CRUD methods
func (s *Service) CreateAppointment(ctx context.Context, dto CreateAppointmentDTO, tenantIDs []primitive.ObjectID, createdBy primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) GetAppointment(ctx context.Context, id string, tenantIDs []primitive.ObjectID, populate bool) (*AppointmentResponse, error)
func (s *Service) ListAppointments(ctx context.Context, filters appointmentFilters, tenantIDs []primitive.ObjectID, page, limit int, populate bool) (*PaginatedAppointmentsResponse, error)
func (s *Service) UpdateAppointment(ctx context.Context, id string, dto UpdateAppointmentDTO, tenantIDs []primitive.ObjectID, updatedBy primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) DeleteAppointment(ctx context.Context, id string, tenantIDs []primitive.ObjectID, deletedBy primitive.ObjectID) error

// Status management
func (s *Service) UpdateStatus(ctx context.Context, id string, dto UpdateStatusDTO, tenantIDs []primitive.ObjectID, changedBy primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) ConfirmAppointment(ctx context.Context, id string, tenantIDs []primitive.ObjectID, confirmedBy primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) StartAppointment(ctx context.Context, id string, tenantIDs []primitive.ObjectID, startedBy primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) CompleteAppointment(ctx context.Context, id string, notes string, tenantIDs []primitive.ObjectID, completedBy primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) CancelAppointment(ctx context.Context, id string, reason string, tenantIDs []primitive.ObjectID, cancelledBy primitive.ObjectID) (*AppointmentResponse, error)

// Mobile-specific methods
func (s *Service) RequestAppointment(ctx context.Context, dto MobileAppointmentRequestDTO, ownerID primitive.ObjectID, tenantIDs []primitive.ObjectID) (*AppointmentResponse, error)
func (s *Service) GetOwnerAppointments(ctx context.Context, ownerID primitive.ObjectID, tenantIDs []primitive.ObjectID, page, limit int) (*PaginatedAppointmentsResponse, error)

// Business methods
func (s *Service) GetCalendarView(ctx context.Context, from, to time.Time, veterinarianID *string, tenantIDs []primitive.ObjectID) ([]AppointmentResponse, error)
func (s *Service) CheckAvailability(ctx context.Context, vetID string, scheduledAt time.Time, duration int, excludeID *string, tenantIDs []primitive.ObjectID) (bool, error)
func (s *Service) GetStatusHistory(ctx context.Context, id string, tenantIDs []primitive.ObjectID) ([]AppointmentStatusTransition, error)

// Internal validation methods
func (s *Service) validateAppointmentTime(scheduledAt time.Time) error
func (s *Service) validateStatusTransition(currentStatus, newStatus string) error
func (s *Service) scheduleReminders(appointment *Appointment) error
```

### 5. `handler.go` - HTTP Handlers
```go
type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

// Admin/Staff endpoints
// @Summary Create appointment
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) CreateAppointment(c *gin.Context) (any, error)

// @Summary Get appointment
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) GetAppointment(c *gin.Context) (any, error)

// @Summary List appointments
// @Tags admin-appointments  
// @Security BearerAuth
func (h *Handler) ListAppointments(c *gin.Context) (any, error)

// @Summary Update appointment
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) UpdateAppointment(c *gin.Context) (any, error)

// @Summary Delete appointment
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) DeleteAppointment(c *gin.Context) (any, error)

// @Summary Update appointment status
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) UpdateStatus(c *gin.Context) (any, error)

// @Summary Get calendar view
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) GetCalendarView(c *gin.Context) (any, error)

// @Summary Check veterinarian availability
// @Tags admin-appointments
// @Security BearerAuth
func (h *Handler) CheckAvailability(c *gin.Context) (any, error)

// Mobile endpoints
// @Summary Request appointment
// @Tags mobile-appointments
// @Security MobileBearerAuth
func (h *Handler) RequestAppointment(c *gin.Context) (any, error)

// @Summary Get owner appointments
// @Tags mobile-appointments
// @Security MobileBearerAuth
func (h *Handler) GetOwnerAppointments(c *gin.Context) (any, error)

// @Summary Get appointment details
// @Tags mobile-appointments
// @Security MobileBearerAuth
func (h *Handler) GetOwnerAppointment(c *gin.Context) (any, error)
```

### 6. `router.go` - Route Registration
```go
func RegisterAdminRoutes(private *gin.RouterGroup, db *mongo.Database) {
    repo := NewAppointmentRepository(db)
    // Get other services from context or dependency injection
    service := NewService(repo, patientSvc, ownerSvc, userSvc, notificationSvc)
    handler := NewHandler(service)
    
    appointments := private.Group("/appointments")
    {
        appointments.POST("", httpx.Router(handler.CreateAppointment))
        appointments.GET("", httpx.Router(handler.ListAppointments))
        appointments.GET("/calendar", httpx.Router(handler.GetCalendarView))
        appointments.GET("/availability", httpx.Router(handler.CheckAvailability))
        appointments.GET("/:id", httpx.Router(handler.GetAppointment))
        appointments.PUT("/:id", httpx.Router(handler.UpdateAppointment))
        appointments.DELETE("/:id", httpx.Router(handler.DeleteAppointment))
        appointments.PATCH("/:id/status", httpx.Router(handler.UpdateStatus))
    }
}

func RegisterMobileRoutes(mobile *gin.RouterGroup, db *mongo.Database) {
    repo := NewAppointmentRepository(db)
    service := NewService(repo, patientSvc, ownerSvc, userSvc, notificationSvc)
    handler := NewHandler(service)
    
    appointments := mobile.Group("/appointments")
    {
        appointments.POST("/request", httpx.Router(handler.RequestAppointment))
        appointments.GET("", httpx.Router(handler.GetOwnerAppointments))
        appointments.GET("/:id", httpx.Router(handler.GetOwnerAppointment))
    }
}
```

### 7. `errors.go` - Custom Errors
```go
var (
    ErrAppointmentNotFound     = errors.New("appointment not found")
    ErrInvalidAppointmentTime  = errors.New("invalid appointment time")
    ErrAppointmentConflict     = errors.New("appointment time conflicts with existing appointment")
    ErrInvalidStatusTransition = errors.New("invalid status transition")
    ErrAppointmentAlreadyStarted = errors.New("appointment already started")
    ErrAppointmentAlreadyCompleted = errors.New("appointment already completed")
    ErrAppointmentAlreadyCancelled = errors.New("appointment already cancelled")
    ErrPatientNotFound         = errors.New("patient not found for appointment")
    ErrVeterinarianNotFound    = errors.New("veterinarian not found")
    ErrInvalidTimeRange        = errors.New("invalid time range for appointment")
    ErrPastAppointmentTime     = errors.New("cannot schedule appointment in the past")
)
```

## Integration Points

### 1. Database Collections
- `appointments` - Main appointment documents
- `appointment_status_transitions` - Status change history

### 2. MongoDB Indexes
```javascript
// Appointments collection indexes
db.appointments.createIndex({"tenant_ids": 1, "scheduled_at": 1})
db.appointments.createIndex({"tenant_ids": 1, "veterinarian_id": 1, "scheduled_at": 1})
db.appointments.createIndex({"tenant_ids": 1, "patient_id": 1, "scheduled_at": -1})
db.appointments.createIndex({"tenant_ids": 1, "owner_id": 1, "scheduled_at": -1})
db.appointments.createIndex({"tenant_ids": 1, "status": 1, "scheduled_at": 1})
db.appointments.createIndex({"deleted_at": 1}, {sparse: true})

// Status transitions collection indexes  
db.appointment_status_transitions.createIndex({"appointment_id": 1, "created_at": -1})
db.appointment_status_transitions.createIndex({"tenant_ids": 1, "changed_by": 1, "created_at": -1})
```

### 3. Notification Integration
- Appointment created: notify veterinarian and owner
- Appointment confirmed: notify owner  
- Appointment reminder: 24h and 2h before appointment
- Appointment cancelled: notify all parties
- Status changes: notify relevant parties

### 4. Business Rules
- Appointments can only be scheduled during business hours
- Minimum 15 minutes, maximum 8 hours duration
- Cannot schedule in the past (except for admin emergency cases)
- Veterinarian availability checking prevents double-booking
- Status transitions follow defined workflow
- Mobile users can only request appointments, staff must approve
- Automatic cancellation of unconfirmed appointments after 24h

### 5. API Endpoints Summary

#### Admin Routes (/api/appointments)
- `POST /` - Create appointment
- `GET /` - List appointments with filters
- `GET /calendar` - Calendar view
- `GET /availability` - Check availability  
- `GET /:id` - Get appointment details
- `PUT /:id` - Update appointment
- `DELETE /:id` - Delete appointment
- `PATCH /:id/status` - Update status

#### Mobile Routes (/mobile/appointments)
- `POST /request` - Request appointment
- `GET /` - Get owner's appointments
- `GET /:id` - Get appointment details

This architecture follows the exact same patterns as the existing modules while providing comprehensive appointment management functionality for the veterinary platform.