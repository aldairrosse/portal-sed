import type { ApiCyclePhase } from '$lib/types/cycle';
import type {
	Goal,
	GoalCategory,
	InstitutionalGoal,
} from '$lib/types/goal';
import {
	progressPercent,
	effectiveWeightGlobal,
	effectiveWeightShared,
	hierarchicalScore,
} from '$lib/utils/scoring';
import {
	getCategories,
	getGoalsByCategory,
	getInstitutionalGoals,
	getCycleWeights,
	getTeamWeights,
} from '$lib/stores/goalsStore.svelte';

export interface HomeGroupProgress {
	evaluated: number;
	total: number;
	/** Suma ponderada del grupo (%avance * peso meta * peso ponderado). */
	value: number;
}

export interface HomeProgress {
	personal: HomeGroupProgress;
	global: HomeGroupProgress;
	shared: HomeGroupProgress;
	total: number;
}

export interface ComputeHomeProgressArgs {
	personalGoals: Goal[];
	categories: GoalCategory[];
	institutionalGoals: InstitutionalGoal[];
	pWeight: number;
	pjWeight: number;
	phase: ApiCyclePhase;
}

/**
 * Progreso ponderado del ciclo para el Home ("Tu progreso del ciclo").
 *
 * Fórmulas idénticas a las de exportación/detalle de asignación
 * (`objetivos/asignacion/+page.svelte#buildRows`):
 * - institucional: `pct * pesoMeta/100 * pesoPonderado/100`
 * - personal: `pct * pesoMeta/100 * pesoPonderadoCat/100`
 *
 * En fase `asignacion` el avance es 0 y los contadores son 0 de X.
 * Una meta cuenta como evaluada solo si su `progressPercent` recalculado > 0.
 */
export function computeHomeProgress({
	personalGoals,
	categories,
	institutionalGoals,
	pWeight,
	pjWeight,
	phase,
}: ComputeHomeProgressArgs): HomeProgress {
	const personal: HomeGroupProgress = {
		evaluated: 0,
		total: personalGoals.length,
		value: 0,
	};
	const global: HomeGroupProgress = {
		evaluated: 0,
		total: 0,
		value: 0,
	};
	const shared: HomeGroupProgress = {
		evaluated: 0,
		total: 0,
		value: 0,
	};

	if (phase !== 'asignacion') {
		// ─── Personal: raw por categoría → hierarchicalScore (equivale a
		// pesoPonderado por categoría cuando no hay override explícito) ──
		let rawPersonal = 0;
		let overridePersonal = 0;
		let hasOverride = false;
		for (const cat of categories) {
			const catGoals = personalGoals.filter((g) => g.categoryId === cat.id);
			if (catGoals.length === 0) continue;
			const catShare = catGoals.reduce((acc, g) => {
				const pct = progressPercent(
					g.progress ?? 0,
					g.targetValue,
					g.baselineValue,
					g.direction,
				);
				if (pct > 0) personal.evaluated += 1;
				return acc + (g.weight / 100) * pct;
			}, 0);
			rawPersonal += (cat.weight / 100) * catShare;
			if (cat.effectiveWeight !== undefined && cat.effectiveWeight !== null) {
				hasOverride = true;
				for (const g of catGoals) {
					const pct = progressPercent(
						g.progress ?? 0,
						g.targetValue,
						g.baselineValue,
						g.direction,
					);
					overridePersonal +=
						((pct * (g.weight ?? 0)) / 100) * (cat.effectiveWeight / 100);
				}
			}
		}
		personal.value = hasOverride
			? overridePersonal
			: hierarchicalScore(rawPersonal, pWeight, pjWeight);

		// ─── Institucional: global + compartida ──────────────────────────
		for (const goal of institutionalGoals) {
			const pct = progressPercent(
				goal.currentValue ?? 0,
				goal.targetValue ?? 0,
				goal.baselineValue,
				goal.direction,
			);
			const target =
				goal.source === 'global'
					? ((): HomeGroupProgress => {
							global.total += 1;
							return global;
						})()
					: ((): HomeGroupProgress => {
							shared.total += 1;
							return shared;
						})();
			if (pct > 0) target.evaluated += 1;
			const eff =
				goal.effectiveWeight ??
				(goal.source === 'global'
					? effectiveWeightGlobal(goal.weight ?? 0, pWeight)
					: effectiveWeightShared(goal.weight ?? 0, pWeight, pjWeight));
			// eff ya incluye peso meta * factor jerárquico (G o J*P), igual que
			// backend scoring_service (pct * w/100 * G/100); no re-multiplicar por peso.
			target.value += (pct * eff) / 100;
		}
	} else {
		global.total = institutionalGoals.filter(
			(g) => g.source === 'global',
		).length;
		shared.total = institutionalGoals.filter(
			(g) => g.source === 'shared',
		).length;
	}

	// Cap 0-100 por segmento y total (personal+global+compartida ≤100, como FinalScore backend).
	const clamp = (v: number) => Math.min(100, Math.max(0, v));
	personal.value = clamp(personal.value);
	global.value = clamp(global.value);
	shared.value = clamp(shared.value);
	const __total = clamp(personal.value + global.value + shared.value);
	return {
		personal,
		global,
		shared,
		total: __total,
	};
}

/**
 * Conveniencia: lee categorías, metas institucionales y pesos del store y
 * calcula el progreso del empleado (ids de `EmployeeAssignment.goalIds`).
 * Usa `getCategories`/`getGoalsByCategory`/`getInstitutionalGoals` y
 * `getCycleWeights`/`getTeamWeights` como fuente de verdad.
 */
export function getHomeProgress(
	personalGoalIds: string[],
	phase: ApiCyclePhase,
): HomeProgress {
	const categories = getCategories();
	const idSet = new Set(personalGoalIds);
	const personalGoals = categories.flatMap((cat) =>
		getGoalsByCategory(cat.id).filter((g) => idSet.has(g.id)),
	);
	const institutionalGoals = getInstitutionalGoals();
	const { pWeight } = getCycleWeights();
	const { pjWeight } = getTeamWeights();
	// Pesos aún null = no cargados: fallback sin config P=100/PJ=100 (cap 100),
	// igual que backend GetOrFallback (G:0/P:100, J:0/PJ:100). +page.svelte muestra
	// skeleton hasta isWeightsLoaded() para evitar el parpadeo con estos interinos.
	return computeHomeProgress({
		personalGoals,
		categories,
		institutionalGoals,
		pWeight: pWeight ?? 100,
		pjWeight: pjWeight ?? 100,
		phase,
	});
}
