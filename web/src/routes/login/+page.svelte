<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { devLogin, getSession } from '$lib/api/session.svelte';

	let pageLoading = $state(true);
	let error = $state<string | null>(null);

	const session = $derived(getSession());

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

		if (!import.meta.env.DEV) {
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
			goto('/');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Error al iniciar sesión';
			pageLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Iniciar sesión — SED</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-base-200">
	<div class="w-full max-w-sm rounded-2xl bg-base-100 p-8 shadow-sm">
		{#if pageLoading}
			<div class="flex flex-col items-center gap-4 py-12">
				<span class="loading loading-spinner loading-lg text-primary"></span>
				<p class="text-sm text-base-content/60">Iniciando sesión con SSO...</p>
			</div>
		{:else if error}
			<div class="flex flex-col items-center gap-4 py-8">
				<div class="alert alert-error">
					<span>{error}</span>
				</div>
				{#if import.meta.env.DEV}
					<button class="btn btn-primary btn-sm" onclick={() => handleDevLogin('dev-rh@empresa.com')}>
						Reintentar con demo
					</button>
				{/if}
			</div>
		{:else}
			<div class="text-center">
				<h1 class="text-2xl font-bold">Portal SED</h1>
				<p class="mt-2 text-sm text-base-content/60">Inicia sesión para continuar</p>

				{#if import.meta.env.DEV}
					<div class="divider">Acceso demo</div>
					<p class="text-xs text-base-content/40 mb-4">Solo disponible en desarrollo</p>

					<div class="flex flex-col gap-2">
						<button
							class="btn btn-outline btn-sm justify-start"
							onclick={() => handleDevLogin('dev-rh@empresa.com')}
						>
							<span class="badge badge-accent badge-xs">RH</span>
							Frankil Perez
						</button>
						<button
							class="btn btn-outline btn-sm justify-start"
							onclick={() => handleDevLogin('dev-jefe@empresa.com')}
						>
							<span class="badge badge-info badge-xs">Jefe</span>
							Juan Carlos
						</button>
						<button
							class="btn btn-outline btn-sm justify-start"
							onclick={() => handleDevLogin('dev-colaborador@empresa.com')}
						>
							<span class="badge badge-ghost badge-xs">Colab.</span>
							Maria Lopez
						</button>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
