<script lang="ts">
	import type { NineBoxQuadrantDef } from '$lib/types/nine-box';
	import { updateQuadrantDef } from '$lib/stores/nineBoxStore.svelte';
	import { X, Save } from '@lucide/svelte';

	interface Props {
		quadrantDef: NineBoxQuadrantDef;
		onClose: () => void;
	}

	let { quadrantDef, onClose }: Props = $props();

	let title = $state(quadrantDef.title);
	let description = $state(quadrantDef.description);
	let colorHex = $state(quadrantDef.colorHex);
	let saving = $state(false);
	let error = $state<string | null>(null);

	const HEX_REGEX = /^#[0-9A-Fa-f]{6}$/;

	let isValid = $derived(
		title.trim().length > 0 &&
		HEX_REGEX.test(colorHex)
	);

	async function handleSave() {
		if (!isValid) return;

		saving = true;
		error = null;

		try {
			await updateQuadrantDef(quadrantDef.quadrant, {
				title: title.trim(),
				description: description.trim(),
				colorHex: colorHex.trim()
			});
			onClose();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Error al guardar';
		} finally {
			saving = false;
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) onClose();
	}
</script>

<svelte:window onkeydown={(e) => {
		if (e.key === 'Escape') onClose();
	}} />

<dialog class="modal" open aria-modal="true" aria-label="Configurar cuadrante" onclose={onClose} onclick={handleBackdropClick}>
	<div class="modal-box max-w-md">
		<!-- Header -->
		<div class="flex items-start justify-between gap-2 mb-4">
			<div>
				<h3 class="font-bold text-lg">
					Configurar cuadrante {quadrantDef.quadrant}
				</h3>
				<p class="text-sm text-base-content/50 mt-1">
					{quadrantDef.label}
				</p>
			</div>
			<form method="dialog">
				<button type="submit" class="btn btn-ghost btn-sm btn-square" aria-label="Cerrar">
					<X class="w-4 h-4" />
				</button>
			</form>
		</div>

		<!-- Color preview -->
		<div
			class="w-full h-10 rounded-lg mb-4 border border-base-300"
			style="background-color: {colorHex};"
		></div>

		<!-- Form fields -->
		<div class="flex flex-col gap-4">
			<!-- Title -->
			<div class="form-control">
				<label for="qc-title" class="label">
					<span class="label-text">Título</span>
				</label>
				<input
					id="qc-title"
					type="text"
					class="input input-bordered w-full"
					bind:value={title}
					placeholder="Ej: Plan de acción inmediato"
				/>
			</div>

			<!-- Description -->
			<div class="form-control">
				<label for="qc-desc" class="label">
					<span class="label-text">Descripción</span>
				</label>
				<textarea
					id="qc-desc"
					class="textarea textarea-bordered w-full h-24"
					bind:value={description}
					placeholder="Descripción del cuadrante..."
				></textarea>
			</div>

			<!-- Color hex -->
			<div class="form-control">
				<label for="qc-color" class="label">
					<span class="label-text">Color (hex)</span>
				</label>
				<div class="flex gap-2 items-center">
					<input
						id="qc-color"
						type="color"
						class="w-10 h-10 rounded cursor-pointer border border-base-300"
						bind:value={colorHex}
					/>
					<input
						type="text"
						class="input input-bordered flex-1 font-mono"
						bind:value={colorHex}
						placeholder="#FF5733"
					/>
				</div>
				{#if colorHex && !HEX_REGEX.test(colorHex)}
					<p class="text-xs text-error mt-1" role="alert">Formato inválido. Usá #RRGGBB.</p>
				{/if}
			</div>

			<!-- Error -->
			{#if error}
				<div class="alert alert-error text-sm py-2">{error}</div>
			{/if}

			<!-- Save button -->
			<button
				type="button"
				class="btn btn-primary w-full"
				disabled={!isValid || saving}
				onclick={handleSave}
			>
				{#if saving}
					<span class="loading loading-spinner loading-sm"></span>
				{:else}
					<Save class="w-4 h-4" />
				{/if}
				Guardar
			</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>Cerrar</button>
	</form>
</dialog>
