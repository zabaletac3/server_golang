---
plan name: audit
plan description: Audit module phased
plan status: active
---

## Idea
Implementar el módulo de auditoría en 3 fases: logging básico, retención, y dashboard

## Implementation
- Fase 1a: Crear schema.go con AuditLog
- Fase 1b: Crear repository.go con búsquedas
- Fase 1c: Crear service.go y handler.go
- Fase 2a: Middleware para logging automático
- Fase 2b: Logging de autenticación
- Fase 2c: TTL para auto-cleanup
- Fase 3a: Dashboard de actividad
- Fase 3b: Exportar logs

## Required Specs
<!-- SPECS_START -->
- audit
<!-- SPECS_END -->