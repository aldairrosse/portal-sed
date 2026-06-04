<script lang="ts">
	import { AlertTriangle, X } from '@lucide/svelte';

	interface Props {
		open: boolean;
		title?: string;
		message?: string;
		itemName: string;
		onConfirm: () => void;
		onCancel: () => void;
	}

	let {
		open,
		title = 'Confirmar eliminación',
		message,
		itemName,
		onConfirm,
		onCancel
	}: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function handleCancel() {
		onCancel();
	}

	function handleConfirm() {
		onConfirm();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleCancel();
		}
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	role="dialog"
	aria-modal="true"
	aria-labelledby="delete-confirm-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box">
		<div class="flex items-start gap-3">
			<div class="w-10 h-10 rounded-full bg-error/10 flex items-center justify-center flex-shrink-0">
				<AlertTriangle class="w-5 h-5 text-error" strokeWidth={2} />
			</div>
			<div class="flex-1 min-w-0">
				<h3 id="delete-confirm-title" class="text-lg font-semibold text-base-content">
					{title}
				</h3>
				<p class="text-base-content/60 mt-2 text-sm">
					{message ?? 'Esta acción no se puede deshacer.'}
				</p>
				<div class="mt-3 px-3 py-2 rounded-lg bg-base-200 text-sm font-medium text-base-content">
					{itemName}
				</div>
			</div>
			<button
				class="btn btn-ghost btn-square btn-sm flex-shrink-0"
				onclick={handleCancel}
				aria-label="Cancelar"
			>
				<X class="w-4 h-4" />
			</button>
		</div>

		<div class="modal-action mt-6">
			<button class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
			<button class="btn btn-error btn-sm" onclick={handleConfirm}>Eliminar</button>
		</div>
	</div>
</dialog>
