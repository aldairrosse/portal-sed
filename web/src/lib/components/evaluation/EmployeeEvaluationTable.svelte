<script lang="ts">
	import EvaluationStatusBadge from './EvaluationStatusBadge.svelte';
	import ProgressIndicator from '$lib/components/goals/ProgressIndicator.svelte';
	import { getEvaluationStatus } from '$lib/stores/evaluationStore.svelte';
	import { getPillars, getCompetenciesByPillar } from '$lib/stores/competencyStore.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import { getGoals, getAssignments } from '$lib/stores/goalsStore.svelte';
	import { PROFILE_LABELS, PHASE_LABELS, type CyclePhase } from '$lib/types/evaluation';
	import { getPhase } from '$lib/stores/devContext.svelte';
	import type { EmployeeAssignment } from '$lib/types/goal';
	import type { Snippet } from 'svelte';
	import { toCsv } from '$lib/utils/export';

	interface Props {
		employees: EmployeeAssignment[];
		onSelect: (employeeId: string) => void;
		selectedEmployeeId?: string;
		disabled?: boolean;
		detail?: Snippet;
	}

	let { employees, onSelect, selectedEmployeeId = '', disabled = false, detail }: Props = $props();

	let searchQuery = $state('');

	const pillars = $derived(getPillars());
	const allCompetencies = $derived(pillars.flatMap((p) => getCompetenciesByPillar(p.id)));
	const goals = $derived(getGoals());
	const assignments = $derived(getAssignments());

	const filteredEmployees = $derived(
		searchQuery.trim() === ''
			? employees
			: employees.filter(
					(e) =>
						e.employeeName.toLowerCase().includes(searchQuery.toLowerCase()) ||
						getProfileLabel(e.employeeId).toLowerCase().includes(searchQuery.toLowerCase())
				)
	);

	const progressMap = $derived(new Map(
		filteredEmployees.map((emp) => {
			const empGoals = goals.filter((g) => emp.goalIds.includes(g.id));
			const totalTarget = empGoals.reduce((sum, g) => sum + g.targetValue, 0);
			const totalProgress = empGoals.reduce((sum, g) => sum + (g.progress ?? 0), 0);
			const pct = totalTarget > 0 ? Math.min((totalProgress / totalTarget) * 100, 100) : null;
			return [emp.employeeId, pct] as const;
		})
	));

	const currentPhase = $derived(getPhase());

	const completionSummary = $derived({
		total: filteredEmployees.length,
		completed: filteredEmployees.filter((e) => hasCompletedPhase(e.employeeId)).length
	});

	function hasCompletedPhase(employeeId: string): boolean {
		const assignment = employees.find((e) => e.employeeId === employeeId);
		const empGoals = goals.filter((g) => assignment?.goalIds.includes(g.id));

		switch (currentPhase) {
			case 'inicio-anio':
				// Completed if has goals assigned
				return assignment !== undefined && assignment.goalIds.length > 0;
			case 'medio-anio':
				// Completed if has updated progress on any goal
				return empGoals.some((g) => g.progress !== undefined && g.progress > 0);
			case 'fin-anio':
				// Completed if all competencies rated and all goals closed
				return getStatus(employeeId) === 'completed';
			default:
				return false;
		}
	}

	function getProfileLabel(employeeId: string): string {
		const node = getNodeById(employeeId);
		if (!node) return '—';
		return PROFILE_LABELS[node.profileId as keyof typeof PROFILE_LABELS] ?? node.profileId;
	}

	function getStatus(employeeId: string) {
		const assignment = employees.find((e) => e.employeeId === employeeId);
		return getEvaluationStatus(employeeId, allCompetencies.length, assignment?.goalIds ?? []);
	}

	const statusLabelMap: Record<string, string> = {
		pending: 'Pendiente',
		'in-progress': 'En progreso',
		completed: 'Completada'
	};

	function handleExportCsv() {
		toCsv(
			filteredEmployees.map((emp) => ({
				Empleado: emp.employeeName,
				Perfil: getProfileLabel(emp.employeeId),
				'Progreso global %':
					progressMap.get(emp.employeeId) !== null
						? `${Math.round(progressMap.get(emp.employeeId)!)}%`
						: '',
				Estado: statusLabelMap[getStatus(emp.employeeId)]
			})),
			'evaluaciones.csv'
		);
	}
</script>

<div class="flex flex-col gap-6">
	{#if !selectedEmployeeId}
		{#if completionSummary.total > 0}
			<div class="flex items-center gap-2">
				<span class="text-xs font-semibold text-base-content/60">{PHASE_LABELS[currentPhase]}:</span>
				<span class="badge badge-sm {completionSummary.completed / completionSummary.total >= 0.8 ? 'badge-success' : 'badge-warning'}">
					{completionSummary.completed} de {completionSummary.total} completaron
				</span>
			</div>
		{/if}
		<!-- Search input + export -->
		<div class="flex items-center gap-2">
			<div class="w-full max-w-sm">
				<input
					id="employee-search"
					type="text"
					class="input input-bordered input-sm w-full"
					placeholder="Buscar por nombre o perfil..."
					bind:value={searchQuery}
					aria-label="Buscar empleado"
				/>
			</div>
			<button
				class="btn btn-outline btn-sm"
				disabled={filteredEmployees.length === 0}
				onclick={handleExportCsv}
			>
				Exportar CSV
			</button>
		</div>
	{/if}

	{#if selectedEmployeeId}
		{@render detail?.()}
	{:else if filteredEmployees.length === 0}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			No se encontraron empleados.
		</p>
	{:else}
		<!-- Table -->
		<div class="overflow-x-auto">
			<table class="table table-sm">
				<thead>
					<tr>
						<th class="text-xs font-semibold text-base-content/60">Empleado</th>
						<th class="text-xs font-semibold text-base-content/60">Perfil</th>
						<th class="text-xs font-semibold text-base-content/60">Progreso global</th>
						<th class="text-xs font-semibold text-base-content/60">Estado</th>
						<th class="text-xs font-semibold text-base-content/60">Acción</th>
					</tr>
				</thead>
				<tbody>
					{#each filteredEmployees as emp (emp.employeeId)}
						<tr class="hover:bg-base-200">
							<td>
								<div class="flex items-center gap-2.5">
									<div class="avatar avatar-placeholder">
										<div class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center">
											<span class="text-xs font-bold">
												{emp.employeeName.charAt(0).toUpperCase()}
											</span>
										</div>
									</div>
									<span class="font-medium text-sm">{emp.employeeName}</span>
								</div>
							</td>
							<td>
								<span class="text-xs text-base-content/50">{getProfileLabel(emp.employeeId)}</span>
							</td>
							<td>
								{#if progressMap.get(emp.employeeId) !== null}
									<ProgressIndicator value={progressMap.get(emp.employeeId)!} />
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td>
								<EvaluationStatusBadge status={getStatus(emp.employeeId)} />
							</td>
							<td>
							<button
								type="button"
								class="btn btn-primary btn-xs"
								onclick={() => onSelect(emp.employeeId)}
								disabled={disabled}
							>
									Evaluar
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
