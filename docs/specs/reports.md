# Spec: reports

Scope: feature

# Reports Module Specification

## Overview
Estadísticas, informes y dashboards para la gestión de la veterinaria.

## Domain Entities

### Report (schema.go) - Saved reports
- ID, TenantID, Name
- Type (sales, appointments, inventory, etc.)
- DateRange (from, to)
- Filters (JSON)
- CreatedBy
- CreatedAt

---

## Business Logic

### Validaciones

#### Al guardar reporte (SaveReportDTO)
| Campo | Regla |
|-------|-------|
| Name | Requerido |
| Type | Requerido |
| DateFrom | Requerido |
| DateTo | Requerido, debe ser >= DateFrom |
| Filters | Opcional |

### Reglas de Negocio

1. **Rango de fechas**: Máximo 1 año de diferencia
2. **Datos agregados**: Los reportes usan aggregations de MongoDB
3. **Exportación**: PDF o Excel, máximo 10,000 registros
4. **Programación**: Reports pueden enviarse por email daily/weekly/monthly
5. **Dashboard**: Métricas en tiempo real desde cache
6. **Permisos**: Solo admin puede ver reportes

### Eventos del Sistema

| Evento | Cuándo ocurre | Acción |
|--------|---------------|--------|
| `report.generated` | Se genera reporte | Ninguna |
| `report.exported` | Se exporta reporte | Ninguna |
| `report.scheduled` | Se programa reporte | Agregar a scheduler |

### Notificaciones Enviadas

| Tipo | Destinatario | Cuándo |
|------|--------------|--------|
| `report.ready` | Admin (email) | Reporte programado listo |
| `report.dashboard.alert` | Admin | Métrica fuera de umbrales |

---

## Report Types

### Appointments Report
| Métrica | Descripción |
|---------|-------------|
| Total | Total de citas en el rango |
| By Status | Conteo por estado |
| By Veterinarian | Citas por veterinario |
| By Type | Citas por tipo |
| No-show Rate | Porcentaje de no-shows |
| Avg Duration | Duración promedio |

### Sales/Revenue Report
| Métrica | Descripción |
|---------|-------------|
| Total Revenue | Ingreso total |
| By Payment Method | Por método de pago |
| By Service Type | Por tipo de servicio |
| Daily/Weekly/Monthly | Tendencia |
| By Veterinarian | Ingreso por vet |
| Avg Ticket | Ticket promedio |

### Patients Report
| Métrica | Descripción |
|---------|-------------|
| New Patients | Pacientes nuevos |
| Active Patients | Pacientes activos |
| Top Patients | Por número de visitas |
| By Species | Distribución por especie |

### Inventory Report
| Métrica | Descripción |
|---------|-------------|
| Low Stock | Productos con stock bajo |
| Expiring Soon | Por expirar en 30 días |
| Valuation | Valor total del inventario |
| By Category | Por categoría |

---

## Features

### Phase 1: Basic Reports
- [ ] Sales report (by date, payment method)
- [ ] Appointments report (by date, vet, status)
- [ ] Revenue by service type
- [ ] Top patients

### Phase 2: Inventory Reports
- [ ] Products with low stock
- [ ] Expiring products
- [ ] Inventory valuation

### Phase 3: Advanced
- [ ] Custom report builder
- [ ] Export to PDF/Excel
- [ ] Scheduled reports (email)
- [ ] Dashboard metrics
- [ ] Alertas configurables

---

## Integration Points

- **appointments**: Appointment data para reportes
- **patients**: Patient data para reportes
- **payments**: Sales data para reportes
- **inventory**: Stock data para reportes

---

## Endpoints (REST)

### Admin
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | /reports/appointments | Reporte de citas |
| GET | /reports/sales | Reporte de ventas |
| GET | /reports/revenue | Reporte de ingresos |
| GET | /reports/patients | Reporte de pacientes |
| GET | /reports/inventory | Reporte de inventario |
| GET | /reports/dashboard | Métricas del dashboard |
| POST | /reports/saved | Guardar configuración |
| GET | /reports/saved | Listar reportes guardados |
| GET | /reports/saved/:id | Obtener reporte guardado |
| DELETE | /reports/saved/:id | Eliminar reporte |
| POST | /reports/export | Exportar a PDF/Excel |
| POST | /reports/schedule | Programar reporte |

---

## DTOs (Data Transfer Objects)

### ReportQueryDTO
```go
type ReportQueryDTO struct {
    DateFrom string `json:"date_from" binding:"required"` // RFC3339
    DateTo   string `json:"date_to" binding:"required"`   // RFC3339
    // Filtros adicionales según tipo
}
```

### SaveReportDTO
```go
type SaveReportDTO struct {
    Name     string         `json:"name" binding:"required"`
    Type     string        `json:"type" binding:"required"`
    DateFrom string        `json:"date_from" binding:"required"`
    DateTo   string        `json:"date_to" binding:"required"`
    Filters  map[string]interface{} `json:"filters"`
}
```

### DashboardMetrics
```go
type DashboardMetrics struct {
    TodayAppointments    int     `json:"today_appointments"`
    WeekAppointments    int     `json:"week_appointments"`
    MonthAppointments   int     `json:"month_appointments"`
    TodayRevenue        float64 `json:"today_revenue"`
    WeekRevenue         float64 `json:"week_revenue"`
    MonthRevenue        float64 `json:"month_revenue"`
    ActivePatients      int     `json:"active_patients"`
    LowStockProducts    int     `json:"low_stock_products"`
    PendingLabOrders    int     `json:"pending_lab_orders"`
    Hospitalized       int     `json:"hospitalized"`
}
```

---

## Errores Específicos del Módulo

| Código | Descripción |
|--------|-------------|
| DATE_RANGE_INVALID | Rango de fechas inválido |
| DATE_RANGE_TOO_LARGE | Rango máximo de 1 año |
| EXPORT_TOO_LARGE | Máximo 10,000 registros para exportar |
| REPORT_TYPE_INVALID | Tipo de reporte inválido |

---

## Notas de Implementación

- Follow existing module architecture: handler → service → repository → schema
- Multi-tenant isolation
- Usar MongoDB aggregation pipelines
- Cachear dashboard metrics por 5 minutos
- Limitar exportación a 10,000 registros
- Programar reportes usando el scheduler existente
