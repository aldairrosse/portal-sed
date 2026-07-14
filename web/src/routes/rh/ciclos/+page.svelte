<script lang="ts">
	import {
		loadCycles,
		getCycles,
		isLoading,
		getError,
		createCycle,
		advancePhase,
		hasCycleForYear,
		assignAll
	} from '$lib/stores/cycleStore.svelte';
	import { API_PHASE_LABELS } from '$lib/types/cycle';
	import type { ApiCyclePhase } from '$lib/types/cycle';
	import { Calendar, Plus, ArrowRight, CheckCircle2, AlertCircle, Users } from '@lucide/svelte';

	const currentYear = new Date().getFullYear();
	const cycles = $derived(getCycles());
	const loading = $derived(isLoading());
	const showCreate = $derived(!hasCycleForYear(currentYear));

	let newYear = $state(currentYear);
	let creating = $state(false);
	let advancingId = $state<string | null>(null);
	let confirmAdvance = $state<{ id: string; toPhase: ApiCyclePhase } | null>(null);
	let localError = $state<string | null>(null);
	let assigningId = $state<string | null>(null);
	let confirmAssign = $state<string | null>(null);
	let assignResult = $state<{ assigned: number; skipped: number; total: number } | null>(null);

	$effect(() => {
		loadCycles();
	});

	function getPrevPhase(current: ApiCyclePhase): ApiCyclePhase | null {
		if (current === 'avance') return 'asignacion';
		if (current === 'cierre') return 'avance';
		return null;
	}

	function getNextPhase(current: ApiCyclePhase): ApiCyclePhase | null {
		if (current === 'asignacion') return 'avance';
		if (current === 'avance') return 'cierre';
		return null;
	}

	function sortedCycles() {
		return [...cycles].sort((a, b) => b.year - a.year);
	}

	function phaseBadgeClass(phase: ApiCyclePhase): string {
		if (phase === 'cierre') return 'badge-success';
		if (phase === 'avance') return 'badge-warning';
		return 'badge-neutral';
	}

	function cycleStatusLabel(cycle: typeof cycles[0]): string {
		if (cycle.finished_at) return 'Cerrado';
		return 'Activo';
	}

	function cycleStatusBadgeClass(cycle: typeof cycles[0]): string {
		if (cycle.finished_at) return 'badge-ghost';
		return 'badge-primary';
	}

	async function handleCreate() {
		creating = true;
		localError = null;
		const result = await createCycle(newYear);
		creating = false;
		if (!result) {
			localError = getError();
		}
	}

	function requestAdvance(cycleId: string, toPhase: ApiCyclePhase) {
		confirmAdvance = { id: cycleId, toPhase };
	}

	function cancelAdvance() {
		confirmAdvance = null;
	}

	async function confirmAdvanceAction() {
		if (!confirmAdvance) return;
		advancingId = confirmAdvance.id;
		localError = null;
		const ok = await advancePhase(confirmAdvance.id, confirmAdvance.toPhase);
		advancingId = null;
		confirmAdvance = null;
		if (!ok) {
			localError = getError();
		}
	}

	function requestAssignAll(cycleId: string) {
		confirmAssign = cycleId;
		assignResult = null;
	}

	function cancelAssign() {
		confirmAssign = null;
		assignResult = null;
	}

	async function confirmAssignAction() {
		if (!confirmAssign) return;
		assigningId = confirmAssign;
		localError = null;
		assignResult = null;
		const result = await assignAll(confirmAssign);
		assigningId = null;
		if (result) {
			assignResult = result;
		} else {
			localError = getError();
			confirmAssign = null;
		}
	}
</script>

<svelte:head>
	<title>Gestión de ciclos — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
				<Calendar class="w-6 h-6" />
				Gestión de ciclos
			</h1>
			<p class="text-sm text-base-content/50 mt-1">
				Administra los ciclos anuales de evaluación de desempeño.
			</p>
		</div>
	</div>

	{#if localError}
		<div class="alert alert-error">
			<AlertCircle class="w-5 h-5" />
			<span>{localError}</span>
		</div>
	{/if}

	{#if loading && cycles.length === 0}
		<div class="flex justify-center py-12">
			<span class="loading loading-spinner loading-lg"></span>
		</div>
	{:else if cycles.length === 0}
		<div class="text-center py-12 text-base-content/40">
			<Calendar class="w-12 h-12 mx-auto mb-3 opacity-30" />
			<p>No hay ciclos registrados.</p>
		</div>
	{:else}
		<div class="flex flex-col gap-3">
			{#each sortedCycles() as cycle (cycle.id)}
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-4">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-3">
								<h2 class="card-title text-base">
									Ciclo {cycle.year}
								</h2>
							</div>
							<span class="badge badge-sm {cycleStatusBadgeClass(cycle)}">
								{cycleStatusLabel(cycle)}
							</span>
						</div>

						<div class="flex items-center gap-2 mt-2 text-xs text-base-content/50">
							<span class="badge badge-sm {phaseBadgeClass(cycle.current_phase)}">
								{API_PHASE_LABELS[cycle.current_phase]}
							</span>
							{#if !cycle.finished_at && getNextPhase(cycle.current_phase)}
								<ArrowRight class="w-3 h-3" />
								<span class="text-base-content/50">
									{API_PHASE_LABELS[getNextPhase(cycle.current_phase)!]}
								</span>
							{/if}
						</div>

						<div class="card-actions justify-end mt-2">
							{#if getPrevPhase(cycle.current_phase)}
								<button
									class="btn btn-ghost btn-sm"
									onclick={() => requestAdvance(cycle.id, getPrevPhase(cycle.current_phase)!)}
									disabled={advancingId === cycle.id}
								>
									{#if advancingId === cycle.id}
										<span class="loading loading-spinner loading-xs"></span>
									{:else}
										Retroceder
									{/if}
								</button>
							{/if}
							{#if !cycle.finished_at && getNextPhase(cycle.current_phase)}
								{#if cycle.current_phase === 'asignacion'}
									<button
										class="btn btn-accent btn-sm"
										onclick={() => requestAssignAll(cycle.id)}
										disabled={assigningId === cycle.id}
									>
										{#if assigningId === cycle.id}
											<span class="loading loading-spinner loading-xs"></span>
										{:else}
											<Users class="w-4 h-4" />
											Asignar todos
										{/if}
									</button>
								{/if}
								<button
									class="btn btn-primary btn-sm"
									onclick={() => requestAdvance(cycle.id, getNextPhase(cycle.current_phase)!)}
									disabled={advancingId === cycle.id}
								>
									{#if advancingId === cycle.id}
										<span class="loading loading-spinner loading-xs"></span>
									{:else}
										Avanzar
									{/if}
								</button>
								{/if}
							</div>

						{#if cycle.finished_at}
							<div class="flex items-center gap-1 mt-1 text-xs text-success/70">
								<CheckCircle2 class="w-3 h-3" />
								Completado
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}

	{#if showCreate}
		<div class="card bg-base-100 border border-base-300 border-dashed">
			<div class="card-body p-4">
				<h3 class="card-title text-sm">
					<Plus class="w-4 h-4" />
					Crear nuevo ciclo
				</h3>
				<div class="flex items-end gap-3 mt-2">
					<div class="form-control">
						<label class="label" for="cycle-year">
							<span class="label-text text-xs">Año</span>
						</label>
						<input
							id="cycle-year"
							type="number"
							class="input input-bordered input-sm w-24"
							bind:value={newYear}
							min={2020}
							max={2099}
						/>
					</div>
					<button
						class="btn btn-primary btn-sm"
						onclick={handleCreate}
						disabled={creating || newYear < 2020}
					>
						{#if creating}
							<span class="loading loading-spinner loading-xs"></span>
						{:else}
							Crear ciclo
						{/if}
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>

{#if confirmAdvance}
	{@const cycle = cycles.find((c) => c.id === confirmAdvance!.id)}
	{@const isAdvance = getNextPhase(cycle?.current_phase ?? 'asignacion') === confirmAdvance.toPhase}
	<dialog class="modal modal-open">
		<div class="modal-box">
			<h3 class="font-bold text-lg">{isAdvance ? 'Avanzar' : 'Retroceder'} fase del ciclo</h3>
			<p class="py-4">
				¿Deseas {isAdvance ? 'avanzar' : 'retroceder'} el ciclo {cycle?.year} de
				<span class="font-medium">{API_PHASE_LABELS[cycle?.current_phase ?? 'asignacion']}</span>
				a
				<span class="font-medium">{API_PHASE_LABELS[confirmAdvance.toPhase]}</span>?
			</p>
			{#if confirmAdvance.toPhase === 'cierre'}
				<div class="alert alert-warning text-sm mb-4">
					<AlertCircle class="w-4 h-4" />
					<span>Al cerrar la fase, finalizará el ciclo de evaluación.</span>
				</div>
			{/if}
			<div class="modal-action">
				<button class="btn btn-ghost" onclick={cancelAdvance}>Cancelar</button>
				<button
					class="btn btn-primary"
					onclick={confirmAdvanceAction}
					disabled={advancingId !== null}
				>
					{#if advancingId}
						<span class="loading loading-spinner loading-xs"></span>
					{/if}
					Confirmar
				</button>
			</div>
		</div>
		<form method="dialog" class="modal-backdrop">
			<button onclick={cancelAdvance}>close</button>
		</form>
	</dialog>
{/if}

{#if confirmAssign}
	<dialog class="modal modal-open">
		<div class="modal-box">
			{#if assignResult}
				<h3 class="font-bold text-lg">Asignación completada</h3>
				<div class="py-4">
					<div class="stat">
						<div class="stat-title">Resultado</div>
						<div class="stat-value text-primary">{assignResult.assigned}</div>
						<div class="stat-desc">empleados asignados</div>
					</div>
					{#if assignResult.skipped > 0}
						<div class="stat">
							<div class="stat-title">Omitidos</div>
							<div class="stat-value text-warning">{assignResult.skipped}</div>
							<div class="stat-desc">ya tenían asignación</div>
						</div>
					{/if}
				</div>
				<div class="modal-action">
					<button class="btn btn-primary" onclick={cancelAssign}>Cerrar</button>
				</div>
			{:else}
				<h3 class="font-bold text-lg">Asignar empleados al ciclo</h3>
				<p class="py-4">
					¿Deseas asignar a todos los empleados activos al ciclo {cycles.find((c) => c.id === confirmAssign)?.year}?
				</p>
				<div class="alert alert-info text-sm mb-4">
					<AlertCircle class="w-4 h-4" />
					<span>Los empleados que ya tengan asignación serán omitidos.</span>
				</div>
				<div class="modal-action">
					<button class="btn btn-ghost" onclick={cancelAssign}>Cancelar</button>
					<button
						class="btn btn-accent"
						onclick={confirmAssignAction}
						disabled={assigningId !== null}
					>
						{#if assigningId}
							<span class="loading loading-spinner loading-xs"></span>
						{/if}
						Asignar todos
					</button>
				</div>
			{/if}
		</div>
		<form method="dialog" class="modal-backdrop">
			<button onclick={cancelAssign}>close</button>
		</form>
	</dialog>
{/if}
