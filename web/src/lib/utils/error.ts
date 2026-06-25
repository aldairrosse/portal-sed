// ponytail: single entry point for converting any caught value into a
// user-facing Spanish string. Handles:
//   - JS Error instances (TypeError from failed fetch, AbortError on timeout, etc.)
//   - openapi-fetch error bodies ({ error: { message, code, details, trace_id } })
//   - plain strings, null/undefined, anything else → fallback
export function humanizeError(err: unknown, fallback = 'Error desconocido'): string {
	if (err == null) return fallback;

	// Error instances: detect connection/timeout/CORS patterns and return clear Spanish
	if (err instanceof Error) {
		const msg = err.message || '';
		const lower = msg.toLowerCase();

		// ponytail: fetch throws TypeError on network failure (browser-dependent message)
		if (err instanceof TypeError && /fetch|network|load failed/i.test(msg)) {
			return 'No se pudo conectar con el servidor. Verificá tu conexión e intentá de nuevo.';
		}

		// ponytail: AbortError on request timeout
		if (err.name === 'AbortError' || /timeout|timed out|aborted/i.test(lower)) {
			return 'La solicitud tardó demasiado. Intentá de nuevo.';
		}

		// ponytail: CORS / mixed-content blocks
		if (/cors|access-control|opaque/i.test(lower)) {
			return 'No se pudo acceder al servidor. Verificá la configuración de red.';
		}

		return msg || fallback;
	}

	// API response error object: { error: { message, details, code, trace_id } } or { message }
	if (typeof err === 'object') {
		const inner = (err as { error?: { message?: string; details?: string[] } }).error;
		if (inner?.message) return inner.message;
		if (inner?.details?.length) return inner.details.join(', ');
		const direct = (err as { message?: string }).message;
		if (direct) return direct;
		return fallback;
	}

	if (typeof err === 'string') return err || fallback;

	return fallback;
}
