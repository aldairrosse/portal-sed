// Package dev provides development-only auth helpers mounted at /api/v1/dev.
// Handlers guard with 404 in production; main.go also conditionally mounts.
// This file intentionally has no preset users — dev uses live DB employees
// via /dev/employees + impersonate. Deletable when dev layer is removed.
package dev
