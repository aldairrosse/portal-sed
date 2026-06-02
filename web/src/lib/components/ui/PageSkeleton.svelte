<script lang="ts">
	interface Props {
		rows?: number;
		avatar?: boolean;
	}

	let { rows = 5, avatar = false }: Props = $props();
</script>

<div class="space-y-3" aria-busy="true" aria-label="Cargando contenido">
	{#if avatar}
		<div class="flex items-center gap-3">
			<div class="skeleton size-10 rounded-full"></div>
			<div class="space-y-2">
				<div class="skeleton h-4 w-32"></div>
				<div class="skeleton h-3 w-24"></div>
			</div>
		</div>
	{/if}

	{#each Array(rows) as _, i (i)}
		<div class="space-y-2">
			<div class="skeleton h-4 w-3/4"></div>
			<div class="skeleton h-3 w-1/2"></div>
		</div>
	{/each}
</div>