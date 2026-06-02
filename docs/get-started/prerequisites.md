# Prerrequisitos

## Obligatorio para SDD

- **Node.js** ≥ 20.19 (OpenSpec CLI).
- **Git**.

## Para implementación futura

| Herramienta | Versión orientativa |
|-------------|---------------------|
| Go | 1.22+ |
| pnpm | 9+ (`corepack enable`) |
| Docker | 24+ |
| Docker Compose | v2 |
| PostgreSQL | 15+ (local o contenedor) |
| air | Hot reload Go (`go install github.com/air-verse/air@latest`) |

## OpenSpec global

```bash
npm install -g @fission-ai/openspec@latest
openspec --version
```

En la raíz del proyecto:

```bash
openspec init --tools cursor --force
```

Si ya existe carpeta `openspec/`, usar `openspec update` para refrescar integración Cursor.
