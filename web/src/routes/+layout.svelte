<script lang="ts">
	import { onMount } from 'svelte';
	import { ensureSession } from '$lib/api/session.svelte';
	import '../app.css';
	import AppShell from '$lib/components/AppShell.svelte';

	let { children } = $props();

	onMount(() => {
		ensureSession();
	});
</script>

<AppShell>
	{@render children()}
</AppShell>