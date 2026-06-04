<script lang="ts">
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import ForbiddenState from '$lib/components/ui/ForbiddenState.svelte';
	import { getProfile, getPhase } from '$lib/stores/devContext.svelte';
	import { PROFILE_LABELS, PHASE_LABELS } from '$lib/types/evaluation';
	import { PROFILE_USERS } from '$lib/dev/profileUsers';

	const profile = $derived(getProfile());
	const phase = $derived(getPhase());
	const user = $derived(PROFILE_USERS[profile]);
	const profileLabel = $derived(PROFILE_LABELS[profile]);
	const phaseLabel = $derived(PHASE_LABELS[phase]);
</script>

<svelte:head>
	<title>Inicio — SED</title>
</svelte:head>

<div class="space-y-8 max-w-4xl mx-auto">
	<!-- Welcome: plain text, no container -->
	<div>
		<h1 class="text-xl font-semibold text-base-content">Portal SED</h1>
		<p class="text-base-content/50 mt-1">Bienvenido al portal de evaluación de desempeño.</p>
	</div>

	<!-- Dev context: emphasis with bg color + padding -->
	<div class="bg-base-200/60 rounded-xl p-4">
		<h2 class="text-sm font-semibold text-base-content mb-3">Contexto de desarrollo activo</h2>
		<dl class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
			<div class="bg-base-100 rounded-lg p-3">
				<dt class="text-base-content/40 text-xs uppercase tracking-wider font-medium">Perfil</dt>
				<dd class="font-medium text-primary mt-1">{profileLabel}</dd>
				<dd class="text-base-content/50 text-xs mt-0.5">{user.name}</dd>
			</div>
			<div class="bg-base-100 rounded-lg p-3">
				<dt class="text-base-content/40 text-xs uppercase tracking-wider font-medium">Fase del ciclo</dt>
				<dd class="font-medium text-secondary mt-1">{phaseLabel}</dd>
			</div>
		</dl>
		<p class="text-xs text-base-content/30 mt-3">
			Este contexto solo está disponible en modo desarrollo y se almacena en sessionStorage.
		</p>
	</div>

	<!-- UI states: plain section, no container -->
	<div>
		<h2 class="text-sm font-semibold text-base-content mb-4">Estados de UI disponibles</h2>
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
			<div class="bg-base-200/40 rounded-xl p-4">
				<h3 class="font-medium text-xs text-base-content/50 uppercase tracking-wider mb-2">Skeleton (carga)</h3>
				<PageSkeleton rows={3} />
			</div>
			<div class="bg-base-200/40 rounded-xl p-4">
				<h3 class="font-medium text-xs text-base-content/50 uppercase tracking-wider mb-2">Empty (sin datos)</h3>
				<EmptyState title="Sin evaluaciones" message="No hay evaluaciones asignadas en este momento." />
			</div>
			<div class="bg-base-200/40 rounded-xl p-4">
				<h3 class="font-medium text-xs text-base-content/50 uppercase tracking-wider mb-2">Error (red)</h3>
				<ErrorState />
			</div>
			<div class="bg-base-200/40 rounded-xl p-4">
				<h3 class="font-medium text-xs text-base-content/50 uppercase tracking-wider mb-2">Forbidden (403)</h3>
				<ForbiddenState />
			</div>
		</div>
		<p class="text-xs text-base-content/30 mt-3">
			Importa estos componentes desde <code class="bg-base-200 px-1.5 py-0.5 rounded-md text-xs">$lib/components/ui/</code>
		</p>
	</div>
</div>