<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { devLogin, getSession } from '$lib/api/session.svelte';
	import { AlertCircle } from '@lucide/svelte';
	import type { EvaluationProfile } from '$lib/types/evaluation';

	let pageLoading = $state(false);
	let error = $state<string | null>(null);

	const session = $derived(getSession());

	const showDemo = import.meta.env.VITE_SHOW_DEMO_USERS === 'true';

	// ponytail: hardcoded test profiles, replace with SSO (OIDC/SAML) when ready
	const DEMO_USERS: { email: string; name: string; puesto: string; profileId: EvaluationProfile; badge: string; short: string }[] = [
		{ email: 'alberto@mobo.mx',    name: 'Alberto Cohen',             puesto: 'Director General',                        profileId: 'director-general', badge: 'badge-error',     short: 'SEO' },
		{ email: 'fgarcia@mobo.mx',    name: 'Fernando García Domínguez', puesto: 'Director Ejecutivo de Finanzas y Riesgos', profileId: 'director',         badge: 'badge-warning',   short: 'Director' },
		{ email: 'abraham@mobo.mx',    name: 'Abraham Esses Cohen',       puesto: 'Gerente de Desarrollo e Ingeniería de Datos', profileId: 'jefe',         badge: 'badge-info',      short: 'Jefe' },
		{ email: 'agil@mobo.mx',       name: 'Cristiann Gil Ruíz',        puesto: 'Director de Recursos Humanos',            profileId: 'rh',               badge: 'badge-success',    short: 'RRHH' },
		{ email: 'fperez@mobo.com.mx', name: 'Frankil Aldair Pérez Rosales', puesto: 'Desarrollador Web Jr.',                 profileId: 'colaborador',      badge: 'badge-ghost',     short: 'Colaborador' }
	];

	// Reactively redirect when session becomes available (e.g. after SSR hydration)
	$effect(() => {
		if (!session.loading && session.user) {
			goto('/');
		}
	});

	onMount(() => {
		if (session.user) {
			return; // $effect will handle redirect
		}

		if (!showDemo) {
			// Future: redirect to SSO provider (OIDC/SAML)
			error = 'SSO no configurado aún';
			pageLoading = false;
			return;
		}

		pageLoading = false;
	});

	async function handleDevLogin(email: string) {
		pageLoading = true;
		error = null;
		try {
			await devLogin(email);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Error al iniciar sesión';
			pageLoading = false;
			return;
		}

		// devLogin completó sin throw, pero pudo haber error interno (ensureSession falló en API mode)
		if (session.error) {
			error = session.error;
			pageLoading = false;
			return;
		}

		if (session.user) {
			goto('/');
		} else {
			pageLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Iniciar sesión — SED</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-base-200">
	{#if session.loading || pageLoading}
		<div class="flex flex-col items-center gap-4 py-12">
			<span class="loading loading-spinner loading-lg text-primary"></span>
			<p class="text-sm text-base-content/60">Iniciando sesión con SSO...</p>
		</div>
	{:else if error}
		<div class="flex flex-col items-center gap-3">
			<AlertCircle class="w-8 h-8 text-error" />
			<p class="text-sm text-center max-w-xs">{error}</p>
			<button class="btn btn-outline btn-sm" onclick={() => { error = null; }}>
				Volver
			</button>
		</div>
	{:else}
		<div class="w-full max-w-sm rounded-2xl bg-base-100 p-8 shadow-sm">
			<div class="text-center">
				<h1 class="text-2xl font-bold">Portal SED</h1>
				<p class="mt-2 text-sm text-base-content/60">Inicia sesión para continuar</p>

				{#if showDemo}
					<div class="divider">Acceso demo</div>

					<div class="flex flex-col gap-2 text-left">
						{#each DEMO_USERS as u (u.email)}
							<button
								class="btn btn-outline btn-sm h-auto justify-start py-2"
								onclick={() => handleDevLogin(u.email)}
							>
								<span class="flex w-full flex-col items-start gap-0.5">
									<span class="flex items-center gap-2">
										<span class="font-medium">{u.name}</span>
										<span class="badge {u.badge} badge-xs">{u.short}</span>
									</span>
									<span class="text-xs text-base-content/50 text-left">{u.puesto}</span>
								</span>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
