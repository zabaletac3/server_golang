---
plan name: appointments-system
plan description: Sistema gestión citas veterinarias
plan status: active
---

## Idea
Implementar un sistema completo de gestión de citas veterinarias que permita a las clínicas agendar, modificar y gestionar citas, mientras que los dueños de mascotas pueden ver y solicitar citas desde la app móvil. El sistema debe incluir calendario, notificaciones automáticas, gestión de disponibilidad de veterinarios, y estados de citas (agendada, confirmada, en proceso, completada, cancelada). Debe integrarse perfectamente con los módulos existentes de pacientes, owners, usuarios y notificaciones, respetando el multi-tenancy.

## Implementation
- Diseñar el modelo de datos para appointments con campos de fecha/hora, paciente, veterinario, estado, tipo de cita, notas, y duración
- Implementar el dominio appointments con casos de uso para crear, listar, actualizar, cancelar y completar citas
- Crear endpoints REST para el staff: GET /appointments (lista con filtros), POST /appointments (crear), PUT /appointments/:id (actualizar), DELETE /appointments/:id (cancelar)
- Implementar endpoints para owners móvil: GET /mobile/appointments (citas de sus mascotas), POST /mobile/appointments/request (solicitar cita)
- Desarrollar lógica de validación de disponibilidad y prevención de conflictos de horarios
- Integrar con el sistema de notificaciones para recordatorios automáticos 24h y 2h antes de la cita
- Implementar vista de calendario con disponibilidad de veterinarios y filtros por fecha/estado/veterinario
- Crear middleware de autorización tenant-scoped para appointments
- Desarrollar sistema de estados de citas con transiciones válidas y logs de cambios
- Agregar documentación Swagger y testing unitario/integración para todos los endpoints

## Required Specs
<!-- SPECS_START -->
- appointments-architecture
<!-- SPECS_END -->