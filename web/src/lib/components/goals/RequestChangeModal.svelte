<script lang="ts">
	import { X } from '@lucide/svelte';
	import { recordChangeRequest } from '$lib/stores/goalsStore.svelte';
	import type { ChangeRequest } from '$lib/types/goal';

	interface Props {
		open: boolean;
		entityType: ChangeRequest['entityType'];
		entityId: string;
		entityName: string;
		requestedBy: string;
		onClose: () => void;
	}

	let { open, entityType, entityId, entityName, requestedBy, onClose }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let reason = $state('');
	let submitted = $state(false);
	let error = $state('');

	const title = $derived(
		entityType === 'category'
			? 'Solicitar cambio en categoría'
			: entityType === 'assignment'
				? 'Solicitar cambio en asignación'
				: 'Solicitar cambio en meta'
	);

	const confirmText = $derived(
		entityType === 'category'
			? 'Se solicitará un cambio en la categoría:'
			: entityType === 'assignment'
				? 'Se solicitará un cambio en la asignación de:'
				: 'Se solicitará un cambio en la meta:'
	);

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			reason = '';
			submitted = false;
			error = '';
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function handleCancel() {
		onClose();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) handleCancel();
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		if (!reason.trim()) {
			error = 'Debe indicar el motivo del cambio.';
			return;
		}
		const changeRequest: ChangeRequest = {
			id: `cr-${Date.now()}`,
			entityType,
			entityId,
			action: 'update',
			changes: { reason: reason.trim() },
			reason: reason.trim(),
			requestedBy,
			requestedAt: new Date().toISOString(),
			status: 'pending'
		};
		recordChangeRequest(changeRequest);
		submitted = true;
		setTimeout(() => onClose(), 2000);
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="request-change-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box">
		{#if submitted}
			<div class="alert alert-success text-sm" role="status">
				<span>Solicitud de cambio enviada correctamente.</span>
			</div>
		{:else}
			<div class="flex items-center justify-between mb-5">
				<h3 id="request-change-title" class="text-lg font-semibold text-base-content">{title}</h3>
				<button
					class="btn btn-ghost btn-square btn-sm"
					onclick={handleCancel}
					aria-label="Cerrar"
				>
					<X class="w-4 h-4" />
				</button>
			</div>

			<p class="text-sm text-base-content/70 mb-4">
				{confirmText} <strong>{entityName}</strong>
			</p>

			<form onsubmit={handleSubmit}>
				{#if error}
					<div class="alert alert-error mb-4 text-sm" role="alert">
						<span>{error}</span>
					</div>
				{/if}

				<div class="form-control mb-4">
					<label class="label" for="request-reason">
						<span class="label-text">Motivo del cambio</span>
					</label>
					<textarea
						id="request-reason"
						class="textarea textarea-bordered w-full"
						rows={4}
						bind:value={reason}
						placeholder="Describa el cambio que desea solicitar..."
						required
						aria-required="true"
					></textarea>
				</div>

				<div class="modal-action">
					<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>
						Cancelar
					</button>
					<button type="submit" class="btn btn-warning btn-sm">
						Enviar solicitud
					</button>
				</div>
			</form>
		{/if}
	</div>
</dialog>
