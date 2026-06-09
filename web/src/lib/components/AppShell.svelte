<script lang="ts">
	import Sidebar from './Sidebar.svelte';
	import DevToolbar from './DevToolbar.svelte';
	import logoBlack from '$lib/assets/logo_black.png';
	import logoWhite from '$lib/assets/logo_white.png';

	let { children } = $props();

	let mobileMenuOpen = $state(false);
</script>

<div class="drawer lg:drawer-open">
	<input id="main-drawer" type="checkbox" class="drawer-toggle" bind:checked={mobileMenuOpen} />

	<div class="drawer-content flex flex-col h-screen overflow-hidden">
		<!-- Mobile-only header -->
		<header class="navbar bg-base-100 lg:hidden sticky top-0 z-40 px-4 py-3">
			<div class="flex-none">
				<label for="main-drawer" class="btn btn-ghost btn-square" aria-label="Abrir menú">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-5 w-5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 6h16M4 12h16M4 18h16"
						/>
					</svg>
				</label>
			</div>
			<div class="flex-1 flex items-center">
				<img src={logoBlack} alt="SED" class="logo-theme-light h-[18px] w-auto max-w-full object-contain" />
				<img src={logoWhite} alt="SED" class="logo-theme-dark h-[18px] w-auto max-w-full object-contain" />
			</div>
		</header>

		<main class="flex-1 p-4 lg:p-8 overflow-y-auto min-w-0">
			{@render children()}
		</main>

		{#if import.meta.env.DEV}
			<DevToolbar />
		{/if}
	</div>

	<div class="drawer-side z-50">
		<label for="main-drawer" class="drawer-overlay" aria-hidden="true"></label>
		<Sidebar onclose={() => (mobileMenuOpen = false)} />
	</div>
</div>