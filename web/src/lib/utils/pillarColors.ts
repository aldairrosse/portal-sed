import type { ColorfullIconName } from '$lib/components/ui/Icon.svelte';

const PILLAR_COLORS = ['#FF672A', '#7754E2', '#11E0AD', '#2E63E8'] as const;
const PILLAR_COLORS_DARK = ['#D25320', '#5439A6', '#0D896A', '#1C3D8F'] as const;
const PILLAR_ICONS = ['light', 'insignia', 'stats', 'notes-check'] as const satisfies readonly ColorfullIconName[];

export function pillarColors(index: number): {
	color: string;
	colorDarker: string;
	icon: ColorfullIconName;
	badgeStyle: string;
} {
	const color = PILLAR_COLORS[index % PILLAR_COLORS.length];
	return {
		color,
		colorDarker: PILLAR_COLORS_DARK[index % PILLAR_COLORS_DARK.length],
		icon: PILLAR_ICONS[index % PILLAR_ICONS.length],
		badgeStyle: `background-color: ${color}; color: #fff`
	};
}
