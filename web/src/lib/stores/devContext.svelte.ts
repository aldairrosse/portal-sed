import type { EvaluationProfile, CyclePhase } from '$lib/types/evaluation';
import { isDev } from '$lib/dev/devEnv';

const STORAGE_KEY = 'sed-dev-context';

export interface DevContextState {
	profile: EvaluationProfile;
	phase: CyclePhase;
}

const DEFAULT_STATE: DevContextState = {
	profile: 'colaborador',
	phase: 'inicio-anio'
};

function loadFromStorage(): DevContextState | null {
	if (!isDev()) return null;
	try {
		const raw = sessionStorage.getItem(STORAGE_KEY);
		if (!raw) return null;
		return JSON.parse(raw) as DevContextState;
	} catch {
		return null;
	}
}

function saveToStorage(state: DevContextState): void {
	if (!isDev()) return;
	try {
		sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
	} catch {
		// ignore
	}
}

const stored = loadFromStorage();

let profile = $state<EvaluationProfile>(stored?.profile ?? DEFAULT_STATE.profile);
let phase = $state<CyclePhase>(stored?.phase ?? DEFAULT_STATE.phase);

export function getProfile(): EvaluationProfile {
	return profile;
}

export function setProfile(value: EvaluationProfile): void {
	profile = value;
	saveToStorage({ profile, phase });
}

export function getPhase(): CyclePhase {
	return phase;
}

export function setPhase(value: CyclePhase): void {
	phase = value;
	saveToStorage({ profile, phase });
}

export function getDevContext(): DevContextState {
	return { profile, phase };
}