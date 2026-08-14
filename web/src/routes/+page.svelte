<script lang="ts">
    import { onMount } from "svelte";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { getActivePhase } from "$lib/api/cycle.svelte";
    import { getSession } from "$lib/api/session.svelte";
    import type { CyclePhase } from "$lib/types/evaluation";
    import {
        getGoals,
        getCategories,
        getAssignments,
        getInstitutionalGoals,
    } from "$lib/stores/goalsStore.svelte";
    import ProgressChart from "$lib/components/ProgressChart.svelte";
    import {
        load as loadCompetencies,
        getPillars,
        getCompetencies,
        getLevelDefinitions,
        getCompetencyAcceptanceLevelsByProfile,
    } from "$lib/stores/competencyStore.svelte";
    import { Users, User } from "@lucide/svelte";
    import type { Goal } from "$lib/types/goal";
    import Icon from "$lib/components/ui/Icon.svelte";
    import { pillarColors } from "$lib/utils/pillarColors";
    import calendar from "$lib/assets/calendario_1.png";

    const session = $derived(getSession());
    const profile = $derived(getProfile());
    const phase = $derived(getActivePhase() ?? "inicio-anio");
    const user = $derived(session.user);

    const assignments = $derived(getAssignments());
    const allGoals = $derived(getGoals());
    const allCategories = $derived(getCategories());

    const pillars = $derived(getPillars());
    const competencies = $derived(getCompetencies());
    const levelDefs = $derived(getLevelDefinitions());
    const competencyLevels = $derived(
        getCompetencyAcceptanceLevelsByProfile(profile),
    );

    onMount(() => {
        loadCompetencies();
    });

    const myAssignment = $derived(
        assignments.find((a) => a.profileId === profile),
    );
    const myEmployeeId = $derived(myAssignment?.employeeId ?? "");
    const myManager = $derived(
        myAssignment?.managerId
            ? assignments.find((a) => a.employeeId === myAssignment.managerId)
            : null,
    );
    const myDirectReports = $derived(
        myEmployeeId
            ? assignments.filter((a) => a.managerId === myEmployeeId)
            : [],
    );

    const hasReports = $derived(
        [
            "jefe",
            "gerente-tienda",
            "divisional",
            "regional",
            "director",
            "director-general",
        ].includes(profile),
    );

    const myGoals = $derived(
        myAssignment
            ? myAssignment.goalIds
                  .map((id) => allGoals.find((g) => g.id === id))
                  .filter((g): g is Goal => Boolean(g))
            : [],
    );

    const institutionalGoals = $derived(getInstitutionalGoals());

    const personalStats = $derived({
        evaluated: myGoals.filter((g) => g.progress !== undefined).length,
        total: myGoals.length,
    });

    const institutionalStats = $derived({
        global: institutionalGoals.filter((g) => g.source === "global"),
        shared: institutionalGoals.filter((g) => g.source === "shared"),
    });

    const globalStats = $derived(
        institutionalStats.global.length > 0
            ? {
                  evaluated: institutionalStats.global.filter(
                      (g) => g.progressPercent !== undefined,
                  ).length,
                  total: institutionalStats.global.length,
              }
            : undefined,
    );

    const sharedStats = $derived(
        institutionalStats.shared.length > 0
            ? {
                  evaluated: institutionalStats.shared.filter(
                      (g) => g.progressPercent !== undefined,
                  ).length,
                  total: institutionalStats.shared.length,
              }
            : undefined,
    );

    const myCompetencies = $derived(
        competencies.map((c) => {
            const lvl = competencyLevels.find(
                (cl) => cl.competencyId === c.id && cl.profileId === profile,
            );
            const def = levelDefs.find((ld) => ld.level === (lvl?.level ?? 0));
            const pillar = pillars.find((p) => p.id === c.pillarId);
            return {
                ...c,
                pillarName: pillar?.name ?? "",
                level: lvl?.level ?? 0,
                levelLabel: def?.label ?? "—",
            };
        }),
    );

    const competenciesByPillar = $derived(
        pillars.map((p) => ({
            pilar: p,
            items: myCompetencies.filter((c) => c.pillarId === p.id),
        })),
    );

    const today = new Date();
    const year = today.getFullYear();

    const phaseTimeSteps = $derived<
        Record<
            CyclePhase,
            { start: Date; end: Date; label: string; index: number }
        >
    >({
        "inicio-anio": {
            start: new Date(year, 0, 1),
            end: new Date(year, 3, 30),
            label: "Inicio de año",
            index: 1,
        },
        "medio-anio": {
            start: new Date(year, 4, 1),
            end: new Date(year, 8, 30),
            label: "Medio de año",
            index: 2,
        },
        "fin-anio": {
            start: new Date(year, 9, 1),
            end: new Date(year, 11, 31),
            label: "Fin de año",
            index: 3,
        },
    });

    const phaseGuidance = $derived<Record<CyclePhase, string>>({
        "inicio-anio":
            profile === "director-general"
                ? "Visualiza las metas estratégicas de la organización para el ciclo."
                : "Fija tus objetivos y KPIs con tu jefe. Define metas claras para el ciclo.",
        "medio-anio":
            profile === "director-general"
                ? "Visualiza el avance de los objetivos organizacionales."
                : "Revisa el avance de tus objetivos. Ajusta lo necesario antes del cierre.",
        "fin-anio":
            profile === "director-general"
                ? "Visualiza el desempeño de la organización."
                : "Evalúa tu desempeño y completa la autoevaluación.",
    });

    function getPhaseStatus(phase_status: CyclePhase) {
        const step = phaseTimeSteps[phase_status];
        if (!step) return "Desconocido";
        if (phase === phase_status) return "Actual";
        const current_index = phaseTimeSteps[phase].index;
        if (step.index < current_index) return "Completado";
        if (step.index - current_index >= 2) return "Pendiente";
        return "Próximo";
    }

    function isNextPhase(phase_status: CyclePhase): boolean {
        const step = phaseTimeSteps[phase_status];
        if (!step) return false;
        const current_index = phaseTimeSteps[phase].index;
        return step.index > current_index;
    }

    function getGoalCategoryName(goal: Goal): string {
        const cat = allCategories.find((c) => c.id === goal.categoryId);
        return cat?.name ?? "";
    }
</script>

<svelte:head>
    <title>Inicio — SED</title>
</svelte:head>

<div class="mx-auto space-y-6">
    <!-- Welcome + profile row -->
    <header>
        <p class="text-sm text-base-content/80">Hola,</p>
        <h1 class="mt-1">{user?.name}</h1>
        <div class="flex flex-wrap items-center gap-3 mt-3">
            <span
                class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary/10 text-primary text-sm font-medium"
            >
                <Icon name="business-bag" size={18} />
                {user?.jobTitle}
            </span>
            {#if hasReports && myDirectReports.length > 0}
                <span
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-base-200 text-base-content/60 text-xs font-medium"
                >
                    <Users class="w-3 h-3" />
                    Lidera a {myDirectReports.length}
                    {myDirectReports.length === 1 ? "persona" : "personas"}
                </span>
            {:else if myManager}
                <span
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-base-200 text-base-content/60 text-xs font-medium"
                >
                    <User class="w-3 h-3" />
                    Reporta a {myManager.employeeName}
                </span>
            {/if}
        </div>
    </header>

    <!-- Cycle status + Progress + Quick access + Goals (responsive row) -->
    <div class="grid grid-cols-1 lg:grid-cols-10 gap-6">
        <!-- Cycle -->
        <section
            class="bg-(--color-base) rounded-xl {profile === 'director-general'
                ? 'lg:col-span-10'
                : 'lg:col-span-6'}"
        >
            <h2
                class="text-sm font-bold text-base-content px-4 py-3 tracking-wide"
            >
                Ciclo {year}
            </h2>
            <!-- Timeline -->
            <div class="w-full py-3 flex items-center">
                {#each Object.entries(phaseTimeSteps) as [key, step] (key)}
                    {@const currentKey = key as CyclePhase}
                    {@const status = getPhaseStatus(currentKey)}
                    {@const isNext = isNextPhase(currentKey)}
                    <div class="flex-1 flex flex-col items-center gap-1">
                        <div
                            class="relative w-full flex flex-col items-center justify-center"
                        >
                            <hr
                                class="absolute w-full rounded-full h-[1px] border-0 z-0 {isNext
                                    ? 'bg-base-content/40'
                                    : 'bg-primary h-[2px]'}"
                            />
                            <div
                                class="z-1 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold {isNext
                                    ? 'bg-(--color-base) text-base-content/40 border border-base-content/40'
                                    : 'bg-primary text-primary-content'}"
                            >
                                {phaseTimeSteps[currentKey].index}
                            </div>
                        </div>
                        <span
                            class="text-xs font-bold {isNext
                                ? 'text-base-content/70'
                                : 'text-primary'}">{step.label}</span
                        >
                        <span
                            class="text-xs {isNext
                                ? 'text-base-content/70'
                                : 'text-base-content'}">{status}</span
                        >
                    </div>
                {/each}
            </div>
            <p class="p-4 text-sm text-base-content/70">
                {phaseGuidance[phase]}
            </p>
        </section>

        {#if profile !== "director-general"}
            <!-- Progress -->
            <section class="lg:col-span-4 bg-(--color-base) rounded-xl">
                <h2
                    class="text-sm font-bold text-base-content px-4 py-3 tracking-wide mb-3"
                >
                    Tu progreso del ciclo
                </h2>
                <ProgressChart
                    personal={personalStats}
                    global={globalStats}
                    shared={sharedStats}
                />
            </section>

            <!-- Quick access -->
            <section class="lg:col-span-4 bg-(--color-base) rounded-xl">
                <h2
                    class="text-sm font-bold text-base-content px-4 py-3 tracking-wide"
                >
                    Accesos rápidos
                </h2>
                <div class="grid grid-cols-2 gap-3 px-4 pt-3 pb-6 min-h-40">
                    <a
                        href="/mi-evaluacion"
                        class="flex flex-col items-center gap-2 p-3 rounded-xl shadow-lg hover:bg-base-200 transition-colors text-center"
                    >
                        <Icon
                            name="file-check"
                            size={44}
                            class="text-primary flex-1"
                        />
                        <span class="text-xs font-bold text-base-content pb-2"
                            >Mi evaluación
                        </span>
                    </a>
                    <a
                        href="/objetivos/asignacion"
                        class="flex flex-col items-center gap-2 p-3 rounded-xl shadow-lg hover:bg-base-200 transition-colors text-center"
                    >
                        <Icon
                            name="analytics-report"
                            size={44}
                            class="text-primary flex-1"
                        />
                        <span class="text-xs font-bold text-base-content pb-2"
                            >Asignación
                        </span>
                    </a>
                </div>
            </section>

            <!-- Goals (responsive row) -->
            <div class="lg:col-span-6">
                <!-- Goals -->
                <section class="bg-(--color-base) rounded-xl">
                    <div class="flex gap-4">
                        <div class="flex-1 flex-grow flex-col px-4 py-3">
                            <div class="flex items-baseline justify-between">
                                <h2
                                    class="text-sm font-bold text-base-content tracking-wide"
                                >
                                    Metas pendientes
                                </h2>
                            </div>
                            {#if myGoals.length > 0}
                                <ul class="space-y-3">
                                    {#each myGoals as goal (goal.id)}
                                        {@const categoryName =
                                            getGoalCategoryName(goal)}
                                        <li>
                                            <div
                                                class="flex items-baseline justify-between mb-1"
                                            >
                                                <span
                                                    class="font-medium text-sm text-base-content"
                                                    >{goal.name}</span
                                                >
                                                <span
                                                    class="text-xs font-mono text-primary ml-3"
                                                >
                                                    {goal.weight}%
                                                </span>
                                            </div>
                                            <div
                                                class="h-1.5 bg-base-200 rounded-full overflow-hidden"
                                            >
                                                <div
                                                    class="h-full bg-primary/70 rounded-full"
                                                    style="width: {goal.weight}%"
                                                ></div>
                                            </div>
                                            {#if categoryName}
                                                <p
                                                    class="text-[11px] text-base-content/40 mt-1"
                                                >
                                                    {categoryName}
                                                </p>
                                            {/if}
                                        </li>
                                    {/each}
                                </ul>
                            {:else}
                                <p class="text-sm text-base-content/50 pt-2">
                                    Aún no tienes metas pendientes.
                                </p>
                            {/if}
                        </div>
                        <div
                            class="flex flex-col items-center justify-center gap-2 px-4 pb-4"
                        >
                            <img
                                src={calendar}
                                alt="Calendario"
                                class="w-auto h-48 object-fit"
                            />
                            <a
                                href="/objetivos/asignacion"
                                class="text-sm text-primary hover:underline"
                            >
                                Ver calendario →
                            </a>
                        </div>
                    </div>
                </section>
            </div>
        {/if}
    </div>

    <!-- Competencies summary -->
    {#if myCompetencies.length > 0}
        <section class="bg-(--color-base) rounded-xl overflow-hidden">
            <div class="grid grid-cols-1 sm:grid-cols-4 gap-0">
                {#each competenciesByPillar as group, i (group.pilar.id)}
                    {#if group.items.length > 0}
                        <div
                            class="flex items-center gap-3 px-4 py-6"
                            style="background-color: {pillarColors(i).color}"
                        >
                            <div
                                class="w-10 h-10 rounded-full flex items-center justify-center shrink-0"
                                style="background-color: {pillarColors(i).colorDarker}"
                            >
                                <Icon
                                    name={pillarColors(i).icon}
                                    variant="colorfull"
                                    size={32}
                                    aria-hidden="true"
                                    class={i !== 0 ? 'translate-x-0.5' : ''}
                                />
                            </div>
                            <div>
                                <h3
                                    class="font-binjay text-lg font-normal text-white mb-2"
                                >
                                    {group.pilar.name}
                                </h3>
                                <ul class="space-y-1">
                                    {#each group.items.slice(0, 1) as comp (comp.id)}
                                        <li class="text-sm text-white/80">
                                            {comp.name}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        </div>
                    {/if}
                {/each}
            </div>
        </section>
    {/if}
</div>
