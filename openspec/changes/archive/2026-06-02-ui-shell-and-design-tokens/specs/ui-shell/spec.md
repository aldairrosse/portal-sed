## ADDED Requirements

### Requirement: Layout autenticado simulado

El sistema SHALL renderizar un layout con barra superior (título SED, usuario simulado), menú lateral colapsable en móvil y área de contenido principal en todas las rutas autenticadas.

#### Scenario: Usuario abre la aplicación

- **WHEN** la aplicación carga en cualquier ruta bajo el layout raíz
- **THEN** se muestra barra superior, menú lateral y contenido de la ruta activa

#### Scenario: Vista móvil

- **WHEN** el viewport es menor al breakpoint `lg`
- **THEN** el menú lateral se oculta y SHALL abrirse mediante control hamburguesa accesible

### Requirement: Navegación lazy por ruta

El sistema SHALL cargar cada módulo de ruta de forma diferida (code-splitting) al navegar por primera vez.

#### Scenario: Primera visita a módulo

- **WHEN** el usuario navega a una ruta de módulo no visitada
- **THEN** el sistema carga el chunk de esa ruta sin recargar la página completa

### Requirement: Menú filtrado por perfil de evaluación

El menú lateral SHALL mostrar únicamente ítems cuyo perfil de evaluación activo esté incluido en la configuración del ítem.

#### Scenario: Perfil RH

- **WHEN** el perfil activo es `rh`
- **THEN** se muestran ítems de administración RH y no se muestran ítems exclusivos de colaborador como asignación personal

#### Scenario: Perfil colaborador

- **WHEN** el perfil activo es `colaborador`
- **THEN** se muestran ítems de evaluación propia y asignación y no se muestra matriz 9×9

### Requirement: Página 404

El sistema SHALL mostrar una página de error amigable en español cuando la ruta no existe.

#### Scenario: Ruta inválida

- **WHEN** el usuario navega a una URL no registrada
- **THEN** se muestra página 404 con enlace de regreso al inicio

### Requirement: Componentes de estado UI reutilizables

El sistema SHALL proveer componentes reutilizables para skeleton de carga, estado vacío, error de red y sin permiso (403 simulado).

#### Scenario: Demostración de estados en inicio

- **WHEN** el usuario visita la página de inicio placeholder
- **THEN** puede ver ejemplos o enlaces documentados de cada estado UI para uso en módulos futuros
