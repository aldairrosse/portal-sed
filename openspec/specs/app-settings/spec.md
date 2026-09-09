# app-settings Specification

## Purpose
Expose a hidden settings page that manages theme appearance with a light-only default.

## Requirements

### Requirement: Hidden ajustes page SHALL manage theme with light-only default

The app SHALL expose a hidden `/ajustes` route (no nav entry, inside AppShell auth layout) with an "Apariencia" section bound to a persisted theme store (`light` default, `localStorage`, applied via `document.documentElement.dataset.theme`). `app.html` SHALL NOT auto-apply `prefers-color-scheme`; regular users SHALL see the light (`sed`) theme unless they toggle dark in `/ajustes`. The sidebar SHALL use `--color-sidebar`/`--color-sidebar-content` tokens with contrast-safe states.

#### Scenario: First visit defaults to light

- WHEN a user opens the app for the first time on a system with dark-mode preference
- THEN the app renders the light `sed` theme and persists `light` in the theme store
