---
name: figma
description: Integra Figma MCP para DS y rediseño. Lee variables/componentes/estilos, lleva UI a Figma, review y Code Connect.
license: MIT
compatibility: Figma MCP remote https://mcp.figma.com/mcp — requiere fileKey + OAuth. Fallback sin Figma.
metadata:
  author: portal-sed
  version: "1.0"
---

# Figma — Skill

Trigger: `figma`, `llevar a figma`, `variables figma`, `code connect`.

## MCP Detection
1. `figma_whoami` — si falla → modo fallback (avisa, continúa sin Figma, usa proposal como fuente).
2. Si ok → usar MCP para resto.

## Operations
- **Leer DS**: `figma_get_variable_defs` (tokens), `figma_get_libraries` + `figma_search_design_system` (componentes), `figma_get_design_context` (nodo).
- **Llevar UI a Figma**: `figma_search_design_system` primero → `figma_use_figma` (construir desde DS) + `figma_generate_figma_design` (captura pixel-perfect si web app) en paralelo → refinar contra captura → borrar captura.
- **Dashboard inverso (code→Figma)**: ver `.opencode/rules/dashboard-isolation.md` — IDENTIFY DATA → ISOLATE VIEW → CODE→FIGMA con data real; bypass auth solo como `ARCHITECTURE`.
- **Review visual**: `figma_get_screenshot` vs implementación (ver `visual-validation`).
- **Comentarios → requisitos**: leer comentarios Figma → convertir a requisitos, marcar ambiguos como `NEEDS CLARIFICATION`, no implementar ambiguos.
- **Code Connect** preferente: `figma_get_code_connect_map` / `figma_add_code_connect_map` / `figma_send_code_connect_mappings` para mapear nodos ↔ `web/src/lib/components/**`.

## Fallback (sin MCP)
- No bloquea workflow. Documenta: "Figma no disponible — usando proposal local". Mantén tokens en `app.css`.

## Checks
- [ ] MCP detectado o fallback avisado
- [ ] Variables/componentes leídos si MCP ok
- [ ] Ambiguos no implementados
- [ ] Code Connect usado cuando existe mapeo

## Prohibited
- Inventar `fileKey`/`nodeId`
- Pedir `.env` / secrets
- Implementar comentario ambiguo sin clarificación
