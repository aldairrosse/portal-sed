<script lang="ts">
	import { theme } from '$lib/stores/theme';
</script>

<svelte:head>
	<title>Ajustes — SED</title>
</svelte:head>

<div class="flex flex-col gap-6 max-w-2xl">
	<div class="flex flex-col gap-1.5">
		<h1 class="text-2xl font-semibold">Ajustes</h1>
		<p class="text-sm text-base-content/50">
			Personaliza tu experiencia en el portal.
		</p>
	</div>

	<!-- Apariencia -->
	<section class="card card-elevated p-6">
		<div class="flex flex-col gap-1.5">
			<h2 class="text-lg font-semibold">Apariencia</h2>
			<p class="text-sm text-base-content/50">
				Elige el tema del portal. Tu preferencia se guarda en este dispositivo.
			</p>
		</div>

		<div class="flex items-center justify-between gap-4 mt-5">
			<div class="flex flex-col gap-0.5">
				<span class="text-sm font-medium">Tema oscuro</span>
				<span class="text-xs text-base-content/50">
					{$theme === 'dark' ? 'Activado' : 'Desactivado'}
				</span>
			</div>
			<input
				type="checkbox"
				class="toggle"
				checked={$theme === 'dark'}
				onchange={(e) => theme.set(e.currentTarget.checked ? 'dark' : 'light')}
				aria-label="Alternar tema oscuro"
			/>
		</div>
	</section>

	<!-- Futuras secciones de ajustes van aquí -->
</div>
