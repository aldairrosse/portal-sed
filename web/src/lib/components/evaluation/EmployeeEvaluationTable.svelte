<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getEvaluationStatus,
		getCompetencyRatings,
		isLoading,
		getError,
		load as loadEvaluations,
	} from '$lib/stores/evaluationStore.svelte';
	import {
		getPillars,
		getCompetenciesByPillar,
	} from '$lib/stores/competencyStore.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import {
		getGoals,
		getAssignmentByEmployee,
		isLoading as isGoalsLoading,
	} from '$lib/stores/goalsStore.svelte';
	import {
		getAssignmentStatus as getMisAssignmentStatus,
		getItems as getMisItems,
	} from '$lib/stores/misEvaluadosStore.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import EvaluationStatusBadge from './EvaluationStatusBadge.svelte';
	import { PROFILE_LABELS, PHASE_LABELS } from '$lib/types/evaluation';
	import { titleCase } from '$lib/utils/text';
	import { getActivePhase } from '$lib/api/cycle.svelte';
	import { isMedioAnio as isMedioAnioPhase } from '$lib/types/cycle';
	import type { EmployeeAssignment } from '$lib/types/goal';
	import type { EvaluationStatus } from '$lib/types/evaluation-result';
	import type { Snippet } from 'svelte';
	import { load as reloadRhEvaluados } from '$lib/stores/rhEvaluadosStore.svelte';
	import { FileDown, ChevronRight, Pencil } from '@lucide/svelte';
	import { toCsv } from '$lib/utils/export';
	import ChangeDepartmentProfileModal from './ChangeDepartmentProfileModal.svelte';

	// ponytail: remove when OpenAPI schema includes profileName
	interface EmployeeListItemRow {
		id: string;
		firstName: string;
		lastName: string;
		profileName: string;
		isActive: boolean;
		selfAvg?: number | null;
		self_avg?: number | null;
		rhAvg?: number | null;
		rh_avg?: number | null;
		// Status por fase activa desde el backend (null => fallback cliente).
		evaluationStatus?: string | null;
		evaluation_status?: string | null;
		metasStatusFase?: string | null;
		metas_status_fase?: string | null;
		phase?: string | null;
		phaseKind?: string | null;
		phase_kind?: string | null;
		evaluatedCount?: number | null;
	}

	interface Props {
		employees?: EmployeeAssignment[];
		rows?: EmployeeListItemRow[];
		mode?: 'rh' | 'manager';
		onSelect: (employeeId: string) => void;
		selectedEmployeeId?: string;
		disabled?: boolean;
		detail?: Snippet;
		competencyRatings?: Map<
			string,
			{ selfAvg: number; rhAvg: number; status: string }
		>;
		competencyDetailHref?: (employeeId: string) => string;
		// Global RH count (303) — cuando se provee, el chip usa n de total global
		// en lugar de n de página (50). No afecta mis-evaluados (manager).
		globalCompleted?: number | null;
		globalTotal?: number | null;
	}

	let {
		employees = [],
		rows = [],
		onSelect,
		selectedEmployeeId = '',
		disabled = false,
		mode = 'manager',
		detail,
		competencyRatings,
		globalCompleted = null,
		globalTotal = null,
	}: Props = $props();

	const loadingEval = $derived(isLoading());
	const errorEval = $derived(getError());

	onMount(() => {
		loadEvaluations();
	});

	let searchQuery = $state('');
	let changeTargetId = $state<string | null>(null);

	const pillars = $derived(getPillars());
	const allCompetencies = $derived(
		pillars.flatMap((p) => getCompetenciesByPillar(p.id)),
	);
	const goals = $derived(getGoals());

	function getAssignmentStatus(
		employeeId: string,
	): 'no_iniciado' | 'borrador' | 'enviada' {
		// Loaded evaluatees list first (helper defaults missing -> 'no_iniciado').
		if (getMisItems().length > 0) return getMisAssignmentStatus(employeeId);
		// RH fallback only after goals finish loading: never invent from partial data.
		if (!isGoalsLoading()) {
			const g = getAssignmentByEmployee(employeeId)?.status;
			if (g === 'enviada' || g === 'borrador') return g;
			if ((g as string) === '') return 'borrador';
			const legacy = employees.find(
				(x) => x.employeeId === employeeId,
			)?.status;
			if (legacy === 'enviada' || legacy === 'borrador') return legacy;
		}
		return getMisAssignmentStatus(employeeId);
	}

	function assignmentBadge(status: 'no_iniciado' | 'borrador' | 'enviada'): {
		label: string;
		cls: string;
	} {
		if (status === 'enviada') return { label: 'Enviada', cls: 'badge-success' };
		if (status === 'borrador')
			return { label: 'Borrador', cls: 'badge-warning' };
		return { label: 'No iniciado', cls: 'badge-ghost' };
	}

	function isDraftOrBeginning(rowId: string): boolean {
		const s = getAssignmentStatus(rowId);
		return s === 'no_iniciado' || s === 'borrador';
	}

	function isFormulacionPhase(): boolean {
		const p = getActivePhase();
		if (!p) return false;
		const s = p.toLowerCase();
		return (
			s.includes('formul') ||
			s.includes('inicio') ||
			s.includes('planea') ||
			s.includes('asignacion')
		);
	}

	// Single display source: rows (rh/mis-evaluados pass only rows; employees
	// prop stays as goalIds/status sidecar). searchQuery has no input bound.
	const filteredRows = $derived(
		searchQuery.trim() === ''
			? rows
			: rows.filter(
					(r) =>
						`${r.firstName} ${r.lastName}`
							.toLowerCase()
							.includes(searchQuery.toLowerCase()) ||
						r.profileName.toLowerCase().includes(searchQuery.toLowerCase()),
				),
	);

	function goalIdsOf(employeeId: string): string[] {
		return employees.find((e) => e.employeeId === employeeId)?.goalIds ?? [];
	}

	const progressMap = $derived(
		new Map(
			filteredRows.map((row) => {
				const empGoals = goals.filter((g) => goalIdsOf(row.id).includes(g.id));
				const totalTarget = empGoals.reduce((sum, g) => sum + g.targetValue, 0);
				const totalProgress = empGoals.reduce(
					(sum, g) => sum + (g.progress ?? 0),
					0,
				);
				const pct =
					totalTarget > 0
						? Math.min((totalProgress / totalTarget) * 100, 100)
						: null;
				return [row.id, pct] as const;
			}),
		),
	);

	const currentPhase = $derived(getActivePhase() ?? 'inicio-anio');

	// Same derivation as Detail L132-134: medio-anio -> avance, else cierre.
	const phaseKind = $derived<'avance' | 'cierre'>(
		isMedioAnioPhase(currentPhase) ? 'avance' : 'cierre',
	);

	const completionSummary = $derived.by(() => {
		// RH: conteo global (mismo filtro q=, sin paginación) vía meta.completedCount
		if (
			mode === 'rh' &&
			typeof globalTotal === 'number' &&
			typeof globalCompleted === 'number'
		) {
			return { total: globalTotal, completed: globalCompleted };
		}
		return {
			total: filteredRows.length,
			completed: filteredRows.filter((r) => hasCompletedPhase(r.id)).length,
		};
	});

	function hasCompletedPhase(employeeId: string): boolean {
		// DTO primero (fase activa del backend); evita recalcular si hay dato.
		if (dtoStatusOf(employeeId) === 'completed') return true;
		const assignment = employees.find((e) => e.employeeId === employeeId);
		const empGoals = goals.filter((g) => assignment?.goalIds.includes(g.id));

		// Normaliza alias ↔ canónica: inicio-anio ≡ asignacion,
		// medio-anio ≡ avance, fin-anio ≡ cierre.
		const p = currentPhase.toLowerCase().trim().replace(/_/g, '-');
		const isInicio =
			p === 'inicio-anio' ||
			p === 'asignacion' ||
			p.includes('formul') ||
			p.includes('planea');
		if (isInicio) {
			// Completed if has goals assigned
			return assignment !== undefined && assignment.goalIds.length > 0;
		}
		if (isMedioAnioPhase(currentPhase)) {
			// Completed if has updated progress on any goal
			return empGoals.some((g) => g.progress !== undefined && g.progress > 0);
		}
		// avance canónico cae aquí también si isMedioAnioPhase fallara;
		// cierre/fin-anio: mismo criterio que la fila (DTO o cálculo cliente).
		return getStatus(employeeId) === 'completed';
	}

	function getProfileLabel(employeeId: string): string {
		const node = getNodeById(employeeId);
		if (!node) return '—';
		return (
			PROFILE_LABELS[node.profileId as keyof typeof PROFILE_LABELS] ??
			node.profileId
		);
	}

	function dtoStatusOf(employeeId: string): EvaluationStatus | null {
		const row = rows.find((r) => r.id === employeeId);
		const raw = row?.evaluationStatus ?? row?.evaluation_status;
		if (typeof raw !== 'string' || raw.trim() === '') return null;
		const s = raw.toLowerCase();
		return s === 'pending' || s === 'in-progress' || s === 'completed'
			? s
			: null;
	}

	function getStatus(employeeId: string) {
		// Backend primero (fase actual); fallback a cálculo cliente solo si null.
		return (
			dtoStatusOf(employeeId) ??
			getEvaluationStatus(
				employeeId,
				allCompetencies.length,
				goalIdsOf(employeeId),
				phaseKind,
			)
		);
	}

	function metasBadge(row: EmployeeListItemRow): {
		label: string;
		cls: string;
	} | null {
		// DTO primero (fase actual); null => fallback a badge de assignment.
		const raw = row.metasStatusFase ?? row.metas_status_fase;
		if (typeof raw !== 'string' || raw === '') return null;
		if (raw === 'completed') return { label: 'Completadas', cls: 'badge-success' };
		if (raw === 'in-progress')
			return { label: 'En progreso', cls: 'badge-warning' };
		if (raw === 'pending') return { label: 'Pendiente', cls: 'badge-ghost' };
		return { label: 'No iniciado', cls: 'badge-ghost' };
	}

	const statusLabelMap: Record<string, string> = {
		pending: 'Pendiente',
		'in-progress': 'En progreso',
		completed: 'Completada',
	};

	function avgRating(employeeId: string, kind: 'self' | 'rh'): number | null {
		// Backend is phase-aware (detail returns the active-phase evaluation);
		// filter defensively when a rating carries an optional phase field.
		const vals = getCompetencyRatings(employeeId)
			.filter((r) => {
				const p = (r as { phase?: string }).phase;
				return p == null || p === phaseKind;
			})
			.map((r) => (kind === 'self' ? r.selfRating : r.rhRating))
			.filter((v): v is 1 | 2 | 3 | 4 | 5 => typeof v === 'number');
		if (vals.length === 0) return null;
		return vals.reduce((a, b) => a + b, 0) / vals.length;
	}

	function handleExportCsv() {
		toCsv(
			filteredRows.map((row) => ({
				Empleado: `${row.firstName} ${row.lastName}`.trim(),
				Perfil: getProfileLabel(row.id),
				'Progreso global %':
					progressMap.get(row.id) !== null
						? `${Math.round(progressMap.get(row.id)!)}%`
						: '',
				Estado: statusLabelMap[getStatus(row.id)],
			})),
			'evaluaciones.csv',
		);
	}
</script>

<div class="flex flex-col gap-6">
	{#if !selectedEmployeeId}
		<!-- Summary + export -->
		<div class="flex items-center gap-6">
			{#if completionSummary.total > 0}
				<div class="flex items-center gap-2">
					<span class="text-xs font-semibold text-base-content/60"
						>{PHASE_LABELS[currentPhase]}:</span
					>
					<span
						class="badge badge-sm {completionSummary.completed /
							completionSummary.total >=
						0.8
							? 'badge-success'
							: 'badge-warning'}"
					>
						{completionSummary.completed} de {completionSummary.total}
						completaron
					</span>
				</div>
			{/if}

			<button
				class="btn btn-outline btn-sm"
				disabled={filteredRows.length === 0}
				onclick={handleExportCsv}
			>
				<FileDown class="w-4 h-4" />
				Exportar CSV
			</button>
		</div>
	{/if}

	{#if selectedEmployeeId}
		{@render detail?.()}
	{:else if loadingEval}
		<PageSkeleton variant="table" rows={Math.max(rows.length, 3)} />
	{:else if errorEval && rows.length === 0}
		<ErrorState message={errorEval} onretry={loadEvaluations} />
	{:else if filteredRows.length === 0}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			Sin empleados para mostrar
		</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="table table-sm">
				<thead>
					<tr>
						<th class="text-xs font-semibold text-base-content/60">Empleado</th>
						<th class="text-xs font-semibold text-base-content/60">Perfil</th>
						<th class="text-xs font-semibold text-base-content/60 text-center"
							>Metas</th
						>
						<th class="text-xs font-semibold text-base-content/60 text-center"
							>Autoevaluación</th
						>
						<th class="text-xs font-semibold text-base-content/60 text-center"
							>Evaluación</th
						>
						<th class="text-center text-xs font-semibold text-base-content/60"
							>Acciones</th
						>
					</tr>
				</thead>
				<tbody>
					{#each filteredRows as row (row.id)}
						{@const rowStatus = getAssignmentStatus(row.id)}
						{@const _isDraft = isDraftOrBeginning(row.id)}
						{@const _badge = assignmentBadge(rowStatus)}
						{@const _metas = metasBadge(row) ?? _badge}
						{@const selfAvg =
							row.selfAvg ??
							row.self_avg ??
							competencyRatings?.get(row.id)?.selfAvg ??
							avgRating(row.id, 'self')}
						{@const rhAvg =
							row.rhAvg ??
							row.rh_avg ??
							competencyRatings?.get(row.id)?.rhAvg ?? avgRating(row.id, 'rh')}
						<tr class="hover:bg-base-200">
							<td>
								<div class="flex items-center gap-2.5">
									<div class="avatar avatar-placeholder">
										<div
											class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center"
										>
											<span class="text-xs font-bold">
												{row.firstName.charAt(0).toUpperCase()}
											</span>
										</div>
									</div>
									<span class="font-medium text-sm"
										>{row.firstName}
										{row.lastName}</span
									>
									<EvaluationStatusBadge status={getStatus(row.id)} />
								</div>
							</td>
							<td>
								<span class="text-xs text-base-content/50"
									>{titleCase(row.profileName)}</span
								>
							</td>
							<td class="text-center">
								<span class="badge badge-sm {_metas.cls}">
									{_metas.label}
								</span>
							</td>
							<td class="text-center">
								<span
									class="text-sm font-mono {selfAvg != null
										? 'text-base-content'
										: 'text-base-content/30'}"
								>
									{selfAvg != null ? selfAvg.toFixed(2) : '-'}
								</span>
							</td>
							<td class="text-center">
								<span
									class="text-sm font-mono {rhAvg != null
										? 'text-base-content'
										: 'text-base-content/30'}"
								>
									{rhAvg != null ? rhAvg.toFixed(2) : '-'}
								</span>
							</td>
							<td>
								<div class="flex items-center justify-end gap-1">
									{#if mode === 'rh' && isFormulacionPhase()}
										<button
											type="button"
											class="btn btn-outline btn-xs"
											onclick={() => (changeTargetId = row.id)}
										>
											<Pencil class="w-3 h-3" />
											Cambiar
										</button>
									{/if}
									{#if isFormulacionPhase()}
										<a
											href={`/objetivos/asignacion?empId=${row.id}`}
											class="btn btn-outline btn-xs gap-1"
											title="Ver/Editar Metas"
										>
											Metas
										</a>
									{/if}
									<button
										type="button"
										class="btn btn-primary btn-xs"
										onclick={() => onSelect(row.id)}
										{disabled}
									>
										Evaluar
									</button>
									<a
										href={`/evaluacion/9x9/competencias/${row.id}`}
										class="btn btn-ghost btn-square btn-xs"
										aria-label="Ver competencias de {row.firstName} {row.lastName}"
									>
										<ChevronRight class="w-4 h-4" />
									</a>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

{#if changeTargetId}
	<ChangeDepartmentProfileModal
		employeeId={changeTargetId}
		onsave={() => {
			changeTargetId = null;
			loadEvaluations();
			reloadRhEvaluados();
		}}
		onclose={() => {
			changeTargetId = null;
		}}
	/>
{/if}
