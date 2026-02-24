package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Business metrics
	AppointmentsTotal      *prometheus.CounterVec
	AppointmentsDuration   *prometheus.HistogramVec
	TenantsTotal           *prometheus.GaugeVec
	PaymentsTotal          *prometheus.CounterVec
	PaymentsAmountTotal    *prometheus.CounterVec
	RBACChecksTotal        *prometheus.CounterVec
	RBACChecksDeniedTotal  *prometheus.CounterVec
	NotificationsTotal     *prometheus.CounterVec

	// Database metrics
	DBQueryDuration *prometheus.HistogramVec
	DBQueriesTotal  *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// Business metrics
		AppointmentsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "appointments_total",
				Help: "Total number of appointments by status",
			},
			[]string{"status", "type"},
		),
		AppointmentsDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "appointment_duration_seconds",
				Help:    "Appointment duration in seconds",
				Buckets: []float64{300, 600, 900, 1800, 3600, 7200},
			},
			[]string{"type"},
		),
		TenantsTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tenants_total",
				Help: "Total number of tenants by status",
			},
			[]string{"status"},
		),
		PaymentsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payments_total",
				Help: "Total number of payments by provider and status",
			},
			[]string{"provider", "status"},
		),
		PaymentsAmountTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payments_amount_total",
				Help: "Total payment amount by provider",
			},
			[]string{"provider", "currency"},
		),
		RBACChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rbac_checks_total",
				Help: "Total number of RBAC permission checks",
			},
			[]string{"resource", "action", "result"},
		),
		RBACChecksDeniedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rbac_checks_denied_total",
				Help: "Total number of denied RBAC permission checks",
			},
			[]string{"resource", "action"},
		),
		NotificationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notifications_total",
				Help: "Total number of notifications sent by type",
			},
			[]string{"type", "channel"},
		),

		// Database metrics
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"collection", "operation"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"collection", "operation"},
		),
	}

	return m
}

// IncAppointment increments the appointments counter
func (m *Metrics) IncAppointment(status, appointmentType string) {
	m.AppointmentsTotal.WithLabelValues(status, appointmentType).Inc()
}

// ObserveAppointmentDuration observes the duration of an appointment
func (m *Metrics) ObserveAppointmentDuration(appointmentType string, duration float64) {
	m.AppointmentsDuration.WithLabelValues(appointmentType).Observe(duration)
}

// IncPayment increments the payments counter
func (m *Metrics) IncPayment(provider, status string) {
	m.PaymentsTotal.WithLabelValues(provider, status).Inc()
}

// AddPaymentAmount adds to the total payment amount
func (m *Metrics) AddPaymentAmount(provider, currency string, amount float64) {
	m.PaymentsAmountTotal.WithLabelValues(provider, currency).Add(amount)
}

// IncRBACCheck increments the RBAC checks counter
func (m *Metrics) IncRBACCheck(resource, action, result string) {
	m.RBACChecksTotal.WithLabelValues(resource, action, result).Inc()
	if result == "denied" {
		m.RBACChecksDeniedTotal.WithLabelValues(resource, action).Inc()
	}
}

// IncNotification increments the notifications counter
func (m *Metrics) IncNotification(notificationType, channel string) {
	m.NotificationsTotal.WithLabelValues(notificationType, channel).Inc()
}

// ObserveDBQuery observes a database query
func (m *Metrics) ObserveDBQuery(collection, operation string, duration float64) {
	m.DBQueryDuration.WithLabelValues(collection, operation).Observe(duration)
	m.DBQueriesTotal.WithLabelValues(collection, operation).Inc()
}

// IncTenant sets the tenant gauge
func (m *Metrics) SetTenants(status string, count float64) {
	m.TenantsTotal.WithLabelValues(status).Set(count)
}
