<script lang="ts">
	import { MessageCircle, RefreshCcw } from '@lucide/svelte';
	import type { Goal, GoalComment } from '$lib/types/goal';
	import { getSession } from '$lib/api/session.svelte';

	interface Props {
		open: boolean;
		goal: Goal | null;
		comments: GoalComment[];
		assignment?: { id: string; name: string } | null;
		onAdd: (goalId: string, content: string) => void;
		onDelete?: (goalId: string, commentId: string) => void;
		onClose: () => void;
		onRefresh?: () => void;
		currentUserId?: string;
		/** When set, operates in category mode instead of goal mode */
		category?: { id: string; name: string } | null;
	}

	let {
		open,
		goal,
		comments,
		onAdd,
		onClose,
		onRefresh,
		currentUserId,
		category = null,
		assignment = null,
	}: Props = $props();

	const resolvedUserId = $derived(currentUserId ?? getSession().user?.employeeId ?? '');

	let dialogEl: HTMLDialogElement | undefined = $state();
	let newComment = $state('');
	let chatContainer: HTMLDivElement | undefined = $state();

	const entityId = $derived(assignment?.id ?? category?.id ?? goal?.id ?? '');

	const modalTitle = $derived.by(() => {
		if (category) return `Comentarios de categoría: ${category.name}`;
		if (goal) return `Comentarios de meta: ${goal.name}`;
		return 'Comentarios globales';
	});

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			newComment = '';
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	$effect(() => {
		if (chatContainer && comments.length) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
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
		if(!newComment.trim()) return;
		if (entityId) {
			onAdd(entityId, newComment.trim());
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
	<div class="modal-box max-w-lg h-[80vh] flex flex-col p-0">
		<!-- Header -->
		<div class="flex items-center gap-2 px-6 py-4">
			<MessageCircle class="w-5 h-5 text-primary" />
			<h3 id="comment-modal-title" class="font-semibold text-base-content flex-1">
				{modalTitle}
			</h3>
			<button class="btn btn-ghost btn-sm btn-circle" aria-label="Actualizar comentarios" onclick={()=> onRefresh?.()}>
				<RefreshCcw class="w-4 h-4" />
			</button>
			<form method="dialog">
				<button class="btn btn-ghost btn-sm btn-circle" aria-label="Cerrar">✕</button>
			</form>
		</div>

		<!-- Chat messages -->
		<div bind:this={chatContainer} class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
			{#if comments.length === 0}
				<p class="text-sm text-base-content/50 italic text-center py-8">Sin comentarios aún</p>
			{:else}
				{#each comments as comment (comment.id)}
					{@const isMe = comment.authorId === resolvedUserId}
					<div class="chat" class:chat-end={isMe} class:chat-start={!isMe}>
						<div class="chat-header">
							{#if isMe}
								<span class="font-semibold text-primary">{comment.authorName}</span>
							{:else}
								<span class="font-medium">{comment.authorName}</span>
							{/if}
							<time class="text-xs opacity-50">{timeAgo(comment.createdAt)}</time>
						</div>
						<div class="chat-bubble" class:chat-bubble-primary={isMe}>
							{comment.content}
						</div>
	
					</div>
				{/each}
			{/if}
		</div>

		<!-- Input area -->
		<form onsubmit={handleSubmit} class="px-6 py-4">
			<div class="relative">
				<input
					type="text"
					class="input input-bordered w-full pr-20"
					placeholder="Escribe un comentario..."
					bind:value={newComment}
				/>
				<button
					type="submit"
					class="absolute right-0 top-0 bottom-0 btn btn-primary !rounded-l-none rounded-r-md"
					disabled={!newComment.trim()}
				>
					Enviar
				</button>
			</div>
		</form>
	</div>
</dialog>
