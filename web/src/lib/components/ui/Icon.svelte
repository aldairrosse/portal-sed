<script lang="ts">
	export type OutlineIconName =
		| 'analytics-report'
		| 'arrow-right'
		| 'bussines-bag'
		| 'caret-down'
		| 'caret-up'
		| 'check-circle'
		| 'connect'
		| 'cycles'
		| 'delete'
		| 'directions'
		| 'edit-line'
		| 'edit-line-alt'
		| 'file-check'
		| 'home'
		| 'layers'
		| 'logout'
		| 'notification'
		| 'plot'
		| 'send'
		| 'settings-bars'
		| 'squares'
		| 'star'
		| 'users';

	export type ColorfullIconName = 'insignia' | 'light' | 'notes-check' | 'stats';

	export type IconName = OutlineIconName | ColorfullIconName;

	interface Props {
		name: IconName;
		variant?: 'outline' | 'colorfull';
		size?: number;
		color?: string;
		class?: string;
		[key: string]: unknown;
	}

	// Raw SVG markup, keyed by file path (e.g. '../../assets/icons/outline/star.svg').
	const outlineGlob = import.meta.glob('../../assets/icons/outline/*.svg', {
		query: '?raw',
		import: 'default',
		eager: true
	}) as Record<string, string>;

	const colorfullGlob = import.meta.glob('../../assets/icons/colorfull/*.svg', {
		query: '?raw',
		import: 'default',
		eager: true
	}) as Record<string, string>;

	function toIconName(key: string): string {
		return key.split('/').pop()!.replace(/\.svg$/, '');
	}

	// ponytail: loose internal maps; the strict union is enforced by the `name` prop type
	const outlineIcons: Record<string, string> = Object.fromEntries(
		Object.entries(outlineGlob).map(([key, markup]) => [toIconName(key), markup])
	);

	const colorfullIcons: Record<string, string> = Object.fromEntries(
		Object.entries(colorfullGlob).map(([key, markup]) => [toIconName(key), markup])
	);

	// Guard against a broken glob path silently rendering nothing.
	if (import.meta.env.DEV) {
		if (Object.keys(outlineIcons).length !== 23) {
			console.warn(`Icon: expected 23 outline icons, found ${Object.keys(outlineIcons).length}`);
		}
		if (Object.keys(colorfullIcons).length !== 4) {
			console.warn(
				`Icon: expected 4 colorfull icons, found ${Object.keys(colorfullIcons).length}`
			);
		}
	}

	let { name, variant = 'outline', size = 24, color, class: className, ...rest }: Props =
		$props();

	const icon = $derived(variant === 'colorfull' ? colorfullIcons[name] : outlineIcons[name]);
</script>

<span
	{...rest}
	class="icon {variant} {className ?? ''}"
	style="display:inline-flex;position:relative;width:{size}px;height:{size}px;{color ? `color:${color};` : ''}"
>
	{#if icon}
		{@html icon}
	{/if}
</span>

<style>
	.icon :global(svg) {
		display: block;
		width: 100%;
		height: 100%;
	}

	/* Tint outline icons (hardcoded fill="#5E5E5E" is a presentation attribute, CSS wins).
	   Colorfull icons keep their gradient url(#...) fills. */
	.outline :global(svg path) {
		fill: currentColor;
	}
</style>
