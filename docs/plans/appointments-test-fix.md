---
plan name: appointments-test-fix
plan description: Tests, bug fixes, swagger, build
plan status: active
---

## Idea
Completar la implementación del sistema de appointments corrigiendo el bug del scheduler (double-wrapped $set en processAutoCancellations), extrayendo una interface NotificationSender para testabilidad, escribiendo tests unitarios completos del service y handler usando testify, corrigiendo la anotación Swagger incorrecta, regenerando la documentación Swagger, y verificando que el build compile correctamente. El proceso de construcción se detuvo en la sección 7 del spec appointments-fixes (tests) y este plan retoma desde ahí.

## Implementation
- Step 1: Add testify dependency — Run `go get github.com/stretchr/testify@latest` to add the testing assertion library to go.mod/go.sum
- Step 2: Fix scheduler double-wrapped $set bug — Edit `internal/scheduler/scheduler.go:119-126` to remove the `$set` wrapper from the updates bson.M (repo.Update already wraps in $set). Also remove stale comment at lines 108-109 about needing FindUnconfirmedBefore method.
- Step 3: Fix Swagger annotation mismatch — Edit `internal/modules/appointments/handler.go:523` to change `@Router /mobile/appointments/{id}/cancel [post]` to `[patch]` to match the actual route registration in router.go:62.
- Step 4: Extract NotificationSender interface — In `internal/modules/appointments/service.go`, add a `NotificationSender` interface with `Send(ctx, *notifications.SendDTO) error` and `SendToStaff(ctx, *notifications.SendStaffDTO) error`. Change `service.go:23` field type from `*notifications.Service` to `NotificationSender`. Update constructor parameter type accordingly. Verify router.go still compiles (concrete *notifications.Service satisfies the interface).
- Step 5: Create mock definitions for service tests — In `internal/modules/appointments/service_test.go`, create hand-rolled mock structs: mockAppointmentRepo (AppointmentRepository), mockPatientRepo (patients.PatientRepository), mockOwnerRepo (owners.OwnerRepository), mockUserRepo (users.UserRepository), mockNotificationSender (NotificationSender). Each mock uses function fields for per-test overrides.
- Step 6: Write CreateAppointment service tests — Test cases: happy path (valid DTO creates appointment, sends 2 notifications), patient not found, vet not found, time conflict (CheckConflicts returns true), past time rejected, outside business hours (20:00), Sunday rejected, default priority set to 'normal'.
- Step 7: Write UpdateStatus service tests — Test cases: scheduled→confirmed (sets confirmed_at, creates transition, sends notification), confirmed→in_progress (sets started_at), confirmed→cancelled (sets cancelled_at + cancel_reason), invalid transition scheduled→completed (returns ErrInvalidStatus), terminal state completed→confirmed (returns ErrInvalidStatus).
- Step 8: Write RequestAppointment and CancelAppointment service tests — RequestAppointment: happy path (VetID=NilObjectID, Duration=30, staff notification sent), patient not owned by caller (ErrOwnerMismatch), patient not found. CancelAppointment: happy path (validates ownership, updates status), owner mismatch, already cancelled returns current state, invalid transition from completed.
- Step 9: Write remaining service tests — GetAppointment (without/with populate), DeleteAppointment (happy path, not found), GetCalendarView (with/without vet filter), CheckAvailability (no conflict, with conflict), GetStatusHistory, ListAppointments, GetOwnerAppointments, GetOwnerAppointment, validateAppointmentTime edge cases.
- Step 10: Create handler test infrastructure — In `internal/modules/appointments/handler_test.go`, create helper functions: setupTestContext (creates gin.Context with httptest.NewRecorder, sets user_id/tenant_id in context), setupTestHandler (creates Handler with Service backed by mocks).
- Step 11: Write handler unit tests — Test cases: CreateAppointment (valid request 200, invalid JSON, missing required fields), ListAppointments (filter parsing, date validation), UpdateStatus (valid/invalid status), CancelOwnerAppointment (valid with reason, missing reason), GetCalendarView (missing dates, invalid format), CheckAvailability (missing params).
- Step 12: Run tests and fix failures — Execute `make test` (go test ./... -race -count=1). Fix any compilation errors or test failures iteratively until all tests pass.
- Step 13: Regenerate Swagger documentation — Run `make docs` to regenerate swagger.json/yaml/docs.go with the corrected annotation.
- Step 14: Verify full build — Run `make build` to ensure the binary compiles successfully with all changes. Verify no warnings or errors.

## Required Specs
<!-- SPECS_START -->
- appointments-testing
- appointments-fixes
<!-- SPECS_END -->