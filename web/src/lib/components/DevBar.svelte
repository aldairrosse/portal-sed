<script lang="ts">
	import { onMount } from 'svelte';
	import { ensureSession, getSession } from '$lib/api/session.svelte';
	import EmployeeSearchSelect from '$lib/components/ui/EmployeeSearchSelect.svelte';
	import { createEmployeePickerStore } from '$lib/stores/employeePickerStore.svelte';

	const session = $derived(getSession());

	let enabled = $state(false);
	let selected = $state('');
	let loading = $state(false);
	let error = $state('');
	let hidden = $state(false);

	const STORAGE_KEY = 'devbar:hidden';
	const picker = createEmployeePickerStore();
	const options = $derived(picker.getOptions());

	$effect(() => {
		if (enabled && session.user) {
			selected = session.user.employeeId;
			const label = (session.user.name ?? '').trim();
			picker.setActive(session.user.employeeId, label);
		}
	});

	onMount(async () => {
		try {
			const v = localStorage.getItem(STORAGE_KEY);
			if (v === '1') hidden = true;
		} catch {
			// ignore storage errors
		}
		try {
			const res = await fetch('/api/v1/dev/status', { credentials: 'include' });
			if (!res.ok) return;
			enabled = true;
		} catch {
			// dev bar disabled
		}
	});

	function toggleHidden() {
		hidden = !hidden;
		try {
			localStorage.setItem(STORAGE_KEY, hidden ? '1' : '0');
		} catch {
			// ignore
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.ctrlKey && e.shiftKey && e.key.toLowerCase() === 'd') {
			if (!enabled) return;
			e.preventDefault();
			toggleHidden();
		}
	}

	async function impersonate() {
		if (!selected) return;
		loading = true;
		error = '';
		try {
			const res = await fetch('/api/v1/dev/impersonate', {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ employee_id: selected })
			});
			if (!res.ok) {
				const b = await res.json().catch(() => ({}));
				throw new Error(b.error ?? 'Error al suplantar');
			}
			await ensureSession();
			location.reload();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Error';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if enabled && session.user && !hidden}
	<div class="fixed bottom-4 right-4 z-50 flex items-center gap-2 bg-warning px-3 py-2 text-sm shadow-xl rounded-box border border-warning/20 max-w-sm">
		<span class="badge badge-sm">DEV</span>
		<EmployeeSearchSelect picker={picker} value={selected} onChange={(v) => (selected = v)} ariaLabel="Empleado a suplantar" class="flex-1 w-64" />
		<button class="btn btn-sm btn-neutral" onclick={impersonate} disabled={loading || !selected}>
			{#if loading}<span class="loading loading-spinner loading-xs"></span>{:else}Suplantar{/if}
		</button>
		{#if error}<span class="text-error text-xs">{error}</span>{/if}
	</div>
{/if}
