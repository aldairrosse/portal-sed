import type { EvaluationProfile } from '$lib/types/evaluation';
import { getSession } from '$lib/api/session.svelte';

// ponytail: dev toolbar deprecated. Profile now derives from session.user.
// getPhase() removed — use getActivePhase() from cycle.svelte.ts instead.

export function getProfile(): EvaluationProfile {
	return getSession().user?.profileId ?? 'colaborador';
}
