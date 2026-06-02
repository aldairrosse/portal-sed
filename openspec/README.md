# OpenSpec — SED

Estructura base para Spec-Driven Development.

## Inicialización completa

Ejecuta en la raíz del proyecto (requiere Node ≥ 20.19):

```bash
openspec init --tools cursor --force
```

Eso añade/actualiza integración Cursor (comandos `/opsx:*`) y plantillas oficiales.

## Carpetas

- **specs/** — especificaciones duraderas por dominio.
- **changes/** — propuestas en curso (`/opsx:propose`).
- **archive/** — cambios completados.

## Roadmap y decisiones de producto

**[SPEC-ROADMAP.md](./SPEC-ROADMAP.md)** — decisiones cerradas, convenciones, orden UI-first y textos listos para `/opsx:propose`.

Resumen rápido: Fase A pantallas con fixtures → Fase B specs de dominio → Fase C API + auth.

Ver también `docs/get-started/openspec.md`.
