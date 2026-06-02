# OpenSpec en SED

## Estructura

```
openspec/
├── specs/      # Verdad a largo plazo (por dominio)
├── changes/    # Propuestas activas
└── archive/    # Cambios cerrados
```

## Comandos útiles

```bash
openspec list
openspec validate --all
openspec status
```

## Flujo recomendado para SED

1. **identity-access** — roles, login, RBAC.
2. **evaluation-lifecycle** — estados de ciclo y objetivo.
3. **catalog-objectives** — catálogo.
4. **my-evaluatees**, **my-evaluation**, **objectives-fixation-evaluation**.

En Cursor:

- `/opsx:propose "Autenticación y RBAC para portal SED"`
- Revisar artefactos generados.
- `/opsx:apply` cuando esté aprobado.

## Alineación con `principles/`

Cada spec en `openspec/specs/` debe reflejar decisiones de `principles/*.md`. Si divergen, gana OpenSpec tras revisión explícita.
