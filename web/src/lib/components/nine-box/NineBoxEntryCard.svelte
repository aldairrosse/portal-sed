<script lang="ts">
	import type { NineBoxEntry, NineBoxScale } from '$lib/types/nine-box';
	import { getQuadrantDefs } from '$lib/stores/nineBoxStore.svelte';
	import { PROFILE_LABELS } from '$lib/types/evaluation';
	import { X, ChevronRight } from '@lucide/svelte';

	interface Props {
		entries: NineBoxEntry[];
		perf: NineBoxScale;
		pot: NineBoxScale;
		onClose: () => void;
	}

	let { entries, perf, pot, onClose }: Props = $props();

	const quadrantDefs = $derived(getQuadrantDefs());
	const quadrantDef = $derived(quadrantDefs.find((d) => d.id === entries[0]?.quadrant));

	function getProfileLabel(profileId: string): string {
		return PROFILE_LABELS[profileId as keyof typeof PROFILE_LABELS] ?? profileId;
	}
</script>

<dialog
	class="modal"
	open
	onclose={onClose}
	onclick={(e) => { if (e.target === e.currentTarget) onClose(); }}
>
	<div class="modal-box max-w-md">
		<!-- Header -->
		<div class="flex items-start justify-between gap-2 mb-4">
			<div>
				<h3 class="font-bold text-lg">{quadrantDef?.label ?? 'Celda'}</h3>
				<p class="text-sm text-base-content/50 mt-1">
					{entries.length} {entries.length === 1 ? 'empleado' : 'empleados'}
				</p>
			</div>
			<form method="dialog">
				<button type="submit" class="btn btn-ghost btn-sm btn-square" aria-label="Cerrar">
					<X class="w-4 h-4" />
				</button>
			</form>
		</div>

		<!-- Scores legend -->
		<div class="flex gap-4 mb-4 p-3 bg-base-200 rounded-lg">
			<div class="flex-1 text-center">
				<span class="text-xs font-medium text-base-content/70 block">Desempeño</span>
				<span class="text-xl font-bold text-base-content">{perf}</span>
			</div>
			<div class="divider divider-horizontal"></div>
			<div class="flex-1 text-center">
				<span class="text-xs font-medium text-base-content/70 block">Potencial</span>
				<span class="text-xl font-bold text-base-content">{pot}</span>
			</div>
		</div>

		<!-- Employee list -->
		<ul class="divide-y divide-base-200">
			{#each entries as entry (entry.id)}
				<li>
					<a
						href="/evaluacion/9x9/competencias/{entry.employeeId}"
						class="flex items-center justify-between gap-3 py-3 px-1 -mx-1 rounded-lg hover:bg-base-200/50 transition-colors group"
					>
						<div class="flex items-center gap-3">
							<div class="avatar placeholder">
								<div class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center">
									<span class="text-xs font-bold">
										{entry.employeeName.charAt(0).toUpperCase()}
									</span>
								</div>
							</div>
							<div>
								<p class="text-sm font-medium text-base-content group-hover:text-primary transition-colors">
									{entry.employeeName}
								</p>
								<p class="text-xs text-base-content/40">
									{getProfileLabel(entry.profileId)}
								</p>
							</div>
						</div>
						<ChevronRight class="w-4 h-4 text-base-content/30 group-hover:text-primary transition-colors" />
					</a>
				</li>
			{/each}
		</ul>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>Cerrar</button>
	</form>
</dialog>
