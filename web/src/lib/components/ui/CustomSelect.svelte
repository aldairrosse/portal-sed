<script lang="ts">
	import { ChevronDown } from '@lucide/svelte';

	interface Option {
		value: string;
		label: string;
	}

	interface Props {
		options: Option[];
		value: string;
		onChange: (value: string) => void;
		placeholder?: string;
		ariaLabel?: string;
		class?: string;
	}

	let {
		options,
		value,
		onChange,
		placeholder = 'Seleccionar',
		ariaLabel,
		class: className = ''
	}: Props = $props();

	const uid = $props.id();
	const popoverId = `custom-select-${uid}`;
	const anchorName = `--custom-select-anchor-${uid}`;

	let open = $state(false);
	let triggerEl: HTMLButtonElement | undefined = $state();
	let menuEl: HTMLUListElement | undefined = $state();

	const selectedLabel = $derived(
		options.find((o) => o.value === value)?.label ?? placeholder
	);

	function close() {
		if (menuEl?.matches(':popover-open')) {
			menuEl.hidePopover();
		}
	}

	function select(optionValue: string) {
		onChange(optionValue);
		close();
	}

	function handleToggle(e: ToggleEvent) {
		open = e.newState === 'open';
		if (!open) return;
		requestAnimationFrame(() => {
			const selected = menuEl?.querySelector<HTMLButtonElement>(
				'[aria-selected="true"] button'
			);
			const fallback = menuEl?.querySelector<HTMLButtonElement>('[role="option"] button');
			const target = selected ?? fallback;
			if (target) {
				target.scrollIntoView({ block: 'nearest' });
				target.focus({ preventScroll: true });
			}
		});
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!open) return;

		if (e.key === 'Escape') {
			e.preventDefault();
			close();
			return;
		}

		const items = menuEl?.querySelectorAll<HTMLButtonElement>('[role="option"] button');
		if (!items?.length) return;

		const currentIndex = Array.from(items).findIndex((el) => el === document.activeElement);

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			const next = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
			items[next].focus();
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			const prev = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
			items[prev].focus();
		} else if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			const focused = document.activeElement as HTMLElement;
			const optionValue = focused?.getAttribute('data-value');
			if (optionValue) select(optionValue);
		} else if (e.key === 'Home') {
			e.preventDefault();
			items[0].focus();
		} else if (e.key === 'End') {
			e.preventDefault();
			items[items.length - 1].focus();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<button
	bind:this={triggerEl}
	type="button"
	role="combobox"
	popovertarget={popoverId}
	style="anchor-name:{anchorName}"
	aria-haspopup="listbox"
	aria-expanded={open}
	aria-label={ariaLabel}
	class="input input-bordered input-sm w-48 text-left flex items-center justify-between gap-2 cursor-pointer h-8 min-h-0 text-xs truncate flex-shrink-0 {className}"
>
	<span class="truncate">{selectedLabel}</span>
	<span class="flex-shrink-0 transition-transform" class:rotate-180={open}>
		<ChevronDown class="w-3.5 h-3.5" />
	</span>
</button>

<ul
	bind:this={menuEl}
	popover
	id={popoverId}
	style="position-anchor:{anchorName}; width: anchor-size(width);"
	class="dropdown menu menu-xs bg-base-100 rounded-box p-1 shadow-lg border border-base-300 max-h-48 overflow-y-auto"
	role="listbox"
	aria-label={ariaLabel}
	ontoggle={handleToggle}
>
	{#each options as option (option.value)}
		<li role="option" aria-selected={option.value === value} data-value={option.value}>
			<button
				type="button"
				class:menu-active={option.value === value}
				tabindex="-1"
				onclick={() => select(option.value)}
			>
				{option.label}
			</button>
		</li>
	{/each}
</ul>
