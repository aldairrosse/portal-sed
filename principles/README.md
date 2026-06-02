# Principles — SED

Decisiones transversales antes y durante la implementación. **No contiene código de aplicación.**

| Documento | Contenido |
|-----------|-----------|
| [architecture.md](./architecture.md) | Capas, despliegue, escalabilidad |
| [contracts-api.md](./contracts-api.md) | OpenAPI, rendimiento FE/BE |
| [data-and-orm.md](./data-and-orm.md) | Modelo, ORM, índices |
| [roles-and-auth.md](./roles-and-auth.md) | RBAC, sesiones, identidad |
| [modules-and-screens.md](./modules-and-screens.md) | Mapa de pantallas y flujos |
| [styles-and-ui.md](./styles-and-ui.md) | DaisyUI, UX, accesibilidad |
| [evaluations-domain.md](./evaluations-domain.md) | Ciclos, objetivos, reglas |
| [security.md](./security.md) | Threat model ligero, controles |
| [testing.md](./testing.md) | Pirámide de pruebas, CI |

Orden sugerido de lectura: architecture → roles-and-auth → evaluations-domain → contracts-api → resto.

Cada decisión fuerte debe reflejarse también en `openspec/specs/` vía OpenSpec.
