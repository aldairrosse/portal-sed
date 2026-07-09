<script lang="ts">
	import { X, MessageCircle } from '@lucide/svelte';
	import { recordChangeRequest } from '$lib/stores/goalsStore.svelte';
	import * as notifications from '$lib/stores/notifications.svelte';
	import type { ChangeRequest, GoalComment } from '$lib/types/goal';

	interface Props {
		open: boolean;
		entityType: ChangeRequest['entityType'];
		entityId: string;
		entityName: string;
		requestedBy: string;
		onClose: () => void;
		onCreated?: (entityType: ChangeRequest['entityType'], entityId: string) => void;
		comments?: GoalComment[];
		onAddComment?: (entityId: string, content: string) => void;
		currentUserId?: string;
	}

	let {
		open,
		entityType,
		entityId,
		entityName,
		requestedBy,
		onClose,
		onCreated,
		comments = [],
		onAddComment,
	}: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let message = $state('');
	let submitted = $state(false);

	const title = $derived(
		entityType === 'category'
			? 'Solicitar cambio en categoría'
			: entityType === 'assignment'
				? 'Solicitar cambio en asignación'
				: 'Solicitar cambio en meta'
	);

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			message = '';
			submitted = false;
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function handleCancel() { onClose(); }
	function handleBackdropClick(e: MouseEvent) { if (e.target === dialogEl) handleCancel(); }

	function timeAgo(dateStr: string): string {
		const diff = Date.now() - new Date(dateStr).getTime();
		const minutes = Math.floor(diff / 60000);
		if (minutes < 1) return 'ahora';
		if (minutes < 60) return `${minutes}m`;
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h`;
		const days = Math.floor(hours / 24);
		return `${days}d`;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!message.trim()) {
			notifications.error('Debe escribir un mensaje.');
			return;
		}
		try {
			// Always register as a comment
			if (onAddComment) {
				onAddComment(entityId, message.trim());
			}
			// Also record as change request
			await recordChangeRequest({
				id: '',
				entityType,
				entityId,
				action: 'update',
				changes: { reason: message.trim() },
				reason: message.trim(),
				requestedBy,
				requestedAt: new Date().toISOString(),
				status: 'pending'
			});
			submitted = true;
			if (onCreated) {
				setTimeout(() => { onClose(); onCreated(entityType, entityId); }, 800);
			} else {
				setTimeout(() => onClose(), 2000);
			}
		} catch {
			notifications.error('Error al enviar la solicitud.');
		}
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
	<div class="modal-box max-w-lg">
		{#if submitted}
			<div class="alert alert-success text-sm" role="status">
				<span>Solicitud de cambio enviada correctamente.</span>
			</div>
		{:else}
			<div class="flex items-center justify-between mb-4">
				<h3 id="request-change-title" class="text-lg font-semibold text-base-content">{title}</h3>
				<button class="btn btn-ghost btn-square btn-sm" onclick={handleCancel} aria-label="Cerrar">
					<X class="w-4 h-4" />
				</button>
			</div>

			<p class="text-sm text-base-content/70 mb-3">
				<strong>{entityName}</strong>
			</p>

			<!-- Comment history -->
			<div class="max-h-52 overflow-y-auto space-y-2 mb-4">
				{#if comments.length === 0}
					<p class="text-sm text-base-content/40 italic text-center py-3">Sin mensajes aún</p>
				{:else}
					{#each comments as comment (comment.id)}
						<div class="bg-base-200 rounded-lg p-2.5 space-y-0.5">
							<div class="flex items-center gap-2">
								<MessageCircle class="w-3 h-3 text-base-content/40" />
								<span class="text-xs font-medium">{comment.authorName}</span>
								<span class="text-xs text-base-content/30">{timeAgo(comment.createdAt)}</span>
							</div>
							<p class="text-sm text-base-content/70">{comment.content}</p>
						</div>
					{/each}
				{/if}
			</div>

			<!-- New message -->
			<form onsubmit={handleSubmit}>
				<div class="form-control mb-3">
					<textarea
						class="textarea textarea-bordered w-full text-sm"
						rows={3}
						placeholder="Escriba su mensaje..."
						bind:value={message}
						required
						aria-required="true"
					></textarea>
				</div>

				<div class="modal-action">
					<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
					<button type="submit" class="btn btn-warning btn-sm">Enviar solicitud</button>
				</div>
			</form>
		{/if}
	</div>
</dialog>
