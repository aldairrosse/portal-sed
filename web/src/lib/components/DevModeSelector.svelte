<script lang="ts">
	import { ensureSession } from '$lib/api/session.svelte';
	import EmployeeSearchSelect from '$lib/components/ui/EmployeeSearchSelect.svelte';
	import { createEmployeePickerStore } from '$lib/stores/employeePickerStore.svelte';

	let selected = $state('');
	let loading = $state(false);
	let error = $state('');

	const picker = createEmployeePickerStore({ useDev: true });

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

<div class="flex min-h-[60vh] items-center justify-center p-8">
	<div class="card bg-base-100 shadow-xl border border-base-300 w-full max-w-md p-6 flex flex-col gap-4">
		<h2 class="text-lg font-semibold text-center">Modo desarrollo</h2>
		<p class="text-sm text-base-content/60 text-center">Selecciona un empleado para continuar sin sesión</p>
		<EmployeeSearchSelect picker={picker} value={selected} onChange={(v) => (selected = v)} ariaLabel="Empleado a suplantar" class="w-full" />
		<button class="btn btn-primary w-full" onclick={impersonate} disabled={loading || !selected}>
			{#if loading}<span class="loading loading-spinner loading-xs"></span>{:else}Suplantar y entrar{/if}
		</button>
		{#if error}<span class="text-error text-xs text-center">{error}</span>{/if}
	</div>
</div>
