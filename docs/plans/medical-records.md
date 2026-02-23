---
plan name: medical-records
plan description: Medical records module phased
plan status: active
---

## Idea
Implementar el módulo de historial clínico en 3 fases: CRUD básico, timeline, y documentación avanzada

## Implementation
- Fase 1a: Crear schema.go con MedicalRecord, Allergies, MedicalHistory
- Fase 1b: Crear repository.go con CRUD
- Fase 1c: Crear service.go y handler.go
- Fase 2a: Implementar timeline cronológico
- Fase 2b: Agregar filtros por tipo, fecha, veterinario
- Fase 3a: Adjuntos (imágenes, PDFs)
- Fase 3b: Gestión de alergias
- Fase 3c: Condiciones crónicas
- Fase 3d: Exportar a PDF

## Required Specs
<!-- SPECS_START -->
- medical-records
<!-- SPECS_END -->