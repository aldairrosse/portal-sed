<script lang="ts">
	interface Props {
		rows?: number;
		avatar?: boolean;
		variant?: 'default' | 'table' | 'card' | 'form' | 'category';
	}

	let { rows = 5, avatar = false, variant = 'default' }: Props = $props();

	let cardCols = $derived(rows > 6 ? 3 : 2);
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

	{#if variant === 'default'}
		{#each Array(rows) as _, i (i)}
			<div class="space-y-2">
				<div class="skeleton h-4 w-3/4"></div>
				<div class="skeleton h-3 w-1/2"></div>
			</div>
		{/each}

	{:else if variant === 'table'}
		<div class="overflow-hidden rounded-box border border-base-300">
			<!-- header skeleton -->
			<div class="flex gap-4 bg-base-200 p-3">
				<div class="skeleton h-4 w-1/4"></div>
				<div class="skeleton h-4 w-1/4"></div>
				<div class="skeleton h-4 w-1/4"></div>
				<div class="skeleton h-4 w-1/6"></div>
			</div>
			<!-- data rows -->
			{#each Array(rows) as _, i (i)}
				<div class="flex gap-4 border-t border-base-300 p-3">
					<div class="skeleton h-4 w-1/4"></div>
					<div class="skeleton h-4 w-1/4"></div>
					<div class="skeleton h-4 w-1/4"></div>
					<div class="skeleton h-3 w-1/5"></div>
				</div>
			{/each}
		</div>

	{:else if variant === 'card'}
		<div class="grid gap-4" style="grid-template-columns: repeat({cardCols}, 1fr)">
			{#each Array(rows) as _, i (i)}
				<div class="skeleton flex flex-col gap-3 rounded-box p-4">
					<div class="skeleton h-32 w-full rounded-box"></div>
					<div class="skeleton h-4 w-3/4"></div>
					<div class="skeleton h-3 w-1/2"></div>
				</div>
			{/each}
		</div>

	{:else if variant === 'form'}
		{#each Array(rows) as _, i (i)}
			<div class="space-y-1.5">
				<div class="skeleton h-3 w-16"></div>
				<div class="skeleton h-9 w-full rounded-input"></div>
			</div>
		{/each}

	{:else if variant === 'category'}
		<div class="space-y-4">
			{#each Array(rows) as _, i (i)}
				<div class="border border-base-300 rounded-box p-4 space-y-3">
					<!-- category header -->
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="skeleton h-5 w-40"></div>
							<div class="skeleton h-4 w-12 rounded-full"></div>
						</div>
						<div class="flex gap-1">
							<div class="skeleton h-6 w-6 rounded"></div>
							<div class="skeleton h-6 w-6 rounded"></div>
						</div>
					</div>
					<!-- goal rows -->
					{#each Array(2) as _, j (j)}
						<div class="flex items-center gap-3 border-t border-base-200 pt-2">
							<div class="skeleton h-4 w-1/3"></div>
							<div class="skeleton h-4 w-1/6"></div>
							<div class="skeleton h-4 w-1/6"></div>
							<div class="skeleton h-4 w-1/6"></div>
							<div class="flex-1"></div>
							<div class="skeleton h-5 w-5 rounded"></div>
						</div>
					{/each}
					<!-- add goal button -->
					<div class="flex justify-center pt-1">
						<div class="skeleton h-8 w-28 rounded-input"></div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
