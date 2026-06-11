import { describe, it, expect, vi, beforeEach, afterEach, type Mock } from 'vitest';
import { client, baseURL } from './client';

function interceptLocationHref() {
	let assignedHref: string | null = null;

	const descriptor = Object.getOwnPropertyDescriptor(window, 'location')
		?? { value: window.location, writable: true, configurable: true };
	const origGet = descriptor.get?.();
	const origLocation = origGet ?? window.location;

	Object.defineProperty(window, 'location', {
		get() {
			return {
				...origLocation,
				get href() {
					return assignedHref ?? origLocation.href;
				},
				set href(val: string) {
					assignedHref = val;
				}
			};
		},
		set(val: any) {
			if (typeof val === 'string') {
				assignedHref = val;
			}
		},
		configurable: true
	});

	return {
		getAssignedHref: () => assignedHref,
		restore: () => {
			Object.defineProperty(window, 'location', descriptor);
		}
	};
}

describe('client.ts', () => {
	describe('baseURL', () => {
		it('reads VITE_API_URL from environment', () => {
			expect(baseURL).toBe('http://localhost:8080/api/v1');
		});

		it('reflects VITE_API_URL env var changes after module reload', async () => {
			vi.resetModules();
			vi.stubEnv('VITE_API_URL', 'http://custom-api.test/api');
			const mod = await import('./client');
			expect(mod.baseURL).toBe('http://custom-api.test/api');
			vi.unstubAllEnvs();
		});
	});

	describe('fetch interceptor', () => {
		let mockFetch: Mock;

		beforeEach(() => {
			mockFetch = vi.fn();
			vi.stubGlobal('fetch', mockFetch);
		});

		afterEach(() => {
			vi.unstubAllGlobals();
		});

		it('sets credentials: include on requests', async () => {
			mockFetch.mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }));

			await (client as any).GET('/test');

			expect(mockFetch).toHaveBeenCalledWith(
				expect.any(Request),
				expect.objectContaining({ credentials: 'include' })
			);
		});

		it('redirects to /login on 401 response', async () => {
			const { getAssignedHref, restore } = interceptLocationHref();
			mockFetch.mockResolvedValue(new Response(null, { status: 401 }));

			await (client as any).GET('/test');

			expect(getAssignedHref()).toBe('/login');
			restore();
		});
	});
});
