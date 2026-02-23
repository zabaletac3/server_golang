---
plan name: inventory
plan description: Inventory module phased implementation
plan status: active
---

## Idea
Implementar el módulo de inventario en 3 fases: CRUD básico, gestión de stock, y reportes avanzados

## Implementation
- Fase 1a: Crear schema.go con Product y Category
- Fase 1b: Crear repository.go con CRUD básico + filtros
- Fase 1c: Crear service.go con lógica de negocio
- Fase 1d: Crear handler.go y router.go
- Fase 2a: Implementar stock-in y stock-out
- Fase 2b: Agregar alertas de stock bajo
- Fase 2c: Agregar tracking de expiración
- Fase 3a: Agregar categorías jerárquicas
- Fase 3b: Agregar reportes (stock bajo, por expirar, valoración)
- Fase 3c: Integrar con appointments para consumir productos

## Required Specs
<!-- SPECS_START -->
- inventory
<!-- SPECS_END -->