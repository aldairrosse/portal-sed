<script lang="ts">
	import type { KPI } from '$lib/types/goal';

	interface Props {
		kpi: KPI;
	}

	let { kpi }: Props = $props();
</script>

<span class="badge badge-outline badge-sm gap-1 whitespace-nowrap max-w-[12rem] truncate" title={kpi.description}>
	{kpi.name}
</span>
