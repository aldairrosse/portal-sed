# Propuesta: C7 — identity-access

## 1. Intent

Implementar el módulo de autenticación y autorización (RBAC) del portal SED.
C7 reemplaza los placeholders `AuthPlaceholder` (marcados como `TODO(auth:C7)`)
en los cambios C2–C6 con middleware real de sesión y control de permisos.

**Especificaciones base:** `principles/roles-and-auth.md`, `principles/security.md`

## 2. Scope

### In Scope

- **Modelo de sesiones:** Tabla `sessions` con token hash, expiración, revocación.
- **Token generation:** SHA-256 hashing of random 32-byte tokens; solo el hash se persiste.
- **Session CRUD:** Creación, validación por token, refresh de expiración, revocación (single y bulk por empleado).
- **RBAC:** 9 roles (`colaborador`, `jefe`, `vendedor`, `gerente-tienda`, `divisional`, `regional`, `director`, `director-general`, `rh`) con permisos granulares (goal, competency, evaluation, cycle, org, admin).
- **Middleware de autenticación:** `RequireAuth` extrae token Bearer (o cookie), valida sesión, inyecta contexto.
- **Middleware de autorización:** `RequirePermission` y `RequireAnyPermission` verifican rol contra permiso requerido.
- **Login handler:** Dev mode (email sin password), con cookie httpOnly + respuesta JSON con token.
- **Logout handler:** Revoca sesión y limpia cookie.
- **Refresh handler:** Extiende expiración de sesión.
- **Me handler:** Retorna información del usuario autenticado.
- **OpenAPI 3.1 spec:** `auth.yaml` con endpoints, schemas, security scheme.
- **Migración 000003:** Creación de tabla `sessions` con índices.

### Out of Scope

- **Password/SSO authentication:** Dev mode solo email. Marcado como `TODO(auth:prod)`.
- **Conexión de middleware en C2–C6 routes:** Se documenta en design.md qué archivos modificar; la conexión real se hace en C8 (wire-api-replace-mocks).
- **Rate limiting en auth endpoints:** Se asume middleware compartido existente.
- **UI de login:** Se implementa en change separado.
- **MFA / 2FA:** Futuro, no en esta versión.

## 3. Dependencies

- **C1: data-model-core** — Tabla `employees` y `evaluation_profiles` existentes.
- **C2–C6:** Placeholders `AuthPlaceholder` a reemplazar (documentado, no modificado aquí).

## 4. Success Criteria

- [ ] Migración 000003 crea tabla `sessions` con índices.
- [ ] `POST /auth/login` retorna token + cookie para email válido.
- [ ] `POST /auth/logout` revoca sesión y limpia cookie.
- [ ] `POST /auth/refresh` extiende expiración.
- [ ] `GET /auth/me` retorna employee + role + profile.
- [ ] `RequireAuth` rechaza requests sin token (401).
- [ ] `RequirePermission` rechaza requests con rol sin permiso (403).
- [ ] Tokens son SHA-256 hasheados antes de almacenar.
- [ ] OpenAPI spec genera tipos TypeScript correctamente.
- [ ] `go build` compila sin errores.
