# Setup local (futuro)

> **Estado:** carpetas `api/` y `web/` aún no implementadas.

## Pasos previstos

1. Clonar repo y entrar a la raíz.
2. `openspec init --tools cursor --force` si falta integración.
3. Copiar `.env.example` → `.env` (cuando exista).
4. `docker compose up -d db` (cuando exista `docker-compose.yml`).
5. Migraciones: comando documentado en `api/README.md`.
6. Terminal 1: `cd api && air`.
7. Terminal 2: `cd web && pnpm install && pnpm dev`.
8. Abrir URL del frontend (típico `http://localhost:5173`) y API (`http://localhost:8080`).

## Variables de entorno (borrador)

- `DATABASE_URL`
- `SESSION_SECRET`
- `API_PORT`, `CORS_ORIGIN`
- SMTP o proveedor de correo para notificaciones

Definir valores reales en spec OpenSpec `identity-access` y `notifications`.
