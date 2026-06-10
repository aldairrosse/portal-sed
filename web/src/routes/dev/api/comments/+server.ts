import { json } from '@sveltejs/kit';
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'node:fs';
import { resolve } from 'node:path';

const DATA_DIR = resolve(process.cwd(), '..', 'data');
const DATA_FILE = resolve(DATA_DIR, 'comentarios.json');

const EMPTY_DATA = { comentarios: {} };

function read(): Record<string, unknown> {
	if (!existsSync(DATA_FILE)) return structuredClone(EMPTY_DATA);
	try {
		return JSON.parse(readFileSync(DATA_FILE, 'utf-8'));
	} catch {
		return structuredClone(EMPTY_DATA);
	}
}

function write(data: Record<string, unknown>): void {
	if (!existsSync(DATA_DIR)) mkdirSync(DATA_DIR, { recursive: true });
	writeFileSync(DATA_FILE, JSON.stringify(data, null, 2), 'utf-8');
}

export async function GET() {
	const data = read();
	return json(data);
}

export async function POST({ request }) {
	const body = await request.json();
	write(body);
	return json({ ok: true });
}
