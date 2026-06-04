<script lang="ts">
	import { X, MessageCircle } from '@lucide/svelte';
	import type { Goal, GoalComment } from '$lib/types/goal';

	interface Props {
		open: boolean;
		goal: Goal | null;
		comments: GoalComment[];
		onAdd: (goalId: string, content: string) => void;
		onDelete?: (goalId: string, commentId: string) => void;
		onClose: () => void;
		currentUserId?: string;
	}

	let {
		open,
		goal,
		comments,
		onAdd,
		onDelete,
		onClose,
		currentUserId = 'dev-user'
	}: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let newComment = $state('');

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			newComment = '';
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function handleCancel() {
		onClose();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleCancel();
		}
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		if (goal && newComment.trim()) {
			onAdd(goal.id, newComment.trim());
			newComment = '';
		}
	}

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
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="comment-modal-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box max-w-lg">
		<div class="flex items-center justify-between mb-4">
			<div class="flex items-center gap-2">
				<MessageCircle class="w-5 h-5 text-primary" />
				<h3 id="comment-modal-title" class="font-semibold text-base-content">
					{goal?.name ?? 'Comentarios'}
				</h3>
			</div>
			<button class="btn btn-ghost btn-square btn-sm" onclick={handleCancel} aria-label="Cerrar">
				<X class="w-4 h-4" />
			</button>
		</div>

		<!-- Comments list -->
		<div class="max-h-60 overflow-y-auto space-y-3 mb-4">
			{#if comments.length === 0}
				<p class="text-sm text-base-content/50 italic text-center py-4">Sin comentarios aún</p>
			{:else}
				{#each comments as comment (comment.id)}
					<div class="bg-base-200 rounded-lg p-3 space-y-1">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium">{comment.authorName}</span>
								<span class="text-xs text-base-content/40">{timeAgo(comment.createdAt)}</span>
							</div>
							{#if comment.authorId === currentUserId && onDelete && goal}
								<button
									class="btn btn-ghost btn-square btn-xs"
									onclick={() => onDelete(goal.id, comment.id)}
									aria-label="Eliminar comentario"
								>
									<X class="w-3 h-3" />
								</button>
							{/if}
						</div>
						<p class="text-sm text-base-content/70">{comment.content}</p>
					</div>
				{/each}
			{/if}
		</div>

		<!-- New comment form -->
		<form onsubmit={handleSubmit}>
			<div class="form-control">
				<textarea
					class="textarea textarea-bordered w-full text-sm"
					rows="3"
					placeholder="Agregar comentario..."
					bind:value={newComment}
				></textarea>
			</div>
			<div class="modal-action mt-4">
				<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cerrar</button>
				<button type="submit" class="btn btn-primary btn-sm" disabled={!newComment.trim()}>
					Enviar
				</button>
			</div>
		</form>
	</div>
</dialog>
