<script lang="ts">
    import { onMount } from "svelte";
    import {
        getEvaluationStatus,
        isLoading,
        getError,
        load as loadEvaluations,
    } from "$lib/stores/evaluationStore.svelte";
    import {
        getPillars,
        getCompetenciesByPillar,
    } from "$lib/stores/competencyStore.svelte";
    import { getNodeById } from "$lib/stores/orgHierarchyStore.svelte";
    import { getGoals } from "$lib/stores/goalsStore.svelte";
    import { getAssignmentStatus as getMisAssignmentStatus, getItems as getMisItems } from "$lib/stores/misEvaluadosStore.svelte";
    import PageSkeleton from "$lib/components/ui/PageSkeleton.svelte";
    import ErrorState from "$lib/components/ui/ErrorState.svelte";
    import { PROFILE_LABELS, PHASE_LABELS } from "$lib/types/evaluation";
    import { titleCase } from "$lib/utils/text";
    import { getActivePhase } from "$lib/api/cycle.svelte";
    import type { EmployeeAssignment } from "$lib/types/goal";
    import type { Snippet } from "svelte";
    import { load as reloadRhEvaluados } from "$lib/stores/rhEvaluadosStore.svelte";
    import { FileDown, ChevronRight, Pencil } from "@lucide/svelte";
    import { toCsv } from "$lib/utils/export";
    import ChangeDepartmentProfileModal from "./ChangeDepartmentProfileModal.svelte";

    // ponytail: remove when OpenAPI schema includes profileName
    interface EmployeeListItemRow {
        id: string;
        firstName: string;
        lastName: string;
        profileName: string;
        isActive: boolean;
    }

    interface Props {
        employees?: EmployeeAssignment[];
        rows?: EmployeeListItemRow[];
        mode?: "rh" | "manager";
        onSelect: (employeeId: string) => void;
        selectedEmployeeId?: string;
        disabled?: boolean;
        detail?: Snippet;
        competencyRatings?: Map<
            string,
            { selfAvg: number; rhAvg: number; status: string }
        >;
        competencyDetailHref?: (employeeId: string) => string;
    }

    let {
        employees = [],
        rows = [],
        onSelect,
        selectedEmployeeId = "",
        disabled = false,
        mode = "manager",
        detail,
        competencyRatings,
    }: Props = $props();

    const loadingEval = $derived(isLoading());
    const errorEval = $derived(getError());

    onMount(() => {
        loadEvaluations();
    });

    let searchQuery = $state("");
    let changeTargetId = $state<string | null>(null);

    const pillars = $derived(getPillars());
    const allCompetencies = $derived(
        pillars.flatMap((p) => getCompetenciesByPillar(p.id)),
    );
    const goals = $derived(getGoals());

    function getAssignmentStatus(employeeId: string): 'no_iniciado' | 'borrador' | 'enviada' {
        // primary: misEvaluadosStore (evaluatees endpoint)
        const misItems = getMisItems();
        const hit = misItems.find((x) => x.id === employeeId) as unknown as { assignmentStatus?: string } | undefined;
        if (hit?.assignmentStatus) {
            const s = hit.assignmentStatus;
            if (s === 'no_iniciado' || s === 'borrador' || s === 'enviada') return s;
        }
        // fallback via exported helper (covers empty array case)
        const viaHelper = getMisAssignmentStatus(employeeId);
        if (misItems.length > 0) return viaHelper;
        // legacy fallback: employees prop may carry status
        const a = employees.find((x) => x.employeeId === employeeId) as EmployeeAssignment | undefined;
        const raw = a?.status as string | undefined;
        if (raw === 'enviada' || raw === 'borrador' || raw === 'no_iniciado') return raw as 'no_iniciado' | 'borrador' | 'enviada';
        return viaHelper;
    }

    function assignmentBadge(status: 'no_iniciado' | 'borrador' | 'enviada'): { label: string; cls: string } {
        if (status === 'enviada') return { label: 'Enviada', cls: 'badge-success' };
        if (status === 'borrador') return { label: 'Borrador', cls: 'badge-warning' };
        return { label: 'No iniciado', cls: 'badge-ghost' };
    }

    function isDraftOrBeginning(rowId: string): boolean {
        const s = getAssignmentStatus(rowId)
        return s === 'no_iniciado' || s === 'borrador'
    }

    function isFormulacionPhase(): boolean {
        const p = getActivePhase();
        if (!p) return false;
        const s = p.toLowerCase();
        return s.includes("formul") || s.includes("inicio") || s.includes("planea") || s.includes("asignacion");
    }

    const filteredEmployees = $derived(
        searchQuery.trim() === ""
            ? employees
            : employees.filter(
                  (e) =>
                      e.employeeName
                          .toLowerCase()
                          .includes(searchQuery.toLowerCase()) ||
                      getProfileLabel(e.employeeId)
                          .toLowerCase()
                          .includes(searchQuery.toLowerCase()),
              ),
    );

    const progressMap = $derived(
        new Map(
            filteredEmployees.map((emp) => {
                const empGoals = goals.filter((g) =>
                    emp.goalIds.includes(g.id),
                );
                const totalTarget = empGoals.reduce(
                    (sum, g) => sum + g.targetValue,
                    0,
                );
                const totalProgress = empGoals.reduce(
                    (sum, g) => sum + (g.progress ?? 0),
                    0,
                );
                const pct =
                    totalTarget > 0
                        ? Math.min((totalProgress / totalTarget) * 100, 100)
                        : null;
                return [emp.employeeId, pct] as const;
            }),
        ),
    );

    const currentPhase = $derived(getActivePhase() ?? "inicio-anio");

    const completionSummary = $derived({
        total: filteredEmployees.length,
        completed: filteredEmployees.filter((e) =>
            hasCompletedPhase(e.employeeId),
        ).length,
    });

    function hasCompletedPhase(employeeId: string): boolean {
        const assignment = employees.find((e) => e.employeeId === employeeId);
        const empGoals = goals.filter((g) =>
            assignment?.goalIds.includes(g.id),
        );

        switch (currentPhase) {
            case "inicio-anio":
                // Completed if has goals assigned
                return (
                    assignment !== undefined && assignment.goalIds.length > 0
                );
            case "medio-anio":
                // Completed if has updated progress on any goal
                return empGoals.some(
                    (g) => g.progress !== undefined && g.progress > 0,
                );
            case "fin-anio":
                // Completed if all competencies rated and all goals closed
                return getStatus(employeeId) === "completed";
            default:
                return false;
        }
    }

    function getProfileLabel(employeeId: string): string {
        const node = getNodeById(employeeId);
        if (!node) return "—";
        return (
            PROFILE_LABELS[node.profileId as keyof typeof PROFILE_LABELS] ??
            node.profileId
        );
    }

    function getStatus(employeeId: string) {
        const assignment = employees.find((e) => e.employeeId === employeeId);
        return getEvaluationStatus(
            employeeId,
            allCompetencies.length,
            assignment?.goalIds ?? [],
        );
    }

    const statusLabelMap: Record<string, string> = {
        pending: "Pendiente",
        "in-progress": "En progreso",
        completed: "Completada",
    };

    function competencyStatusBadge(status: string): {
        label: string;
        class: string;
    } {
        switch (status) {
            case "completada":
                return { label: "Completada", class: "badge-success" };
            case "autoevaluacion":
                return { label: "Autoevaluación", class: "badge-warning" };
            case "pendiente":
                return { label: "Pendiente", class: "badge-ghost" };
            default:
                return { label: "Sin datos", class: "badge-ghost" };
        }
    }

    function handleExportCsv() {
        toCsv(
            filteredEmployees.map((emp) => ({
                Empleado: emp.employeeName,
                Perfil: getProfileLabel(emp.employeeId),
                "Progreso global %":
                    progressMap.get(emp.employeeId) !== null
                        ? `${Math.round(progressMap.get(emp.employeeId)!)}%`
                        : "",
                Estado: statusLabelMap[getStatus(emp.employeeId)],
            })),
            "evaluaciones.csv",
        );
    }
</script>

<div class="flex flex-col gap-6">
    {#if !selectedEmployeeId}
        <!-- Summary + export -->
        <div class="flex items-center gap-6">
            {#if completionSummary.total > 0}
                <div class="flex items-center gap-2">
                    <span class="text-xs font-semibold text-base-content/60"
                        >{PHASE_LABELS[currentPhase]}:</span
                    >
                    <span
                        class="badge badge-sm {completionSummary.completed /
                            completionSummary.total >=
                        0.8
                            ? 'badge-success'
                            : 'badge-warning'}"
                    >
                        {completionSummary.completed} de {completionSummary.total}
                        completaron
                    </span>
                </div>
            {/if}

            <button
                class="btn btn-outline btn-sm"
                disabled={filteredEmployees.length === 0}
                onclick={handleExportCsv}
            >
                <FileDown class="w-4 h-4" />
                Exportar CSV
            </button>
        </div>
    {/if}

    {#if selectedEmployeeId}
        {@render detail?.()}
    {:else if loadingEval}
        <PageSkeleton variant="table" rows={Math.max(employees.length, 3)} />
    {:else if (errorEval && employees.length === 0)}
        <ErrorState message={errorEval} onretry={loadEvaluations} />
    {:else if rows.length === 0}
        <p class="text-sm text-base-content/30 italic text-center py-8">
            Sin empleados para mostrar
        </p>
    {:else}
        <div class="overflow-x-auto">
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th class="text-xs font-semibold text-base-content/60">Empleado</th>
                        <th class="text-xs font-semibold text-base-content/60">Perfil</th>
                        <th class="text-xs font-semibold text-base-content/60 text-center">Estado Metas</th>
                        <th class="text-xs font-semibold text-base-content/60 text-center">Autoevaluación</th>
                        <th class="text-xs font-semibold text-base-content/60 text-center">Evaluación</th>
                        <th class="text-xs font-semibold text-base-content/60 text-center">Estado</th>
                        <th class="text-center text-xs font-semibold text-base-content/60">Acciones</th>
                    </tr>
                </thead>
                <tbody>
                    {#each rows as row (row.id)}
                        {@const cr = competencyRatings?.get(row.id)}
                        {@const rowStatus = getAssignmentStatus(row.id)}
                        {@const isDraft = isDraftOrBeginning(row.id)}
                        {@const _badge = assignmentBadge(rowStatus)}
                        <tr class="hover:bg-base-200">
                            <td>
                                <div class="flex items-center gap-2.5">
                                    <div class="avatar avatar-placeholder">
                                        <div
                                            class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center"
                                        >
                                            <span class="text-xs font-bold">
                                                {row.firstName
                                                    .charAt(0)
                                                    .toUpperCase()}
                                            </span>
                                        </div>
                                    </div>
                                    <span class="font-medium text-sm"
                                        >{row.firstName}
                                        {row.lastName}</span
                                    >
                                </div>
                            </td>
                            <td>
                                <span class="text-xs text-base-content/50"
                                    >{titleCase(row.profileName)}</span
                                >
                            </td>
                            <td class="text-center">
                                <span class="badge badge-sm {_badge.cls}">
                                    {_badge.label}
                                </span>
                            </td>
                            <td class="text-center">
                                <span
                                    class="text-sm font-mono {cr?.selfAvg !=
                                    null
                                        ? 'text-base-content'
                                        : 'text-base-content/30'}"
                                >
                                    {cr?.selfAvg?.toFixed(1) ?? "—"}
                                </span>
                            </td>
                            <td class="text-center">
                                <span
                                    class="text-sm font-mono {cr?.rhAvg != null
                                        ? 'text-base-content'
                                        : 'text-base-content/30'}"
                                >
                                    {cr?.rhAvg?.toFixed(1) ?? "—"}
                                </span>
                            </td>
                            <td class={competencyRatings ? "text-center" : "text-center"}>
                                {#if cr?.status}
                                    {@const badge = competencyStatusBadge(
                                        cr.status,
                                    )}
                                    <span class="badge badge-sm {badge.class}"
                                        >{badge.label}</span
                                    >
                                {:else}
                                    <span class="text-xs text-base-content/30"
                                        >—</span
                                    >
                                {/if}
                            </td>
                            <td>
                                <div class="flex items-center justify-end gap-1">
                                    {#if mode === 'rh' && currentPhase === 'inicio-anio'}
                                        <button
                                            type="button"
                                            class="btn btn-outline btn-xs"
                                            onclick={() => (changeTargetId = row.id)}
                                        >
                                            <Pencil class="w-3 h-3" />
                                            Cambiar
                                        </button>
                                    {/if}
                                    {#if isFormulacionPhase()}
                                        <a
                                            href={`/objetivos/asignacion?empId=${row.id}`}
                                            class="btn btn-outline btn-xs gap-1"
                                            title="Ver/Editar Metas"
                                        >
                                            Metas
                                        </a>
                                    {/if}
                                    <button
                                        type="button"
                                        class="btn btn-primary btn-xs"
                                        onclick={() => onSelect(row.id)}
                                        {disabled}
                                    >
                                        Evaluar
                                    </button>
                                    <a
                                        href={`/evaluacion/9x9/competencias/${row.id}`}
                                        class="btn btn-ghost btn-square btn-xs"
                                        aria-label="Ver competencias de {row.firstName} {row.lastName}"
                                    >
                                        <ChevronRight class="w-4 h-4" />
                                    </a>
                                </div>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    {/if}
</div>

{#if changeTargetId}
    <ChangeDepartmentProfileModal
        employeeId={changeTargetId}
        onsave={() => {
            changeTargetId = null;
            loadEvaluations();
            reloadRhEvaluados();
        }}
        onclose={() => {
            changeTargetId = null;
        }}
    />
{/if}
