# Spec: appointments-testing

Scope: feature

# Appointments Testing & Bug Fixes Specification

## 1. Scheduler Bug Fix (CRITICAL)

### Bug: Double-wrapped `$set` in `processAutoCancellations`

**File:** `internal/scheduler/scheduler.go:119-126`

**Problem:** The code wraps updates in `bson.M{"$set": bson.M{...}}`, but `repository.Update()` at `repository.go:135` already does `bson.M{"$set": updates}`. This produces `{$set: {$set: {...}}}` which silently fails in MongoDB.

**Fix:** Remove the `$set` wrapper from the scheduler. Change:
```go
updates := bson.M{
    "$set": bson.M{
        "status":        "cancelled",
        "cancelled_at":  now,
        "cancel_reason": "Auto-cancelada: no confirmada en 24 horas",
        "updated_at":    now,
    },
}
```
To:
```go
updates := bson.M{
    "status":        "cancelled",
    "cancelled_at":  now,
    "cancel_reason": "Auto-cancelada: no confirmada en 24 horas",
    "updated_at":    now,
}
```

**Also:** Remove stale comment at lines 108-109 ("Necesita un método nuevo en el repo") — `FindUnconfirmedBefore` is already implemented.

### Bug: Swagger annotation mismatch

**File:** `internal/modules/appointments/handler.go:523`

**Problem:** `@Router /mobile/appointments/{id}/cancel [post]` but `router.go:62` registers as `PATCH`.

**Fix:** Change `[post]` to `[patch]` on line 523.

---

## 2. NotificationSender Interface Extraction

### Rationale
`service.go:23` holds `notificationSvc *notifications.Service` (concrete pointer). This prevents mocking in tests without constructing the full notification dependency chain.

### Interface Definition
Create in `service.go` (or a separate `interfaces.go`):
```go
// NotificationSender abstracts notification delivery for testability
type NotificationSender interface {
    Send(ctx context.Context, dto *notifications.SendDTO) error
    SendToStaff(ctx context.Context, dto *notifications.SendStaffDTO) error
}
```

### Changes Required
- `service.go:23`: Change `notificationSvc *notifications.Service` → `notificationSvc NotificationSender`
- `service.go:27` (constructor): Change parameter type `notificationSvc *notifications.Service` → `notificationSvc NotificationSender`
- `router.go`: No change needed — `*notifications.Service` already satisfies the interface
- `scheduler.go:19`: The scheduler also uses `*notifications.Service` directly. It does NOT need to change since we're only extracting the interface for service tests.

---

## 3. Service Unit Tests (`service_test.go`)

**Package:** `package appointments` (same package — required because `appointmentFilters` is unexported)

### Test Dependencies (Mocks)
All mocks are hand-rolled structs implementing the interfaces:

1. `mockAppointmentRepo` — implements `AppointmentRepository`
2. `mockPatientRepo` — implements `patients.PatientRepository`
3. `mockOwnerRepo` — implements `owners.OwnerRepository`
4. `mockUserRepo` — implements `users.UserRepository`
5. `mockNotificationSender` — implements `NotificationSender`

Each mock stores function fields that tests can override per-case.

### Test Cases

#### 3.1 CreateAppointment
- **Happy path**: Valid DTO → creates appointment, returns response, sends 2 notifications (owner + vet)
- **Patient not found**: `patientRepo.FindByID` returns error → returns `ErrPatientNotFound`
- **Vet not found**: `userRepo.FindByID` returns error → returns `ErrVeterinarianNotFound`
- **Time conflict**: `repo.CheckConflicts` returns `true` → returns `ErrAppointmentConflict`
- **Past time**: ScheduledAt in the past → returns `ErrPastAppointmentTime`
- **Outside business hours**: ScheduledAt at 20:00 → returns `ErrInvalidAppointmentTime`
- **Sunday**: ScheduledAt on a Sunday → returns `ErrInvalidAppointmentTime`
- **Invalid patient ID format**: Non-hex string → returns validation error
- **Default priority**: Empty priority → defaults to "normal"

#### 3.2 UpdateStatus
- **Valid transition (scheduled → confirmed)**: Updates status, sets `confirmed_at`, creates transition record, sends notification
- **Valid transition (confirmed → in_progress)**: Updates status, sets `started_at`
- **Valid transition (confirmed → cancelled)**: Updates status, sets `cancelled_at` + `cancel_reason`, sends notification
- **Invalid transition (scheduled → completed)**: Returns `ErrInvalidStatus`
- **Invalid transition (completed → confirmed)**: Returns `ErrInvalidStatus` (terminal state)

#### 3.3 RequestAppointment (Mobile)
- **Happy path**: Valid request → VetID = NilObjectID, Duration = 30, sends staff notification
- **Patient doesn't belong to owner**: patient.OwnerID != ownerID → returns `ErrOwnerMismatch`
- **Patient not found**: Returns `ErrPatientNotFound`

#### 3.4 CancelAppointment (Mobile)
- **Happy path (scheduled → cancelled)**: Validates ownership, updates status, sends staff notification
- **Owner mismatch**: appointment.OwnerID != ownerID → returns `ErrOwnerMismatch`
- **Already cancelled**: Returns current response without error
- **Invalid transition (completed → cancelled)**: Returns `ErrInvalidStatus`

#### 3.5 GetAppointment
- **Without populate**: Returns basic response
- **With populate**: Returns response with Patient, Owner, Vet summaries

#### 3.6 DeleteAppointment
- **Happy path**: Calls repo.Delete
- **Not found**: Returns error from repo

#### 3.7 GetCalendarView
- **Without vet filter**: Uses FindByDateRange
- **With vet filter**: Uses FindByVeterinarian

#### 3.8 CheckAvailability
- **No conflict**: Returns `Available: true`
- **With conflict**: Returns `Available: false` + conflict times

#### 3.9 validateAppointmentTime
- **Valid time**: Weekday, 10:00 → no error
- **Past time**: Returns `ErrPastAppointmentTime`
- **Before 8:00**: Returns `ErrInvalidAppointmentTime`
- **After 18:00**: Returns `ErrInvalidAppointmentTime`
- **Sunday**: Returns `ErrInvalidAppointmentTime`

---

## 4. Handler Unit Tests (`handler_test.go`)

**Package:** `package appointments` (same package for access to unexported types)

### Test Infrastructure
- Use `httptest.NewRecorder()` + `gin.CreateTestContext()`
- Set gin context values for `user_id`, `tenant_id` using `c.Set()`
- Create a `testService` that wraps a `*Service` with mock dependencies

### Test Cases

#### 4.1 CreateAppointment Handler
- **Valid request**: Returns 200 with `AppointmentResponse`
- **Invalid JSON body**: Returns validation error
- **Missing required fields**: Returns validation error with field details
- **Missing tenant ID**: Returns error

#### 4.2 ListAppointments Handler
- **With filters**: Query params correctly parsed into filter map
- **Date parsing**: Invalid RFC3339 → returns validation error
- **Pagination**: skip/limit extracted correctly

#### 4.3 UpdateStatus Handler
- **Valid status**: Returns updated appointment
- **Invalid status value**: Binding validation rejects it

#### 4.4 CancelOwnerAppointment Handler
- **Valid request**: Passes reason to service
- **Missing reason**: Returns validation error (binding:"required")

#### 4.5 GetCalendarView Handler
- **Missing date_from or date_to**: Returns validation error
- **Invalid date format**: Returns validation error

#### 4.6 CheckAvailability Handler
- **Missing required params**: Returns validation error
- **Invalid duration**: Returns validation error

---

## 5. Swagger Regeneration

After fixing the annotation on `handler.go:523`, run:
```bash
make docs
```

This runs `swag init` to regenerate `internal/app/docs/swagger.json`, `swagger.yaml`, and `docs.go`.

---

## 6. Build Verification

After all changes:
```bash
make build
```

This runs `make docs` first (via dependency), then `go build -o tmp/api ./cmd/api`.

---

## 7. Test Execution

```bash
make test
```

Runs `go test ./... -race -count=1`. All tests must pass with zero failures.

---

## 8. Dependencies to Add

```bash
go get github.com/stretchr/testify@latest
```

Use `testify/assert` for assertions and `testify/require` for fatal assertions.