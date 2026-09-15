# Contratos API ↔ Frontend

## Fuente de verdad

**OpenAPI 3.1** mantenida junto al backend (`api/openapi.yaml` o generada con `ogen` / `swag`).

## Flujo

1. Spec OpenSpec describe comportamiento y casos límite.
2. OpenAPI describe endpoints, schemas, errores estándar.
3. CI valida OpenAPI y rompe build si el cliente TS no compila.
4. `openapi-typescript` genera tipos; `openapi-fetch` o fetch tipado para llamadas.

## Rendimiento de carga de datos

| Patrón | Uso |
|--------|-----|
| Listado ligero | Solo id, labels, estado, fechas clave |
| Detalle | Segundo request al abrir fila o pantalla |
| Cursor pagination | `?cursor=&limit=` en listas grandes |
| Batch | `POST /v1/batch` solo si spec lo justifica (evitar over-fetch N+1 en FE) |
| ETag / If-None-Match | Catálogos que cambian poco |

## Errores

- Formato único: `code`, `message`, `details[]`, `trace_id`.
- Mapeo en Svelte a toasts y mensajes de formulario.

## Versionado

- Prefijo `/api/v1/`.
- Breaking changes → v2 + periodo de deprecación documentado en OpenSpec.

## Alternativa futura

gRPC + connect-web solo si hay necesidad de streaming masivo; por defecto REST JSON.
