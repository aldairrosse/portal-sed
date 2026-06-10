<script lang="ts">
	import DATA, { type Requisito, type SeccionData } from '$lib/dev/requisitosData';
	import {
		getComentarios,
		agregarComentario,
		eliminarComentario,
		exportarComentariosJSON,
		importarComentariosJSON,
		hayComentarios
	} from '$lib/dev/commentsStore.svelte';
	import { FileDown, FileUp, FileText, MessageSquare, Trash2, ChevronDown, ChevronRight } from '@lucide/svelte';

	let expandedSections = $state<Set<string>>(new Set(DATA.map((s) => s.seccion)));
	let expandedEntregables = $state<Set<number>>(new Set());
	let expandedComentarios = $state<Set<number>>(new Set());
	let comentarioTexto = $state<Record<number, string>>({});

	function toggleSection(seccion: string) {
		if (expandedSections.has(seccion)) {
			expandedSections.delete(seccion);
		} else {
			expandedSections.add(seccion);
		}
		expandedSections = new Set(expandedSections);
	}

	function toggleEntregables(id: number) {
		if (expandedEntregables.has(id)) {
			expandedEntregables.delete(id);
		} else {
			expandedEntregables.add(id);
		}
		expandedEntregables = new Set(expandedEntregables);
	}

	function toggleComentarios(id: number) {
		if (expandedComentarios.has(id)) {
			expandedComentarios.delete(id);
		} else {
			expandedComentarios.add(id);
		}
		expandedComentarios = new Set(expandedComentarios);
	}

	function getComentarioTexto(requisitoId: number): string {
		return comentarioTexto[requisitoId] ?? '';
	}

	function setComentarioTexto(requisitoId: number, value: string) {
		comentarioTexto[requisitoId] = value;
		comentarioTexto = { ...comentarioTexto };
	}

	async function handleAgregarComentario(requisitoId: number) {
		const txt = getComentarioTexto(requisitoId);
		if (!txt.trim()) return;
		await agregarComentario(requisitoId, txt);
		setComentarioTexto(requisitoId, '');
	}

	async function handleExport() {
		await exportarComentariosJSON();
	}

	let showImportModal = $state(false);
	let importResult = $state<{ ok: boolean; msg: string } | null>(null);
	let importBusy = $state(false);

	function handleImport() {
		importResult = null;
		showImportModal = true;
	}

	function cancelImport() {
		showImportModal = false;
		importResult = null;
	}

	async function handleImportFile(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;
		importBusy = true;
		const res = await importarComentariosJSON(file);
		importBusy = false;
		target.value = '';

		if (res.ok) {
			showImportModal = false;
			importResult = { ok: true, msg: 'Comentarios importados correctamente.' };
			setTimeout(() => (importResult = null), 4000);
		} else {
			importResult = { ok: false, msg: res.error ?? 'Error desconocido.' };
			setTimeout(() => (importResult = null), 6000);
		}
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString('es-MX', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	let filtroTexto = $state('');
	let seccionesFiltradas = $derived.by(() => {
		if (!filtroTexto.trim()) return DATA;
		const q = filtroTexto.toLowerCase();
		return DATA.map((sec) => {
			const reqs = sec.requisitos.filter(
				(r) =>
					r.requerimiento.toLowerCase().includes(q) ||
					r.notas.toLowerCase().includes(q) ||
					r.entregables.some((e) => e.item.toLowerCase().includes(q))
			);
			return { ...sec, requisitos: reqs };
		}).filter((sec) => sec.requisitos.length > 0);
	});

	function toggleAllSections(expand: boolean) {
		if (expand) {
			expandedSections = new Set(DATA.map((s) => s.seccion));
		} else {
			expandedSections = new Set();
		}
	}

	const totalReqs = DATA.reduce((sum, s) => sum + s.requisitos.length, 0);
</script>

<svelte:head>
	<title>Requisitos del portal — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex flex-col gap-2">
		<div class="flex items-center justify-between flex-wrap gap-4">
			<div>
				<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
					<FileText class="w-6 h-6" />
					Requisitos del portal
				</h1>
				<p class="text-sm text-base-content/50 mt-1">
					{DATA.length} secciones · {totalReqs} requerimientos
				</p>
			</div>
			<div class="flex items-center gap-2 flex-wrap">
				<button class="btn btn-ghost btn-xs" onclick={() => toggleAllSections(true)}>
					Expandir todo
				</button>
				<button class="btn btn-ghost btn-xs" onclick={() => toggleAllSections(false)}>
					Colapsar todo
				</button>
				<button class="btn btn-outline btn-xs" onclick={handleExport}>
					<FileDown class="w-3.5 h-3.5" />
					Exportar comentarios
				</button>
				<button class="btn btn-outline btn-xs" onclick={handleImport}>
					<FileUp class="w-3.5 h-3.5" />
					Importar comentarios
				</button>
			</div>
		</div>

		<label class="input input-bordered flex items-center gap-2 max-w-md">
			<svg class="w-4 h-4 opacity-50" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
			</svg>
			<input
				type="text"
				class="grow"
				placeholder="Filtrar requerimientos..."
				bind:value={filtroTexto}
			/>
		</label>
	</div>

	<div class="flex flex-col gap-4">
		{#each seccionesFiltradas as seccion (seccion.seccion)}
			<article class="border border-base-300 rounded-xl overflow-hidden">
				<button
					class="w-full flex items-center justify-between px-5 py-4 hover:bg-base-200/50 transition-colors text-left"
					onclick={() => toggleSection(seccion.seccion)}
				>
					<div>
						<h2 class="text-lg font-semibold text-base-content">{seccion.seccion}</h2>
						<p class="text-xs text-base-content/40 mt-0.5">{seccion.descripcion}</p>
					</div>
					{#if expandedSections.has(seccion.seccion)}
						<ChevronDown class="w-5 h-5 text-base-content/30" />
					{:else}
						<ChevronRight class="w-5 h-5 text-base-content/30" />
					{/if}
				</button>

				{#if expandedSections.has(seccion.seccion)}
					<div class="overflow-x-auto">
						<table class="table table-sm table-zebra">
							<thead>
								<tr class="text-xs uppercase tracking-wider text-base-content/40">
									<th class="w-10 text-center">#</th>
									<th class="w-[30%] min-w-[200px]">Requerimiento</th>
									<th class="w-[35%] min-w-[220px]">Entregables realizados</th>
									<th class="w-[20%] min-w-[160px]">Notas / Detalles</th>
									<th class="w-[15%] min-w-[120px]">Comentarios</th>
								</tr>
							</thead>
							<tbody>
								{#each seccion.requisitos as req (req.id)}
									<tr class="align-top">
										<td class="text-center text-base-content/40 font-mono text-xs">
											{req.id}
										</td>
										<td class="text-sm font-medium text-base-content">
											{req.requerimiento}
										</td>
										<td>
											<div class="flex flex-col gap-1.5">
												{#each req.entregables as ent, i}
													<div class="text-xs text-base-content/70 leading-relaxed">
														{ent.item}
													</div>
													{#if expandedEntregables.has(req.id) && ent.archivos.length > 0}
														<div class="flex flex-col gap-0.5 ml-2 mb-1">
															{#each ent.archivos as archivo}
																<code class="text-[10px] text-base-content/30 font-mono">{archivo}</code>
															{/each}
														</div>
													{/if}
												{/each}
												{#if req.entregables.length > 0}
													<button
														class="text-[10px] text-primary/60 hover:text-primary flex items-center gap-1 mt-0.5"
														onclick={() => toggleEntregables(req.id)}
													>
														{expandedEntregables.has(req.id) ? 'Ocultar' : 'Ver'} archivos del código
													</button>
												{/if}
											</div>
										</td>
										<td>
											<p class="text-xs text-base-content/60 leading-relaxed">
												{req.notas}
											</p>
										</td>
										<td>
											<div class="flex flex-col gap-2">
												<button
													class="btn btn-ghost btn-xs gap-1 text-base-content/40 hover:text-base-content"
													onclick={() => toggleComentarios(req.id)}
												>
													<MessageSquare class="w-3 h-3" />
													{hayComentarios(req.id) ? getComentarios(req.id).length : 'Comentar'}
												</button>

												{#if expandedComentarios.has(req.id)}
													<div class="flex flex-col gap-2 min-w-[150px]">
														{#each getComentarios(req.id) as comment}
															<div class="bg-base-200/50 rounded-lg p-2 text-xs">
																<p class="text-base-content/70">{comment.texto}</p>
																<div class="flex items-center justify-between mt-1">
																	<span class="text-[10px] text-base-content/30">{formatDate(comment.fecha)}</span>
																	<button
																		class="text-error/50 hover:text-error"
																		onclick={async () => { await eliminarComentario(req.id, comment.id); }}
																	>
																		<Trash2 class="w-3 h-3" />
																	</button>
																</div>
															</div>
														{/each}
														<div class="flex gap-1">
							<input
									type="text"
									class="input input-xs input-bordered flex-1 min-w-0"
									placeholder="Escribe un comentario..."
									value={comentarioTexto[req.id] ?? ''}
									oninput={(e) => setComentarioTexto(req.id, (e.target as HTMLInputElement).value)}
									onkeydown={(e) => {
																	if (e.key === 'Enter') handleAgregarComentario(req.id);
																}}
															/>
															<button
																class="btn btn-primary btn-xs"
																onclick={() => handleAgregarComentario(req.id)}
															>
																Enviar
															</button>
														</div>
													</div>
												{/if}
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</article>
		{/each}
	</div>

	<footer class="text-center text-[10px] text-base-content/20 py-4 leading-relaxed">
		Ruta solo visible en desarrollo
	</footer>
</div>

{#if showImportModal}
	<div class="modal modal-open" role="dialog" onclick={cancelImport}>
		<div class="modal-box" onclick={(e) => e.stopPropagation()}>
			<h3 class="text-lg font-bold mb-2">Importar comentarios</h3>

			{#if importResult && !importResult.ok}
				<div class="alert alert-error text-sm mb-3">
					{importResult.msg}
				</div>
			{/if}

			<p class="text-sm text-base-content/60 mb-3">
				Seleccioná un archivo JSON con la siguiente estructura:
			</p>

			<pre class="bg-base-200 rounded-lg p-3 text-xs leading-relaxed overflow-x-auto"><code>{`{
  "comentarios": {
    "7": [
      {
        "id": 1749586800000,
        "texto": "Comentario de ejemplo",
        "fecha": "2026-06-10T16:00:00.000Z"
      }
    ]
  }
}`}</code></pre>

			<p class="text-xs text-base-content/40 mt-3">
				La clave numérica (ej. <code class="text-base-content/60">"7"</code>) es el ID del requerimiento.
				IDs disponibles: 1 al 33. Cada comentario requiere <code class="text-base-content/60">id</code>, <code class="text-base-content/60">texto</code> y <code class="text-base-content/60">fecha</code>.
			</p>

			<div class="flex items-center gap-3 mt-4">
				<input
					type="file"
					accept=".json"
					onchange={handleImportFile}
					class="file-input file-input-bordered file-input-sm flex-1"
					disabled={importBusy}
				/>
			</div>

			<div class="modal-action">
				<button class="btn btn-ghost" onclick={cancelImport}>Cancelar</button>
			</div>
		</div>
	</div>
{/if}

{#if importResult && importResult.ok}
	<div class="toast toast-top toast-end z-[60]">
		<div class="alert alert-success text-sm shadow-lg">
			{importResult.msg}
		</div>
	</div>
{/if}
