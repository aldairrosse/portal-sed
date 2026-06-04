<script lang="ts">
	import EvaluationStatusBadge from '$lib/components/evaluation/EvaluationStatusBadge.svelte';
	import ComparisonTable from '$lib/components/evaluation/ComparisonTable.svelte';
	import GoalClosureCard from '$lib/components/evaluation/GoalClosureCard.svelte';
	import { getPhase, getProfile } from '$lib/stores/devContext.svelte';
	import { getPillars, getCompetenciesByPillar, getLevelDefinitions, getCompetencyAcceptanceLevel } from '$lib/stores/competencyStore.svelte';
	import {
		getCompetencyRatings,
		getGoalClosures,
		getEvaluationStatus,
		addManagerComment
	} from '$lib/stores/evaluationStore.svelte';
	import { MANAGER_MAP } from '$lib/types/goal';
	import type { EvaluationProfile } from '$lib/types/evaluation';
	import { getAssignments, getAssignmentsByProfile, getGoalsByCategory, getCategories, getKpisForGoal } from '$lib/stores/goalsStore.svelte';

	const phase = $derived(getPhase());
	const profile = $derived(getProfile());
	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());
	const categories = $derived(getCategories());
	const allAssignments = $derived(getAssignments());

	const isFinAnio = $derived(phase === 'fin-anio');

	// Build inverse MANAGER_MAP to find direct reports for this profile
	const inverseManagerMap = $derived.by(() => {
		const map: Record<string, EvaluationProfile[]> = {};
		for (const [subordinate, manager] of Object.entries(MANAGER_MAP)) {
			if (!map[manager]) map[manager] = [];
			map[manager].push(subordinate as EvaluationProfile);
		}
		return map;
	});

	const subordinateProfiles = $derived(inverseManagerMap[profile] ?? []);

	const subordinateAssignments = $derived(
		subordinateProfiles.flatMap((prof) => getAssignmentsByProfile(prof))
	);

	let selectedEmployeeId = $state('');
	let searchQuery = $state('');
	let showDropdown = $state(false);

	const filteredAssignments = $derived(
		searchQuery.trim() === ''
			? subordinateAssignments
			: subordinateAssignments.filter((a) =>
					a.employeeName.toLowerCase().includes(searchQuery.toLowerCase())
				)
	);

	const selectedEmployeeName = $derived(
		subordinateAssignments.find((a) => a.employeeId === selectedEmployeeId)?.employeeName ?? ''
	);

	const allCompetencies = $derived(pillars.flatMap((p) => getCompetenciesByPillar(p.id)));
	const ratings = $derived(getCompetencyRatings(selectedEmployeeId));
	const closures = $derived(getGoalClosures(selectedEmployeeId));

	let activeTab = $state<'resumen' | 'competencias' | 'metas'>('resumen');

	function getCompetencies(pillarId: string) {
		return getCompetenciesByPillar(pillarId);
	}

	function getAcceptanceLevel(competencyId: string): number | undefined {
		return getCompetencyAcceptanceLevel(competencyId, profile)?.level;
	}

	function buildAcceptanceLevels(competencyIds: string[]): Record<string, number> {
		const result: Record<string, number> = {};
		for (const compId of competencyIds) {
			const level = getAcceptanceLevel(compId);
			if (level) result[compId] = level;
		}
		return result;
	}

	function handleSelectEmployee(employeeId: string) {
		selectedEmployeeId = employeeId;
		const emp = subordinateAssignments.find((a) => a.employeeId === employeeId);
		searchQuery = emp?.employeeName ?? '';
		showDropdown = false;
	}

	function handleManagerComment(goalId: string, comment: string) {
		addManagerComment(selectedEmployeeId, goalId, comment);
	}

	const allCompIds = $derived(allCompetencies.map((c) => c.id));
	const acceptanceLevels = $derived(buildAcceptanceLevels(allCompIds));

	const status = $derived(
		selectedEmployeeId
			? getEvaluationStatus(selectedEmployeeId, allCompetencies.length, [])
			: 'pending'
	);
</script>

<svelte:head>
	<title>Mis evaluados — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-bold text-base-content">Mis evaluados</h1>
		<p class="text-sm text-base-content/50 mt-1">
			Revisión de evaluaciones de tu equipo
		</p>
	</div>

	<!-- Employee search input -->
	<div class="relative w-full max-w-sm">
		<label class="label" for="employee-search">
			<span class="label-text">Buscar evaluado</span>
		</label>
		<input
			id="employee-search"
			type="text"
			class="input input-bordered input-sm w-full"
			placeholder="Escribí un nombre..."
			value={searchQuery}
			oninput={(e) => {
				searchQuery = (e.target as HTMLInputElement).value;
				showDropdown = true;
			}}
			onfocus={() => (showDropdown = true)}
			onblur={() => setTimeout(() => (showDropdown = false), 150)}
			aria-label="Buscar evaluado"
			aria-expanded={showDropdown}
			aria-autocomplete="list"
			role="combobox"
		/>
		{#if showDropdown && filteredAssignments.length > 0}
			<ul
				class="absolute z-50 mt-1 w-full bg-base-100 border border-base-300 rounded-lg shadow-lg max-h-60 overflow-auto"
				role="listbox"
			>
				{#each filteredAssignments as assignment (assignment.employeeId)}
					<li>
						<button
							type="button"
							class="w-full text-left px-4 py-2 text-sm hover:bg-base-200 {selectedEmployeeId === assignment.employeeId ? 'bg-base-200 font-semibold' : ''}"
							onmousedown={() => handleSelectEmployee(assignment.employeeId)}
							role="option"
							aria-selected={selectedEmployeeId === assignment.employeeId}
						>
							{assignment.employeeName}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	{#if selectedEmployeeId}
		<!-- Status badge -->
		<div class="flex items-center gap-2">
			<span class="text-sm text-base-content/60">Evaluado: <strong>{selectedEmployeeName}</strong></span>
			<EvaluationStatusBadge {status} />
		</div>

		<!-- Tabs -->
		<div role="tablist" class="tabs tabs-bordered">
			<button
				role="tab"
				class="tab {activeTab === 'resumen' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'resumen')}
				aria-selected={activeTab === 'resumen'}
			>
				Resumen
			</button>
			<button
				role="tab"
				class="tab {activeTab === 'competencias' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'competencias')}
				aria-selected={activeTab === 'competencias'}
			>
				Competencias
			</button>
			<button
				role="tab"
				class="tab {activeTab === 'metas' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'metas')}
				aria-selected={activeTab === 'metas'}
			>
				Metas
			</button>
		</div>

		<!-- Tab: Resumen -->
		{#if activeTab === 'resumen'}
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Resumen de competencias</h2>
				<ComparisonTable
					{ratings}
					competencies={allCompetencies}
					{acceptanceLevels}
					{levelDefinitions}
					showRhColumn={isFinAnio}
				/>
			</section>
		{/if}

		<!-- Tab: Competencias -->
		{#if activeTab === 'competencias'}
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Competencias</h2>
				{#each pillars as pillar (pillar.id)}
					{@const competencies = getCompetencies(pillar.id)}
					<div class="card bg-base-100 border border-base-300 mb-4">
						<div class="card-body px-4 py-4">
							<h3 class="text-base font-semibold text-base-content mb-4">{pillar.name}</h3>
							<div class="flex flex-col gap-3">
								{#each competencies as competency (competency.id)}
									{@const rating = ratings.find((r) => r.competencyId === competency.id)}
									<div class="border-t border-base-200 pt-3 first:border-t-0 first:pt-0">
										<div class="flex items-start justify-between">
											<div>
												<p class="text-sm font-medium text-base-content">{competency.name}</p>
												<p class="text-xs text-base-content/50">{competency.description}</p>
											</div>
											<div class="flex items-center gap-2 shrink-0">
												{#if rating?.selfRating}
													<span class="badge badge-sm" title="Autoevaluación">
														Auto: {rating.selfRating}
													</span>
												{:else}
													<span class="text-xs text-base-content/30">Sin autoevaluación</span>
												{/if}
												{#if rating?.rhRating}
													<span class="badge badge-sm" title="Evaluación RH">
														RH: {rating.rhRating}
													</span>
												{/if}
											</div>
										</div>
										{#if rating?.selfComment}
											<p class="text-xs text-base-content/50 mt-1 italic">
												"{rating.selfComment}"
											</p>
										{/if}
									</div>
								{/each}
							</div>
						</div>
					</div>
				{/each}
			</section>
		{/if}

		<!-- Tab: Metas -->
		{#if activeTab === 'metas'}
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Cierre de metas</h2>
				{#each categories as category (category.id)}
					{@const goals = getGoalsByCategory(category.id)}
					{#if goals.length > 0}
						<div class="card bg-base-100 border border-base-300 mb-4">
							<div class="card-body px-4 py-4">
								<h3 class="text-base font-semibold text-base-content mb-3">{category.name}</h3>
								<div class="flex flex-col gap-4">
									{#each goals as goal (goal.id)}
										{@const kpis = getKpisForGoal(goal.id)}
										{@const closure = closures.find((c) => c.goalId === goal.id)}
										<GoalClosureCard
											{goal}
											{kpis}
											{closure}
											mode="manager"
											canEdit={isFinAnio}
											onManagerComment={handleManagerComment}
										/>
									{/each}
								</div>
							</div>
						</div>
					{/if}
				{/each}
			</section>
		{/if}
	{:else}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			Selecciona un evaluado para comenzar.
		</p>
	{/if}
</div>
