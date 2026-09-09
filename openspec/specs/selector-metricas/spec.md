# selector-metricas Specification

## Purpose
Provide a paginated hierarchical org selector and phase-filtered RH metrics for teams and managers.

## Requirements

### Requirement: Selector jerárquico paginado

El formulario de meta global SHALL ofrecer un selector anidado paginado que navega el árbol org por ramas ordenadas por `path`, en lugar de un buscador plano.

#### Scenario: Navegar rama paginada

- **WHEN** RH abre el selector en el nodo raíz con más de 50 hijos
- **THEN** ve la primera página de 50 hijos ordenados por jerarquía con control de paginación y breadcrumb

### Requirement: Métricas RH por fase con equipo y jefe

`metrics_service.go` SHALL computar métricas filtradas por `phase` (default `cycle.current_phase`), separando agregado del equipo y fila del jefe.

#### Scenario: Métricas de avance

- **WHEN** RH pide métricas con `phase=avance`
- **THEN** los totales reflejan solo datos de avance, con tabla de equipo y fila de jefe separadas

#### Scenario: Fase omitida

- **WHEN** RH pide métricas sin `phase`
- **THEN** el sistema usa `cycle.current_phase`
