<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ensureSession, getSession } from '$lib/api/session.svelte';
	import '../app.css';
	import AppShell from '$lib/components/AppShell.svelte';
	import ToastContainer from '$lib/components/ui/ToastContainer.svelte';

	let { children } = $props();

	// Minimum loader display time (ms) — prevents flash on fast sessions
	const MIN_LOADER_MS = 300;

	const session = $derived(getSession());

	// Pages that render without the app shell and skip the auth guard
	const standaloneRoutes = ['/login', '/logout'];
	let isStandalone = $derived(standaloneRoutes.includes($page.url.pathname));

	// Only true when loading is done AND user exists.
	// AppShell never renders until this is confirmed — zero flash.
	let authResolved = $derived(!session.loading && !!session.user);

	// Minimum loader time elapsed — prevents flash on fast sessions
	let minTimeElapsed = $state(false);

	// True when auth is resolved AND minimum loader time has passed
	let ready = $derived(authResolved && minTimeElapsed);

	// Auth guard: redirect to login if not authenticated
	// Fires only when auth is fully resolved (no user) and loader time passed
	$effect(() => {
		if (session.loading) return;

		// Authenticated user landed on a standalone route (/login) — bounce to home
		// so the demo view never flashes before the redirect.
		if (isStandalone && session.user) {
			goto('/');
			return;
		}

		if (isStandalone) return;
		if (ready) return;
		if (!authResolved && minTimeElapsed) {
			// Session resolved with no user, min time passed → redirect
			goto('/login');
		}
	});

	onMount(() => {
		ensureSession();

		// Release loader after minimum display time
		const timer = setTimeout(() => {
			minTimeElapsed = true;
		}, MIN_LOADER_MS);

		return () => clearTimeout(timer);
	});
</script>

{#if isStandalone || ready}
	{#if isStandalone}
		<!-- Standalone routes (/login) only render once the session is resolved
		     AND the user is not already authenticated. If they are, the
		     $effect above redirects to '/' and the children never mount. -->
		{#if session.loading || session.user}
			<div class="flex min-h-screen items-center justify-center">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{:else}
			{@render children()}
		{/if}
	{:else}
		<AppShell>
			{@render children()}
		</AppShell>
	{/if}
{:else}
	<div class="flex min-h-screen items-center justify-center">
		<span class="loading loading-spinner loading-lg"></span>
	</div>
{/if}

<ToastContainer />