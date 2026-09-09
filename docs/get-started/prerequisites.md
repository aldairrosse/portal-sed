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
mise use --global openspec@1.12.0
openspec --version
```

Nota: `mise upgrade openspec` solo si falta o la versión instalada difiere.
