---
plan name: laboratory
plan description: Laboratory module phased
plan status: active
---

## Idea
Implementar el módulo de laboratorio en 3 fases: órdenes básicas, tracking, y catálogo

## Implementation
- Fase 1a: Crear schema.go con LabOrder y LabTest
- Fase 1b: Crear repository.go con CRUD
- Fase 1c: Crear service.go y handler.go
- Fase 2a: Tracking de estado (pending→collected→sent→received)
- Fase 2b: Notificaciones cuando resultados estén listos
- Fase 3a: Catálogo de pruebas
- Fase 3b: Precios y reportes

## Required Specs
<!-- SPECS_START -->
- laboratory
<!-- SPECS_END -->