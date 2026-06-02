# Testing

## Pirámide

| Nivel | Herramienta | Qué cubrir |
|-------|-------------|------------|
| Unit Go | `testing` + testify | Servicios, reglas de estado, RBAC |
| Unit FE | Vitest | Stores, mappers API, validación forms |
| Integración API | testcontainers PostgreSQL | Repositorios, migraciones |
| Contrato | OpenAPI diff en CI | Breaking changes |
| E2E | Playwright | Login, fijar objetivo, evaluar, permisos |
| Carga | k6 (fase 2) | Listados con cursor bajo volumen |

## CI (objetivo)

1. Lint Go + TS.
2. Tests unitarios.
3. `openspec validate --all`.
4. Build Docker (sin deploy automático hasta pedido).

## Definición de hecho (por change OpenSpec)

- Tests para reglas nuevas.
- OpenAPI actualizado.
- Sin regresión de permisos en casos documentados.

## Skill sugerida

Plantilla de casos de prueba en cada `openspec/changes/*/spec.md` sección **Acceptance**.
