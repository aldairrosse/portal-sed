<script lang="ts">
	interface Props {
		open: boolean;
		title: string;
		message: string;
		confirmLabel?: string;
		cancelLabel?: string;
		variant?: 'error' | 'warning' | 'info';
		disabled?: boolean;
		onconfirm: () => void;
		oncancel: () => void;
	}

	let {
		open,
		title,
		message,
		confirmLabel = 'Confirmar',
		cancelLabel = 'Cancelar',
		variant = 'info',
		onconfirm,
		oncancel,
		disabled = false
	}: Props = $props();

	function handleCancel() {
		oncancel();
	}

	function handleConfirm() {
		onconfirm();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			oncancel();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			oncancel();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		class="modal modal-open"
		role="dialog"
		aria-modal="true"
		aria-label={title}
		tabindex="-1"
		onclick={handleBackdropClick}
	>
		<div class="modal-box">
			<h3 class="text-lg font-semibold">{title}</h3>
			<p class="py-4 text-base-content/70">{message}</p>
			<div class="modal-action">
				<button class="btn btn-ghost btn-sm" onclick={handleCancel}>
					{cancelLabel}
				</button>
				<button
					class="btn btn-{variant} btn-sm"
					onclick={handleConfirm}
					disabled={disabled}
				>
					{confirmLabel}
				</button>
			</div>
		</div>
	</div>
{/if}
