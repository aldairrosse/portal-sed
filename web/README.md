# Web (Vite + Svelte 5 + DaisyUI)

Portal SED — Evaluación de desempeño.

## Comandos

```bash
pnpm dev       # Desarrollo en http://localhost:5173
pnpm build     # Build de producción
pnpm preview   # Previsualizar build de producción
pnpm check     # Type-check + svelte-check
pnpm lint      # ESLint
pnpm format    # Prettier
```

## Stack

- **Framework:** SvelteKit (SPA, `ssr = false`)
- **UI:** DaisyUI 5 + Tailwind CSS 3
- **Lenguaje:** TypeScript 5
- **Paquete:** pnpm

## Notas

- Selector de persona y fase en desarrollo (`import.meta.env.DEV`) — no aparece en producción.
- Tema `sed` DaisyUI con colores primario `#1e3a5f`, secundario `#4a90a4`, acento `#c4a035`.
- `box-shadow` decorativo desactivado globalmente.