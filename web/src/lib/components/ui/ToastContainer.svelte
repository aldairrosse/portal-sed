<script lang="ts">
	import { dismiss, getNotifications } from '$lib/stores/notifications.svelte';
	import type { Notification } from '$lib/stores/notifications.svelte';

	let notifications = $derived(getNotifications());

	// Track which notifications already have timers to avoid duplicates
	let handledIds = new Set<string>();

	$effect(() => {
		const current = notifications;
		for (const n of current) {
			if (n.duration && n.duration > 0 && !handledIds.has(n.id)) {
				handledIds.add(n.id);
				setTimeout(() => dismiss(n.id), n.duration);
			}
		}
		// Clean up stale ids
		handledIds = new Set(current.map((n) => n.id));
	});

	function icon(type: Notification['type']): string {
		switch (type) {
			case 'success':
				return '✅';
			case 'error':
				return '❌';
			case 'warning':
				return '⚠️';
			case 'info':
				return 'ℹ️';
		}
	}
</script>

{#if notifications.length > 0}
	<div class="toast toast-top toast-end z-[100] pointer-events-none">
		{#each notifications as n (n.id)}
			<div
				class="alert alert-{n.type} text-sm shadow-lg pointer-events-auto flex items-center gap-2 animate-toast-in"
				role="alert"
			>
				<span class="flex items-center gap-2">
					<span class="text-base leading-none">{icon(n.type)}</span>
					<span>{n.message}</span>
				</span>
				{#if n.dismissible !== false}
					<button
						class="btn btn-ghost btn-xs btn-square shrink-0"
						aria-label="Descartar notificación"
						onclick={() => dismiss(n.id)}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-4 w-4"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M18 6L6 18M6 6l12 12" />
						</svg>
					</button>
				{/if}
			</div>
		{/each}
	</div>
{/if}

<style>
	@keyframes toast-in {
		from {
			opacity: 0;
			transform: translateY(-0.5rem);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-toast-in {
		animation: toast-in 0.2s ease-out;
	}
</style>
