const API = '/dev/api/comments';
const STORAGE_KEY = 'sed-dev-comentarios';

export interface CommentEntry {
	id: number;
	texto: string;
	fecha: string;
}

export type ComentariosMap = Record<number, CommentEntry[]>;

let map = $state<ComentariosMap>({});
let loaded = $state(false);

async function loadFromDisk(): Promise<ComentariosMap> {
	try {
		const res = await fetch(API);
		if (!res.ok) throw new Error('API error');
		const data = await res.json();
		return (data as { comentarios: ComentariosMap }).comentarios ?? {};
	} catch {
		return {};
	}
}

async function saveToDisk(data: ComentariosMap): Promise<void> {
	try {
		await fetch(API, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ comentarios: data })
		});
	} catch {
		// fallback: try localStorage
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
		} catch {
			// ignore
		}
	}
}

async function init() {
	if (loaded) return;
	map = await loadFromDisk();
	loaded = true;
}

const initPromise = init();

export function getComentarios(requisitoId: number): CommentEntry[] {
	return map[requisitoId] ?? [];
}

export async function agregarComentario(requisitoId: number, texto: string): Promise<void> {
	await initPromise;
	if (!texto.trim()) return;
	const entry: CommentEntry = {
		id: Date.now(),
		texto: texto.trim(),
		fecha: new Date().toISOString()
	};
	if (!map[requisitoId]) map[requisitoId] = [];
	map[requisitoId] = [...map[requisitoId], entry];
	await saveToDisk(map);
}

export async function eliminarComentario(requisitoId: number, commentId: number): Promise<void> {
	await initPromise;
	const arr = map[requisitoId];
	if (!arr) return;
	map[requisitoId] = arr.filter((c) => c.id !== commentId);
	await saveToDisk(map);
}

export async function exportarComentariosJSON(): Promise<void> {
	await initPromise;
	const blob = new Blob([JSON.stringify({ comentarios: map }, null, 2)], {
		type: 'application/json'
	});
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = `comentarios-requisitos-${new Date().toISOString().slice(0, 10)}.json`;
	a.click();
	URL.revokeObjectURL(url);
}

export interface ImportResult {
	ok: boolean;
	error?: string;
}

export async function importarComentariosJSON(file: File): Promise<ImportResult> {
	let text: string;
	try {
		text = await file.text();
	} catch {
		return { ok: false, error: 'No se pudo leer el archivo.' };
	}

	let parsed: unknown;
	try {
		parsed = JSON.parse(text);
	} catch {
		return { ok: false, error: 'El archivo no es un JSON válido.' };
	}

	if (typeof parsed !== 'object' || parsed === null || !('comentarios' in parsed)) {
		return { ok: false, error: 'Estructura inválida: falta la propiedad "comentarios".' };
	}

	const comentarios = (parsed as { comentarios: unknown }).comentarios;
	if (typeof comentarios !== 'object' || comentarios === null) {
		return { ok: false, error: '"comentarios" debe ser un objeto.' };
	}

	for (const [key, val] of Object.entries(comentarios)) {
		if (!/^\d+$/.test(key)) {
			return { ok: false, error: `Clave "${key}" no es un ID numérico válido.` };
		}
		if (!Array.isArray(val)) {
			return { ok: false, error: `La clave "${key}" debe contener un arreglo.` };
		}
		for (const entry of val) {
			if (!entry || typeof entry.id !== 'number' || typeof entry.texto !== 'string' || typeof entry.fecha !== 'string') {
				return {
					ok: false,
					error: `Entrada inválida en requerimiento "${key}": cada comentario debe tener "id" (número), "texto" (texto) y "fecha" (texto).`
				};
			}
		}
	}

	map = comentarios as ComentariosMap;
	await saveToDisk(map);
	return { ok: true };
}

export function hayComentarios(requisitoId: number): boolean {
	return (map[requisitoId]?.length ?? 0) > 0;
}

export async function getTotalComentarios(): Promise<number> {
	await initPromise;
	return Object.values(map).reduce((sum, arr) => sum + arr.length, 0);
}
