<script lang="ts">
	import { getProfile, getPhase } from '$lib/stores/devContext.svelte';
	import { PROFILE_LABELS, type EvaluationProfile, type CyclePhase } from '$lib/types/evaluation';
	import { PROFILE_USERS } from '$lib/dev/profileUsers';
	import { getGoals, getCategories, getAssignments, getKpis, getGoalKpiLinks } from '$lib/stores/goalsStore.svelte';
	import pillarsData from '$lib/fixtures/competency/pillars.json';
	import competenciesData from '$lib/fixtures/competency/competencies.json';
	import acceptanceLevelsData from '$lib/fixtures/competency/acceptance-levels.json';
	import competencyAcceptanceData from '$lib/fixtures/competency/competency-acceptance-levels.json';
	import {
		ClipboardCheck,
		Target,
		TrendingUp,
		Users,
		Grid3x3,
		Award,
		BarChart3,
		Building,
		User
	} from '@lucide/svelte';
	import type { Goal, GoalCategory, KPI } from '$lib/types/goal';
	import type { Pillar, Competency, LevelDefinition } from '$lib/types/competency';

	const profile = $derived(getProfile());
	const phase = $derived(getPhase());
	const user = $derived(PROFILE_USERS[profile]);
	const profileLabel = $derived(PROFILE_LABELS[profile]);

	const assignments = $derived(getAssignments());
	const allGoals = $derived(getGoals());
	const allCategories = $derived(getCategories());
	const allKpis = $derived(getKpis());
	const allLinks = $derived(getGoalKpiLinks());

	const pillars = pillarsData as Pillar[];
	const competencies = competenciesData as Competency[];
	const levelDefs = acceptanceLevelsData as LevelDefinition[];
	const competencyLevels = competencyAcceptanceData as { competencyId: string; profileId: string; level: number }[];

	const myAssignment = $derived(assignments.find((a) => a.profileId === profile));
	const myEmployeeId = $derived(myAssignment?.employeeId ?? '');
	const myManager = $derived(
		myAssignment?.managerId
			? assignments.find((a) => a.employeeId === myAssignment.managerId)
			: null
	);
	const myDirectReports = $derived(
		myEmployeeId ? assignments.filter((a) => a.managerId === myEmployeeId) : []
	);

	const hasReports = $derived(
		['jefe', 'gerente-tienda', 'divisional', 'regional', 'director'].includes(profile)
	);
	const isRh = $derived(profile === 'rh');

	const areaMap: Record<EvaluationProfile, string> = {
		colaborador: 'Operaciones · Sucursal Centro',
		jefe: 'Servicio al Cliente',
		vendedor: 'Ventas · Tienda Polanco',
		'gerente-tienda': 'Tienda Polanco',
		divisional: 'División Comercial',
		regional: 'Región Centro',
		director: 'Dirección General',
		rh: 'Recursos Humanos'
	};
	const myArea = $derived(areaMap[profile]);

	const myGoals = $derived(
		myAssignment
			? myAssignment.goalIds
					.map((id) => allGoals.find((g) => g.id === id))
					.filter((g): g is Goal => Boolean(g))
			: []
	);

	const myKpis = $derived(
		(() => {
			const kpiIds = new Set<string>();
			myGoals.forEach((g) => {
				allLinks
					.filter((l) => l.goalId === g.id)
					.forEach((l) => kpiIds.add(l.kpiId));
			});
			return Array.from(kpiIds)
				.map((id) => allKpis.find((k) => k.id === id))
				.filter((k): k is KPI => Boolean(k));
		})()
	);

	const myCompetencies = $derived(
		competencies.map((c) => {
			const lvl = competencyLevels.find(
				(cl) => cl.competencyId === c.id && cl.profileId === profile
			);
			const def = levelDefs.find((ld) => ld.level === (lvl?.level ?? 0));
			const pillar = pillars.find((p) => p.id === c.pillarId);
			return {
				...c,
				pillarName: pillar?.name ?? '',
				level: lvl?.level ?? 0,
				levelLabel: def?.label ?? '—'
			};
		})
	);

	const competenciesByPillar = $derived(
		pillars.map((p) => ({
			pilar: p,
			items: myCompetencies.filter((c) => c.pillarId === p.id)
		}))
	);

	const today = new Date();
	const year = today.getFullYear();
	const startOfYear = new Date(year, 0, 1);
	const endOfYear = new Date(year, 11, 31);
	const dayOfYear =
		Math.floor((today.getTime() - startOfYear.getTime()) / (86400000)) + 1;
	const yearProgress = Math.min(100, Math.max(0, (dayOfYear / 365) * 100));

	const phaseGuidance: Record<CyclePhase, string> = {
		'inicio-anio':
			'Fija tus objetivos y KPIs con tu jefe. Define metas claras para el ciclo.',
		'medio-anio':
			'Revisa el avance de tus objetivos. Ajusta lo necesario antes del cierre.',
		'fin-anio':
			'Evalúa tu desempeño y completa la autoevaluación.'
	};

	function getGoalCategoryName(goal: Goal): string {
		const cat = allCategories.find((c) => c.id === goal.categoryId);
		return cat?.name ?? '';
	}

	function formatKpiTarget(kpi: KPI): string {
		const val = kpi.targetValue ?? 0;
		if (kpi.unit === 'porcentaje') return `${val}%`;
		if (kpi.unit === 'moneda') return `$${val.toLocaleString('es-MX')}`;
		return `${val}`;
	}
</script>

<svelte:head>
	<title>Inicio — SED</title>
</svelte:head>

<div class="max-w-5xl mx-auto space-y-16">
	<!-- Welcome + profile row -->
	<header>
		<p class="text-sm text-base-content/40">Hola,</p>
		<h1 class="mt-1">{user.name}</h1>
		<div class="flex flex-wrap items-center gap-3 mt-3">
			<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium">
				<ClipboardCheck class="w-3 h-3" />
				{profileLabel}
			</span>
			<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-base-200 text-base-content/60 text-xs font-medium">
				<Building class="w-3 h-3" />
				{myArea}
			</span>
			{#if hasReports && myDirectReports.length > 0}
				<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-base-200 text-base-content/60 text-xs font-medium">
					<Users class="w-3 h-3" />
					Lidera a {myDirectReports.length}
					{myDirectReports.length === 1 ? 'persona' : 'personas'}
				</span>
			{:else if myManager}
				<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-base-200 text-base-content/60 text-xs font-medium">
					<User class="w-3 h-3" />
					Reporta a {myManager.employeeName}
				</span>
			{/if}
		</div>
	</header>

	<!-- Cycle status + Competencies + Quick access (responsive row) -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-12">
		<!-- Cycle -->
		<section class="lg:col-span-2">
			<h2 class="text-xs font-semibold text-base-content/50 tracking-wide mb-3">
				Ciclo {year}
			</h2>
			<div class="relative h-2 bg-base-200 rounded-full overflow-hidden">
				<div
					class="absolute inset-y-0 left-0 bg-success/70 rounded-full"
					style="width: {yearProgress}%"
				></div>
				<div
					class="absolute inset-y-0 w-px bg-base-content/10"
					style="left: 33.33%"
				></div>
				<div
					class="absolute inset-y-0 w-px bg-base-content/10"
					style="left: 66.66%"
				></div>
			</div>
			<div class="flex justify-between mt-2">
				<span class="inline-flex items-center gap-1.5 text-xs {phase === 'inicio-anio' ? 'text-primary font-semibold' : 'text-base-content/40'}">
					{#if phase === 'inicio-anio'}
						<span class="inline-flex items-center px-2 py-0.5 rounded-full bg-primary/10 text-primary text-[11px] font-semibold">
							Inicio de año
						</span>
					{:else}
						Inicio
					{/if}
				</span>
				<span class="inline-flex items-center gap-1.5 text-xs {phase === 'medio-anio' ? 'text-primary font-semibold' : 'text-base-content/40'}">
					{#if phase === 'medio-anio'}
						<span class="inline-flex items-center px-2 py-0.5 rounded-full bg-primary/10 text-primary text-[11px] font-semibold">
							Medio año
						</span>
					{:else}
						Medio año
					{/if}
				</span>
				<span class="inline-flex items-center gap-1.5 text-xs {phase === 'fin-anio' ? 'text-primary font-semibold' : 'text-base-content/40'}">
					{#if phase === 'fin-anio'}
						<span class="inline-flex items-center px-2 py-0.5 rounded-full bg-primary/10 text-primary text-[11px] font-semibold">
							Fin de año
						</span>
					{:else}
						Fin de año
					{/if}
				</span>
			</div>
			<p class="mt-3 text-sm text-base-content/70">{phaseGuidance[phase]}</p>
		</section>

		<!-- Quick access cards -->
		<section>
			<h2 class="text-xs font-semibold text-base-content/50 tracking-wide mb-3">
				Acceso rápido
			</h2>
			<div class="grid grid-cols-2 gap-3">
				<a
					href="/mi-evaluacion"
					class="flex flex-col items-center gap-2 p-3 rounded-xl bg-base-200/50 hover:bg-base-200 transition-colors text-center"
				>
					<ClipboardCheck class="w-5 h-5 text-primary/70" />
					<span class="text-xs font-medium text-base-content">Mi evaluación</span>
				</a>
				<a
					href="/objetivos/asignacion"
					class="flex flex-col items-center gap-2 p-3 rounded-xl bg-base-200/50 hover:bg-base-200 transition-colors text-center"
				>
					<Target class="w-5 h-5 text-primary/70" />
					<span class="text-xs font-medium text-base-content">Asignación</span>
				</a>
			</div>
		</section>
	</div>

	<!-- Competencies summary -->
	{#if myCompetencies.length > 0}
		<section>
			<h2 class="text-xs font-semibold text-base-content/50 tracking-wide mb-3">
				Competencias
			</h2>
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
				{#each competenciesByPillar as group (group.pilar.id)}
					{#if group.items.length > 0}
						<div>
							<h3 class="text-xs font-medium text-base-content/50 mb-2">
								{group.pilar.name}
							</h3>
							<ul class="space-y-1">
								{#each group.items as comp (comp.id)}
									<li class="text-sm text-base-content">
										{comp.name}
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				{/each}
			</div>
		</section>
	{/if}

	<!-- Goals + KPIs (responsive row) -->
	{#if myGoals.length > 0 || myKpis.length > 0}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
			<!-- Goals -->
			{#if myGoals.length > 0}
				<section>
					<div class="flex items-baseline justify-between mb-3">
						<h2 class="text-xs font-semibold text-base-content/50 tracking-wide">
							Objetivos
						</h2>
						<a href="/objetivos/asignacion" class="text-xs text-primary hover:underline">
							Ver todos →
						</a>
					</div>
					<ul class="space-y-3">
						{#each myGoals as goal (goal.id)}
							{@const categoryName = getGoalCategoryName(goal)}
							<li>
								<div class="flex items-baseline justify-between mb-1">
									<span class="font-medium text-sm text-base-content">{goal.name}</span>
									<span class="text-xs font-mono text-primary ml-3">
										{goal.weight}%
									</span>
								</div>
								<div class="h-1.5 bg-base-200 rounded-full overflow-hidden">
									<div
										class="h-full bg-primary/70 rounded-full"
										style="width: {goal.weight}%"
									></div>
								</div>
								{#if categoryName}
									<p class="text-[11px] text-base-content/40 mt-1">{categoryName}</p>
								{/if}
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			<!-- KPIs -->
			{#if myKpis.length > 0}
				<section>
					<div class="flex items-baseline justify-between mb-3">
						<h2 class="text-xs font-semibold text-base-content/50 tracking-wide">
							KPIs
						</h2>
						<a href="/objetivos/asignacion" class="text-xs text-primary hover:underline">
							Ver catálogo →
						</a>
					</div>
					<ul class="space-y-3">
						{#each myKpis as kpi (kpi.id)}
							<li class="flex items-center gap-3">
								<BarChart3 class="w-3.5 h-3.5 text-base-content/30 flex-shrink-0" />
								<span class="font-medium text-sm text-base-content flex-1 min-w-0 truncate">{kpi.name}</span>
								<span class="text-xs text-primary font-mono flex-shrink-0">
									Meta: {formatKpiTarget(kpi)}
								</span>
							</li>
						{/each}
					</ul>
				</section>
			{/if}
		</div>
	{/if}

	<!-- Team summary (boss view) -->
	{#if hasReports && myDirectReports.length > 0}
		<section>
			<div class="flex items-baseline justify-between mb-3">
				<h2 class="text-xs font-semibold text-base-content/50 tracking-wide">
					Tu equipo
				</h2>
				<a href="/mis-evaluados" class="text-xs text-primary hover:underline">
					Ver todos →
				</a>
			</div>
			<ul class="space-y-3">
				{#each myDirectReports as report (report.id)}
					{@const reportGoals = report.goalIds.length}
					{@const reportProgress = Math.min(100, Math.round(reportGoals * 12.5))}
					{@const reportStatus = phase === 'inicio-anio'
						? { label: 'Asignando metas', tone: 'accent' }
						: reportGoals >= 4
							? { label: 'Evaluado', tone: 'success' }
							: { label: 'En proceso', tone: 'info' }}
					<li class="flex items-center gap-3 py-2">
						<div class="w-8 h-8 rounded-full bg-base-200 flex items-center justify-center text-xs font-semibold text-base-content/50 flex-shrink-0">
							{report.employeeName.charAt(0)}
						</div>
						<div class="flex-1 min-w-0">
							<p class="font-medium text-sm text-base-content truncate">
								{report.employeeName}
							</p>
							<p class="text-xs text-base-content/40">
								{reportGoals} metas asignadas
							</p>
						</div>
						<div class="flex items-center gap-2.5 flex-shrink-0">
							<div class="w-16 h-1.5 bg-base-200 rounded-full overflow-hidden">
								<div
									class="h-full rounded-full transition-all {reportStatus.tone === 'success' ? 'bg-success/70' : reportStatus.tone === 'info' ? 'bg-info/70' : reportStatus.tone === 'accent' ? 'bg-accent/70' : 'bg-primary/70'}"
									style="width: {reportProgress}%"
								></div>
							</div>
							<span class="text-[11px] font-medium px-2 py-0.5 rounded-full whitespace-nowrap {reportStatus.tone === 'success' ? 'bg-success/10 text-success' : reportStatus.tone === 'info' ? 'bg-info/10 text-info' : reportStatus.tone === 'accent' ? 'bg-accent/10 text-accent' : 'bg-primary/10 text-primary'}">
								{reportStatus.label}
							</span>
						</div>
					</li>
				{/each}
			</ul>
		</section>
	{/if}
</div>
