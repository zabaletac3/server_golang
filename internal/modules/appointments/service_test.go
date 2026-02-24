package appointments

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/modules/patients"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	testTenantID      = primitive.NewObjectID()
	testUserID        = primitive.NewObjectID()
	testPatientID     = primitive.NewObjectID()
	testOwnerID       = primitive.NewObjectID()
	testVetID         = primitive.NewObjectID()
	testAppointmentID = primitive.NewObjectID()
	testSpeciesID     = primitive.NewObjectID()
)

func getNextMonday10AM() time.Time {
	now := time.Now()
	daysUntilMonday := int(time.Monday) - int(now.Weekday())
	if daysUntilMonday <= 0 {
		daysUntilMonday += 7
	}
	nextMonday := now.AddDate(0, 0, daysUntilMonday)
	return time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 10, 0, 0, 0, time.Local)
}

func getNextSunday() time.Time {
	now := time.Now()
	daysUntilSunday := int(time.Sunday) - int(now.Weekday())
	if daysUntilSunday <= 0 {
		daysUntilSunday += 7
	}
	return now.AddDate(0, 0, daysUntilSunday)
}

type mockAppointmentRepo struct {
	CreateFunc                 func(ctx context.Context, appointment *Appointment) error
	FindByIDFunc               func(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error)
	ListFunc                   func(ctx context.Context, filters appointmentFilters, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error)
	UpdateFunc                 func(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	DeleteFunc                 func(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error
	FindByDateRangeFunc        func(ctx context.Context, from, to time.Time, tenantID primitive.ObjectID) ([]Appointment, error)
	FindByPatientFunc          func(ctx context.Context, patientID primitive.ObjectID, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error)
	FindByOwnerFunc            func(ctx context.Context, ownerID primitive.ObjectID, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error)
	FindByVeterinarianFunc     func(ctx context.Context, vetID primitive.ObjectID, from, to time.Time, tenantID primitive.ObjectID) ([]Appointment, error)
	CheckConflictsFunc         func(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantID primitive.ObjectID) (bool, error)
	CreateStatusTransitionFunc func(ctx context.Context, transition *AppointmentStatusTransition) error
	GetStatusHistoryFunc       func(ctx context.Context, appointmentID primitive.ObjectID) ([]AppointmentStatusTransition, error)
	CountByStatusFunc          func(ctx context.Context, status string, tenantID primitive.ObjectID) (int64, error)
	FindUpcomingFunc           func(ctx context.Context, tenantID primitive.ObjectID, hours int) ([]Appointment, error)
	FindUnconfirmedBeforeFunc  func(ctx context.Context, before time.Time) ([]Appointment, error)
	EnsureIndexesFunc          func(ctx context.Context) error
}

func (m *mockAppointmentRepo) Create(ctx context.Context, appointment *Appointment) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, appointment)
	}
	return nil
}

func (m *mockAppointmentRepo) FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id, tenantID)
	}
	return nil, nil
}

func (m *mockAppointmentRepo) List(ctx context.Context, filters appointmentFilters, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters, tenantID, params)
	}
	return nil, 0, nil
}

func (m *mockAppointmentRepo) Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates, tenantID)
	}
	return nil
}

func (m *mockAppointmentRepo) Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id, tenantID)
	}
	return nil
}

func (m *mockAppointmentRepo) FindByDateRange(ctx context.Context, from, to time.Time, tenantID primitive.ObjectID) ([]Appointment, error) {
	if m.FindByDateRangeFunc != nil {
		return m.FindByDateRangeFunc(ctx, from, to, tenantID)
	}
	return nil, nil
}

func (m *mockAppointmentRepo) FindByPatient(ctx context.Context, patientID primitive.ObjectID, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error) {
	if m.FindByPatientFunc != nil {
		return m.FindByPatientFunc(ctx, patientID, tenantID, params)
	}
	return nil, 0, nil
}

func (m *mockAppointmentRepo) FindByOwner(ctx context.Context, ownerID primitive.ObjectID, tenantID primitive.ObjectID, params pagination.Params) ([]Appointment, int64, error) {
	if m.FindByOwnerFunc != nil {
		return m.FindByOwnerFunc(ctx, ownerID, tenantID, params)
	}
	return nil, 0, nil
}

func (m *mockAppointmentRepo) FindByVeterinarian(ctx context.Context, vetID primitive.ObjectID, from, to time.Time, tenantID primitive.ObjectID) ([]Appointment, error) {
	if m.FindByVeterinarianFunc != nil {
		return m.FindByVeterinarianFunc(ctx, vetID, from, to, tenantID)
	}
	return nil, nil
}

func (m *mockAppointmentRepo) CheckConflicts(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantID primitive.ObjectID) (bool, error) {
	if m.CheckConflictsFunc != nil {
		return m.CheckConflictsFunc(ctx, vetID, scheduledAt, duration, excludeID, tenantID)
	}
	return false, nil
}

func (m *mockAppointmentRepo) CreateStatusTransition(ctx context.Context, transition *AppointmentStatusTransition) error {
	if m.CreateStatusTransitionFunc != nil {
		return m.CreateStatusTransitionFunc(ctx, transition)
	}
	return nil
}

func (m *mockAppointmentRepo) GetStatusHistory(ctx context.Context, appointmentID primitive.ObjectID) ([]AppointmentStatusTransition, error) {
	if m.GetStatusHistoryFunc != nil {
		return m.GetStatusHistoryFunc(ctx, appointmentID)
	}
	return nil, nil
}

func (m *mockAppointmentRepo) CountByStatus(ctx context.Context, status string, tenantID primitive.ObjectID) (int64, error) {
	if m.CountByStatusFunc != nil {
		return m.CountByStatusFunc(ctx, status, tenantID)
	}
	return 0, nil
}

func (m *mockAppointmentRepo) FindUpcoming(ctx context.Context, tenantID primitive.ObjectID, hours int) ([]Appointment, error) {
	if m.FindUpcomingFunc != nil {
		return m.FindUpcomingFunc(ctx, tenantID, hours)
	}
	return nil, nil
}

func (m *mockAppointmentRepo) FindUnconfirmedBefore(ctx context.Context, before time.Time) ([]Appointment, error) {
	if m.FindUnconfirmedBeforeFunc != nil {
		return m.FindUnconfirmedBeforeFunc(ctx, before)
	}
	return nil, nil
}

func (m *mockAppointmentRepo) EnsureIndexes(ctx context.Context) error {
	if m.EnsureIndexesFunc != nil {
		return m.EnsureIndexesFunc(ctx)
	}
	return nil
}

type mockPatientRepo struct {
	CreateFunc      func(ctx context.Context, p *patients.Patient) error
	FindAllFunc     func(ctx context.Context, tenantID primitive.ObjectID, params pagination.Params) ([]patients.Patient, int64, error)
	FindByIDFunc    func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error)
	FindByOwnerFunc func(ctx context.Context, tenantID primitive.ObjectID, ownerID primitive.ObjectID, params pagination.Params) ([]patients.Patient, int64, error)
	UpdateFunc      func(ctx context.Context, tenantID primitive.ObjectID, id string, dto *patients.UpdatePatientDTO) (*patients.Patient, error)
	DeleteFunc      func(ctx context.Context, tenantID primitive.ObjectID, id string) error
}

func (m *mockPatientRepo) Create(ctx context.Context, p *patients.Patient) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, p)
	}
	return nil
}

func (m *mockPatientRepo) FindAll(ctx context.Context, tenantID primitive.ObjectID, params pagination.Params) ([]patients.Patient, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, tenantID, params)
	}
	return nil, 0, nil
}

func (m *mockPatientRepo) FindByID(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, tenantID, id)
	}
	return nil, nil
}

func (m *mockPatientRepo) FindByOwner(ctx context.Context, tenantID primitive.ObjectID, ownerID primitive.ObjectID, params pagination.Params) ([]patients.Patient, int64, error) {
	if m.FindByOwnerFunc != nil {
		return m.FindByOwnerFunc(ctx, tenantID, ownerID, params)
	}
	return nil, 0, nil
}

func (m *mockPatientRepo) Update(ctx context.Context, tenantID primitive.ObjectID, id string, dto *patients.UpdatePatientDTO) (*patients.Patient, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, tenantID, id, dto)
	}
	return nil, nil
}

func (m *mockPatientRepo) Delete(ctx context.Context, tenantID primitive.ObjectID, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, tenantID, id)
	}
	return nil
}

type mockOwnerRepo struct {
	CreateFunc          func(ctx context.Context, dto *owners.CreateOwnerDTO) (*owners.Owner, error)
	FindByIDFunc        func(ctx context.Context, id string) (*owners.Owner, error)
	FindByEmailFunc     func(ctx context.Context, email string) (*owners.Owner, error)
	FindAllFunc         func(ctx context.Context, params pagination.Params) ([]*owners.Owner, int64, error)
	UpdateFunc          func(ctx context.Context, id string, dto *owners.UpdateOwnerDTO) (*owners.Owner, error)
	DeleteFunc          func(ctx context.Context, id string) error
	AddPushTokenFunc    func(ctx context.Context, id string, token owners.PushToken) error
	RemovePushTokenFunc func(ctx context.Context, id string, token string) error
}

func (m *mockOwnerRepo) Create(ctx context.Context, dto *owners.CreateOwnerDTO) (*owners.Owner, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, dto)
	}
	return nil, nil
}

func (m *mockOwnerRepo) FindByID(ctx context.Context, id string) (*owners.Owner, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOwnerRepo) FindByEmail(ctx context.Context, email string) (*owners.Owner, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockOwnerRepo) FindAll(ctx context.Context, params pagination.Params) ([]*owners.Owner, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockOwnerRepo) Update(ctx context.Context, id string, dto *owners.UpdateOwnerDTO) (*owners.Owner, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, dto)
	}
	return nil, nil
}

func (m *mockOwnerRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockOwnerRepo) AddPushToken(ctx context.Context, id string, token owners.PushToken) error {
	if m.AddPushTokenFunc != nil {
		return m.AddPushTokenFunc(ctx, id, token)
	}
	return nil
}

func (m *mockOwnerRepo) RemovePushToken(ctx context.Context, id string, token string) error {
	if m.RemovePushTokenFunc != nil {
		return m.RemovePushTokenFunc(ctx, id, token)
	}
	return nil
}

func (m *mockOwnerRepo) AddTenantID(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	return nil
}

type mockUserRepo struct {
	CreateFunc             func(ctx context.Context, dto *users.CreateUserDTO) (*users.User, error)
	CreateWithPasswordFunc func(ctx context.Context, name, email, hashedPassword string) (*users.User, error)
	CreateUserFunc         func(ctx context.Context, user *users.User) (*users.User, error)
	FindAllFunc            func(ctx context.Context, params pagination.Params) ([]*users.User, int64, error)
	FindByIDFunc           func(ctx context.Context, id string) (*users.User, error)
	FindByEmailFunc        func(ctx context.Context, email string) (*users.User, error)
	UpdateFunc             func(ctx context.Context, id string, dto *users.UpdateUserDTO) (*users.User, error)
	DeleteFunc             func(ctx context.Context, id string) error
}

func (m *mockUserRepo) Create(ctx context.Context, dto *users.CreateUserDTO) (*users.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, dto)
	}
	return nil, nil
}

func (m *mockUserRepo) CreateWithPassword(ctx context.Context, name, email, hashedPassword string) (*users.User, error) {
	if m.CreateWithPasswordFunc != nil {
		return m.CreateWithPasswordFunc(ctx, name, email, hashedPassword)
	}
	return nil, nil
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *users.User) (*users.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil, nil
}

func (m *mockUserRepo) FindAll(ctx context.Context, params pagination.Params) ([]*users.User, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*users.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, id string, dto *users.UpdateUserDTO) (*users.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, dto)
	}
	return nil, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

type mockNotificationSender struct {
	SendFunc        func(ctx context.Context, dto *notifications.SendDTO) error
	SendToStaffFunc func(ctx context.Context, dto *notifications.SendStaffDTO) error
}

func (m *mockNotificationSender) Send(ctx context.Context, dto *notifications.SendDTO) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, dto)
	}
	return nil
}

func (m *mockNotificationSender) SendToStaff(ctx context.Context, dto *notifications.SendStaffDTO) error {
	if m.SendToStaffFunc != nil {
		return m.SendToStaffFunc(ctx, dto)
	}
	return nil
}

func newTestService(repo *mockAppointmentRepo, patientRepo *mockPatientRepo, ownerRepo *mockOwnerRepo, userRepo *mockUserRepo, notifSvc *mockNotificationSender) *Service {
	return &Service{
		repo:            repo,
		patientRepo:     patientRepo,
		ownerRepo:       ownerRepo,
		userRepo:        userRepo,
		notificationSvc: notifSvc,
		cfg:             &config.Config{AppointmentBusinessStartHour: 8, AppointmentBusinessEndHour: 18},
	}
}

func TestCreateAppointment_HappyPath(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patient := &patients.Patient{
		ID:        testPatientID,
		TenantID:  testTenantID,
		OwnerID:   testOwnerID,
		Name:      "Buddy",
		SpeciesID: primitive.NewObjectID(),
		Gender:    patients.GenderMale,
		Active:    true,
	}
	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return patient, nil
	}

	veterinarian := &users.User{
		ID:    testVetID,
		Name:  "Dr. Smith",
		Email: "dr.smith@vet.com",
	}
	userRepo.FindByIDFunc = func(ctx context.Context, id string) (*users.User, error) {
		return veterinarian, nil
	}

	repo.CheckConflictsFunc = func(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantID primitive.ObjectID) (bool, error) {
		return false, nil
	}

	repo.CreateFunc = func(ctx context.Context, appointment *Appointment) error {
		appointment.ID = testAppointmentID
		return nil
	}

	notifSendCalls := 0
	notifSvc.SendFunc = func(ctx context.Context, dto *notifications.SendDTO) error {
		notifSendCalls++
		return nil
	}
	notifSvc.SendToStaffFunc = func(ctx context.Context, dto *notifications.SendStaffDTO) error {
		notifSendCalls++
		return nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    getNextMonday10AM(),
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Priority:       AppointmentPriorityNormal,
		Reason:         "Annual checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, testPatientID.Hex(), resp.PatientID)
	assert.Equal(t, testOwnerID.Hex(), resp.OwnerID)
	assert.Equal(t, testVetID.Hex(), resp.VeterinarianID)
	assert.Equal(t, 2, notifSendCalls)
}

func TestCreateAppointment_PatientNotFound(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return nil, errors.New("not found")
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    getNextMonday10AM(),
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrPatientNotFound, err)
}

func TestCreateAppointment_VetNotFound(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patient := &patients.Patient{
		ID:        testPatientID,
		TenantID:  testTenantID,
		OwnerID:   testOwnerID,
		Name:      "Buddy",
		SpeciesID: primitive.NewObjectID(),
		Gender:    patients.GenderMale,
		Active:    true,
	}
	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return patient, nil
	}

	userRepo.FindByIDFunc = func(ctx context.Context, id string) (*users.User, error) {
		return nil, errors.New("not found")
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    getNextMonday10AM(),
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrVeterinarianNotFound, err)
}

func TestCreateAppointment_TimeConflict(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patient := &patients.Patient{
		ID:        testPatientID,
		TenantID:  testTenantID,
		OwnerID:   testOwnerID,
		Name:      "Buddy",
		SpeciesID: primitive.NewObjectID(),
		Gender:    patients.GenderMale,
		Active:    true,
	}
	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return patient, nil
	}

	veterinarian := &users.User{
		ID:    testVetID,
		Name:  "Dr. Smith",
		Email: "dr.smith@vet.com",
	}
	userRepo.FindByIDFunc = func(ctx context.Context, id string) (*users.User, error) {
		return veterinarian, nil
	}

	repo.CheckConflictsFunc = func(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantID primitive.ObjectID) (bool, error) {
		return true, nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    getNextMonday10AM(),
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrAppointmentConflict, err)
}

func TestCreateAppointment_PastTime(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	pastTime := time.Now().Add(-24 * time.Hour)
	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    pastTime,
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrPastAppointmentTime, err)
}

func TestCreateAppointment_OutsideBusinessHours(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	monday10am := getNextMonday10AM()
	eveningTime := time.Date(monday10am.Year(), monday10am.Month(), monday10am.Day(), 20, 0, 0, 0, time.Local)
	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    eveningTime,
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidAppointmentTime, err)
}

func TestCreateAppointment_Sunday(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return &patients.Patient{
			ID:        testPatientID,
			TenantID:  tenantID,
			OwnerID:   testOwnerID,
			Name:      "Max",
			SpeciesID: testSpeciesID,
		}, nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	nextSunday := getNextSunday()
	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    nextSunday,
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "appointment time")
}

func TestCreateAppointment_DefaultPriority(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patient := &patients.Patient{
		ID:        testPatientID,
		TenantID:  testTenantID,
		OwnerID:   testOwnerID,
		Name:      "Buddy",
		SpeciesID: primitive.NewObjectID(),
		Gender:    patients.GenderMale,
		Active:    true,
	}
	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return patient, nil
	}

	veterinarian := &users.User{
		ID:    testVetID,
		Name:  "Dr. Smith",
		Email: "dr.smith@vet.com",
	}
	userRepo.FindByIDFunc = func(ctx context.Context, id string) (*users.User, error) {
		return veterinarian, nil
	}

	repo.CheckConflictsFunc = func(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantID primitive.ObjectID) (bool, error) {
		return false, nil
	}

	var createdAppointment *Appointment
	repo.CreateFunc = func(ctx context.Context, appointment *Appointment) error {
		appointment.ID = testAppointmentID
		createdAppointment = appointment
		return nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := CreateAppointmentDTO{
		PatientID:      testPatientID.Hex(),
		VeterinarianID: testVetID.Hex(),
		ScheduledAt:    getNextMonday10AM(),
		Duration:       30,
		Type:           AppointmentTypeConsultation,
		Priority:       "",
		Reason:         "Checkup",
	}

	resp, err := svc.CreateAppointment(context.Background(), dto, testTenantID, testUserID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, AppointmentPriorityNormal, createdAppointment.Priority)
}

func TestUpdateStatus_ScheduledToConfirmed(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	currentAppointment := &Appointment{
		ID:          testAppointmentID,
		TenantID:    testTenantID,
		PatientID:   testPatientID,
		OwnerID:     testOwnerID,
		Status:      AppointmentStatusScheduled,
		ScheduledAt: getNextMonday10AM(),
	}
	repo.FindByIDFunc = func(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error) {
		return currentAppointment, nil
	}
	repo.UpdateFunc = func(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
		currentAppointment.Status = updates["status"].(string)
		if confirmedAt, ok := updates["confirmed_at"].(time.Time); ok {
			currentAppointment.ConfirmedAt = &confirmedAt
		}
		return nil
	}
	repo.CreateStatusTransitionFunc = func(ctx context.Context, transition *AppointmentStatusTransition) error {
		return nil
	}
	notifSendCalls := 0
	notifSvc.SendFunc = func(ctx context.Context, dto *notifications.SendDTO) error {
		notifSendCalls++
		return nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := UpdateStatusDTO{
		Status: AppointmentStatusConfirmed,
	}

	resp, err := svc.UpdateStatus(context.Background(), testAppointmentID.Hex(), dto, testTenantID, testUserID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, AppointmentStatusConfirmed, resp.Status)
	assert.NotNil(t, resp.ConfirmedAt)
	assert.Equal(t, 1, notifSendCalls)
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	currentAppointment := &Appointment{
		ID:          testAppointmentID,
		TenantID:    testTenantID,
		PatientID:   testPatientID,
		OwnerID:     testOwnerID,
		Status:      AppointmentStatusScheduled,
		ScheduledAt: getNextMonday10AM(),
	}
	repo.FindByIDFunc = func(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error) {
		return currentAppointment, nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := UpdateStatusDTO{
		Status: AppointmentStatusCompleted,
	}

	resp, err := svc.UpdateStatus(context.Background(), testAppointmentID.Hex(), dto, testTenantID, testUserID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid status transition")
}

func TestRequestAppointment_HappyPath(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	patient := &patients.Patient{
		ID:        testPatientID,
		TenantID:  testTenantID,
		OwnerID:   testOwnerID,
		Name:      "Buddy",
		SpeciesID: primitive.NewObjectID(),
		Gender:    patients.GenderMale,
		Active:    true,
	}
	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return patient, nil
	}

	ownerRepo.FindByIDFunc = func(ctx context.Context, id string) (*owners.Owner, error) {
		return &owners.Owner{
			ID:   testOwnerID,
			Name: "John Doe",
		}, nil
	}

	repo.CreateFunc = func(ctx context.Context, appointment *Appointment) error {
		appointment.ID = testAppointmentID
		return nil
	}

	notifStaffSendCalls := 0
	notifSvc.SendToStaffFunc = func(ctx context.Context, dto *notifications.SendStaffDTO) error {
		notifStaffSendCalls++
		return nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := MobileAppointmentRequestDTO{
		PatientID:   testPatientID.Hex(),
		ScheduledAt: getNextMonday10AM(),
		Type:        AppointmentTypeConsultation,
		Priority:    AppointmentPriorityNormal,
		Reason:      "Annual checkup",
	}

	resp, err := svc.RequestAppointment(context.Background(), dto, testTenantID, testOwnerID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, primitive.NilObjectID.Hex(), resp.VeterinarianID)
	assert.Equal(t, 30, resp.Duration)
	assert.Equal(t, 1, notifStaffSendCalls)
}

func TestRequestAppointment_OwnerMismatch(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	otherOwnerID := primitive.NewObjectID()
	patient := &patients.Patient{
		ID:        testPatientID,
		TenantID:  testTenantID,
		OwnerID:   otherOwnerID,
		Name:      "Buddy",
		SpeciesID: primitive.NewObjectID(),
		Gender:    patients.GenderMale,
		Active:    true,
	}
	patientRepo.FindByIDFunc = func(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error) {
		return patient, nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	dto := MobileAppointmentRequestDTO{
		PatientID:   testPatientID.Hex(),
		ScheduledAt: getNextMonday10AM(),
		Type:        AppointmentTypeConsultation,
		Reason:      "Checkup",
	}

	resp, err := svc.RequestAppointment(context.Background(), dto, testTenantID, testOwnerID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrOwnerMismatch, err)
}

func TestCancelAppointment_HappyPath(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	currentAppointment := &Appointment{
		ID:           testAppointmentID,
		TenantID:     testTenantID,
		PatientID:    testPatientID,
		OwnerID:      testOwnerID,
		Status:       AppointmentStatusScheduled,
		ScheduledAt:  getNextMonday10AM(),
		CancelReason: "",
	}
	repo.FindByIDFunc = func(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error) {
		return currentAppointment, nil
	}
	repo.UpdateFunc = func(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
		currentAppointment.Status = updates["status"].(string)
		if cancelledAt, ok := updates["cancelled_at"].(time.Time); ok {
			currentAppointment.CancelledAt = &cancelledAt
		}
		if reason, ok := updates["cancel_reason"].(string); ok {
			currentAppointment.CancelReason = reason
		}
		return nil
	}
	notifStaffSendCalls := 0
	notifSvc.SendToStaffFunc = func(ctx context.Context, dto *notifications.SendStaffDTO) error {
		notifStaffSendCalls++
		return nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	resp, err := svc.CancelAppointment(context.Background(), testAppointmentID.Hex(), "Ya no necesito la cita", testTenantID, testOwnerID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, AppointmentStatusCancelled, resp.Status)
	assert.Equal(t, "Ya no necesito la cita", resp.CancelReason)
	assert.Equal(t, 1, notifStaffSendCalls)
}

func TestCancelAppointment_OwnerMismatch(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	otherOwnerID := primitive.NewObjectID()
	repo.FindByIDFunc = func(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Appointment, error) {
		return &Appointment{
			ID:        testAppointmentID,
			TenantID:  testTenantID,
			PatientID: testPatientID,
			OwnerID:   otherOwnerID,
			Status:    AppointmentStatusScheduled,
		}, nil
	}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	resp, err := svc.CancelAppointment(context.Background(), testAppointmentID.Hex(), "Reason", testTenantID, testOwnerID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrOwnerMismatch, err)
}

func TestValidateAppointmentTime_Before8AM(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	monday10am := getNextMonday10AM()
	earlyMorning := time.Date(monday10am.Year(), monday10am.Month(), monday10am.Day(), 7, 0, 0, 0, time.Local)
	err := svc.validateAppointmentTime(earlyMorning)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAppointmentTime, err)
}

func TestValidateAppointmentTime_After6PM(t *testing.T) {
	repo := &mockAppointmentRepo{}
	patientRepo := &mockPatientRepo{}
	ownerRepo := &mockOwnerRepo{}
	userRepo := &mockUserRepo{}
	notifSvc := &mockNotificationSender{}

	svc := newTestService(repo, patientRepo, ownerRepo, userRepo, notifSvc)

	monday10am := getNextMonday10AM()
	evening := time.Date(monday10am.Year(), monday10am.Month(), monday10am.Day(), 19, 0, 0, 0, time.Local)
	err := svc.validateAppointmentTime(evening)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAppointmentTime, err)
}
