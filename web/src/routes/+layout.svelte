<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ensureSession, getSession } from '$lib/api/session.svelte';
	import { loadCycle } from '$lib/api/cycle.svelte';
	import '../app.css';
	import AppShell from '$lib/components/AppShell.svelte';
	import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
	import DevBar from '$lib/components/DevBar.svelte';
	import DevModeSelector from '$lib/components/DevModeSelector.svelte';
	import { initTheme } from '$lib/stores/theme';

	let { children } = $props();

	// Restore persisted theme before first render (default light; no dark flash)
	initTheme();

	// Minimum loader display time (ms) — prevents flash on fast sessions
	const MIN_LOADER_MS = 300;

	const session = $derived(getSession());

	// Pages that render without the app shell and skip the auth guard
	const standaloneRoutes = ['/login', '/logout'];
	let isStandalone = $derived(standaloneRoutes.includes($page.url.pathname));

	// Only true when loading is done AND user exists.
	// AppShell never renders until this is confirmed — zero flash.
	let authResolved = $derived(!session.loading && !!session.user);

	// Dev unauthenticated view: when /api/v1/dev/status 200 allow unauthenticated shell
	let devEnabled = $state(false);
	let devChecked = $state(false);
	let devBypass = $derived(devChecked && devEnabled && !session.loading && !session.user);

	// Minimum loader time elapsed — prevents flash on fast sessions
	let minTimeElapsed = $state(false);

	// True when auth is resolved AND minimum loader time has passed
	let ready = $derived(authResolved && minTimeElapsed);

	// Auth guard: redirect to login if not authenticated
	// Fires only when auth is fully resolved (no user) and loader time passed
	// Bypassed when dev mode is enabled (unauthenticated dev view shows only DevBar + login via impersonate)
	$effect(() => {
		if (session.loading || !devChecked) return;

		// Authenticated user landed on a standalone route (/login) — bounce to home
		// so the demo view never flashes before the redirect.
		if (isStandalone && session.user) {
			goto('/');
			return;
		}

		if (isStandalone) return;
		if (ready) return;
		if (devBypass || devEnabled) return;
		if (!authResolved && minTimeElapsed) {
			// Session resolved with no user, min time passed → redirect
			// Keep the intended destination for the post-SSO redirect (single use).
			if (!sessionStorage.getItem('return_to')) {
				sessionStorage.setItem('return_to', location.pathname + location.search);
			}
			goto('/login');
		}
	});

	onMount(() => {
		ensureSession().then(() => loadCycle());
		fetch('/api/v1/dev/status', { credentials: 'include' })
			.then((r) => {
				devEnabled = r.ok;
			})
			.catch(() => {})
			.finally(() => {
				devChecked = true;
			});

		// Release loader after minimum display time
		const timer = setTimeout(() => {
			minTimeElapsed = true;
		}, MIN_LOADER_MS);

		return () => clearTimeout(timer);
	});
</script>

{#if isStandalone || ready || devBypass}
	{#if isStandalone}
		<!-- Standalone routes (/login) only render once the session is resolved
		     AND the user is not already authenticated. If they are, the
		     $effect above redirects to '/' and the children never mount. -->
		{#if session.user}
			<div class="flex min-h-screen items-center justify-center">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{:else}
			{@render children()}
		{/if}
	{:else if devBypass}
		<DevModeSelector />
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
<DevBar />