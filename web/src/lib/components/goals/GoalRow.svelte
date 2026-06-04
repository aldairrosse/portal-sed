<script lang="ts">
	import { Pencil, Trash2, MessageSquare, MessageCircle } from '@lucide/svelte';
	import type { Goal, KPI, CyclePhase } from '$lib/types/goal';
	import KpiBadge from './KpiBadge.svelte';
	import ProgressIndicator from './ProgressIndicator.svelte';

	interface Props {
		goal: Goal;
		kpis: KPI[];
		onEdit: (goal: Goal) => void;
		onDelete: (goalId: string) => void;
		mode?: 'editor' | 'reader';
		onRequestChange?: (goal: Goal) => void;
		phase?: CyclePhase;
		canEditProgress?: boolean;
		canComment?: boolean;
		canDelete?: boolean;
		onUpdateProgress?: (goalId: string, progress: number) => void;
		onOpenComments?: (goal: Goal) => void;
	}

	let {
		goal,
		kpis,
		onEdit,
		onDelete,
		mode = 'editor',
		onRequestChange,
		phase = 'asignacion',
		canEditProgress = false,
		canComment = false,
		canDelete = true,
		onUpdateProgress,
		onOpenComments
	}: Props = $props();

	let progressValue = $state(0);

	// Sync when goal prop changes
	$effect(() => {
		progressValue = goal.progress ?? 0;
	});

	const unitLabels: Record<string, string> = {
		porcentaje: '%',
		moneda: '$',
		numero: '#',
		binario: 'Sí/No'
	};

	function handleProgressInput(e: Event) {
		const val = parseFloat((e.target as HTMLInputElement).value);
		if (!isNaN(val)) {
			progressValue = val;
			onUpdateProgress?.(goal.id, val);
		}
	}
</script>

<tr>
	<td class="font-medium text-sm">{goal.name}</td>
	<td class="text-sm text-base-content/70">
		{goal.targetValue}{unitLabels[goal.unit] ?? goal.unit}
	</td>
	<td class="text-sm text-base-content/70">{goal.weight}%</td>
	<td>
		{#if phase === 'avance'}
			<div class="flex items-center gap-2">
				<input
					type="number"
					class="input input-bordered input-xs w-20"
					value={progressValue}
					min="0"
					max={goal.unit === 'porcentaje' ? 100 : undefined}
					oninput={handleProgressInput}
					aria-label="Avance de {goal.name}"
					readonly={!canEditProgress}
				/>
				{#if goal.unit === 'porcentaje'}
					<ProgressIndicator
						value={progressValue}
						max={100}
						color={progressValue >= goal.targetValue ? 'success' : 'error'}
					/>
				{:else}
					<ProgressIndicator value={progressValue} max={goal.targetValue} />
				{/if}
			</div>
		{:else if kpis.length > 0}
			<div class="flex flex-wrap gap-1">
				{#each kpis as kpi (kpi.id)}
					<KpiBadge {kpi} />
				{/each}
			</div>
		{:else}
			<span class="text-xs text-base-content/30 italic">Sin KPI</span>
		{/if}
	</td>
	<td class="text-right">
		<div class="flex items-center justify-end gap-1">
			{#if phase === 'avance'}
				{#if canComment}
					<button
						class="btn btn-ghost btn-square btn-xs relative"
						title="Comentarios"
						onclick={() => onOpenComments?.(goal)}
						aria-label="Comentarios de {goal.name}"
					>
						<MessageCircle class="w-3.5 h-3.5" />
						{#if (goal.comments?.length ?? 0) > 0}
							<span class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5">{goal.comments?.length}</span>
						{/if}
					</button>
				{/if}
			{:else if mode === 'editor'}
				<button
					class="btn btn-ghost btn-square btn-xs"
					title="Editar"
					onclick={() => onEdit(goal)}
					aria-label="Editar {goal.name}"
				>
					<Pencil class="w-3.5 h-3.5" />
				</button>
				{#if canDelete}
					<button
						class="btn btn-ghost btn-square btn-xs text-error"
						title="Eliminar"
						onclick={() => onDelete(goal.id)}
						aria-label="Eliminar {goal.name}"
					>
						<Trash2 class="w-3.5 h-3.5" />
					</button>
				{/if}
			{:else if onRequestChange}
				<button
					class="btn btn-ghost btn-xs text-warning"
					title="Solicitar cambio"
					onclick={() => onRequestChange(goal)}
					aria-label="Solicitar cambio en {goal.name}"
				>
					<MessageSquare class="w-3.5 h-3.5" />
					Solicitar cambio
				</button>
			{/if}
		</div>
	</td>
</tr>
