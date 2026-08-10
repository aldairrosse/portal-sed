<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { getSession } from '$lib/api/session.svelte';
	import { AlertCircle } from '@lucide/svelte';

	let pageLoading = $state(true);
	let error = $state<string | null>(null);

	const session = $derived(getSession());

	const ERROR_MAP: Record<string, string> = {
		'usuario_no_encontrado': 'SSO autenticado pero tu usuario no está registrado en SED. Contacta a RH.',
		'sin_acceso': 'No tienes acceso a este sistema. Solicita acceso en la consola SSO.',
		'usuario_inactivo': 'Tu cuenta está inactiva. Contacta a RH.',
		'token_exchange_fallido': 'Error de comunicación con el SSO. Intenta de nuevo.',
		'id_token_invalido': 'Error de seguridad en la autenticación SSO.',
		'parametros_invalidos': 'Error en la respuesta del SSO. Intenta de nuevo.',
		'error_bd': 'Error interno. Contacta a soporte.',
		'error_sesion': 'Error al crear la sesión. Intenta de nuevo.',
		'error_usuario': 'Error al obtener tus datos del SSO.',
		'estado_invalido': 'La solicitud de autenticación caducó o fue reutilizada. Intenta de nuevo.',
		'stepup_fallido': 'No se pudo completar la verificación en dos pasos. Intenta de nuevo.',
		'stepup_error': 'Error al iniciar la verificación en dos pasos. Intenta de nuevo.',
	};

	$effect(() => {
		if (!session.loading && session.user) {
			goto('/');
		}
	});

	onMount(() => {
		if (session.user) return;

		const params = new URLSearchParams(window.location.search);
		const ssoError = params.get('sso_error');
		if (ssoError) {
			error = ERROR_MAP[ssoError] ?? 'Error de autenticación desconocido.';
			window.history.replaceState({}, '', '/login');
			pageLoading = false;
			return;
		}

		// No session, no error → redirect to SSO
		const count = Number(sessionStorage.getItem('sso_redirect_count')) || 0;
		if (count >= 3) {
			error = 'No se pudo completar la autenticación con el SSO. Contacta a soporte.';
			pageLoading = false;
		} else {
			sessionStorage.setItem('sso_redirect_count', String(count + 1));
			const returnTo = sessionStorage.getItem('return_to') || '/';
			sessionStorage.removeItem('return_to');
			console.warn('[sso] redirect a sso-login, intento', count + 1);
			window.location.href = '/api/v1/auth/sso-login?return_to=' + encodeURIComponent(returnTo);
		}
	});
</script>

<svelte:head>
	<title>Iniciar sesión — SED</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-base-200">
	{#if pageLoading || session.loading}
		<div class="flex flex-col items-center gap-4 py-12">
			<span class="loading loading-spinner loading-lg text-primary"></span>
			<p class="text-sm text-base-content/60">Redirigiendo al SSO...</p>
		</div>
	{:else if error}
		<div class="flex flex-col items-center gap-3 max-w-md px-4">
			<AlertCircle class="w-8 h-8 text-error" />
			<p class="text-sm text-center">{error}</p>
			<a href="/api/v1/auth/sso-login" class="btn btn-outline btn-sm mt-2">
				Intentar de nuevo
			</a>
		</div>
	{/if}
</div>
