<script lang="ts">
    import {
        getRoot,
        getChildren,
        getDescendants,
    } from "$lib/stores/orgHierarchyStore.svelte";
    import { getCompetencyRatings } from "$lib/stores/evaluationStore.svelte";
    import {
        PROFILE_LABELS,
        type EvaluationProfile,
    } from "$lib/types/evaluation";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { ChevronRight, Star } from "@lucide/svelte";

    const PROFILE_NODE_ID: Partial<Record<EvaluationProfile, string>> = {
        'director-general': 'emp-dg-01',
        director: 'emp-director-01',
        jefe: 'emp-jefe-01',
        rh: 'emp-rh-01',
    };

    const profile = $derived(getProfile());

    const scopeNodes = $derived(() => {
        switch (profile) {
            case 'jefe': {
                const nodeId = PROFILE_NODE_ID[profile];
                if (!nodeId) return [];
                return getChildren(nodeId);
            }
            case 'director': {
                const nodeId = PROFILE_NODE_ID[profile];
                if (!nodeId) return [];
                return getDescendants(nodeId);
            }
            case 'director-general':
            case 'rh': {
                const root = getRoot();
                return [root, ...getDescendants(root.id)];
            }
            default:
                return [];
        }
    });

    const employees = $derived(() => {
        const nodes = scopeNodes();
        return nodes.map((node) => {
            const ratings = getCompetencyRatings(node.id);
            const selfRatings = ratings.filter((r) => r.selfRating != null);
            const rhRatings = ratings.filter((r) => r.rhRating != null);

            const selfAvg =
                selfRatings.length > 0
                    ? selfRatings.reduce(
                          (sum, r) => sum + (r.selfRating ?? 0),
                          0,
                      ) / selfRatings.length
                    : null;
            const rhAvg =
                rhRatings.length > 0
                    ? rhRatings.reduce((sum, r) => sum + (r.rhRating ?? 0), 0) /
                      rhRatings.length
                    : null;

            const status =
                ratings.length === 0
                    ? "sin-datos"
                    : rhRatings.length > 0 && selfRatings.length > 0
                      ? "completada"
                      : selfRatings.length > 0
                        ? "autoevaluacion"
                        : "pendiente";

            return {
                id: node.id,
                name: node.name,
                profileId: node.profileId,
                profileLabel:
                    PROFILE_LABELS[
                        node.profileId as keyof typeof PROFILE_LABELS
                    ] ?? node.profileId,
                selfAvg,
                rhAvg,
                ratingsCount: ratings.length,
                status,
            };
        });
    });

    function formatAvg(avg: number | null): string {
        if (avg === null) return "—";
        return avg.toFixed(1);
    }

    function statusBadge(status: string): { label: string; class: string } {
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
</script>

<svelte:head>
    <title>Resultados de competencias — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
    <!-- Header -->
    <div class="flex items-center gap-3">
        <div>
            <h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
                <Star class="w-6 h-6" />
                Resultados de competencias
            </h1>
            <p class="text-sm text-base-content/50">
                Vista general de competencias por empleado
            </p>
        </div>
    </div>

    <!-- Employee table -->
    <div class="card bg-base-100 shadow-sm border border-base-200">
        <div class="overflow-x-auto">
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th class="text-xs font-semibold text-base-content/60">Empleado</th>
                        <th class="text-xs font-semibold text-base-content/60">Perfil</th>
                        <th class="text-xs font-semibold text-base-content/60 text-center"
                            >Autoevaluación</th
                        >
                        <th class="text-xs font-semibold text-base-content/60 text-center">RH</th>
                        <th class="text-xs font-semibold text-base-content/60 text-center">Estado</th>
                        <th class="w-10"></th>
                    </tr>
                </thead>
                <tbody>
                    {#each employees() as emp (emp.id)}
                        {@const badge = statusBadge(emp.status)}
                        <tr class="hover:bg-base-200/50 transition-colors">
                            <td>
                                <div class="flex items-center gap-2.5">
                                    <div class="avatar avatar-placeholder">
                                        <div
                                            class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center"
                                        >
                                            <span class="text-xs font-bold">
                                                {emp.name
                                                    .charAt(0)
                                                    .toUpperCase()}
                                            </span>
                                        </div>
                                    </div>
                                    <span class="font-medium text-sm"
                                        >{emp.name}
                                    </span>
                                </div>
                            </td>
                            <td>
                                <span class="text-xs text-base-content/50"
                                    >{emp.profileLabel}</span
                                >
                            </td>
                            <td class="text-center">
                                <span
                                    class="text-sm font-mono {emp.selfAvg !==
                                    null
                                        ? 'text-base-content'
                                        : 'text-base-content/30'}"
                                >
                                    {formatAvg(emp.selfAvg)}
                                </span>
                            </td>
                            <td class="text-center">
                                <span
                                    class="text-sm font-mono {emp.rhAvg !== null
                                        ? 'text-base-content'
                                        : 'text-base-content/30'}"
                                >
                                    {formatAvg(emp.rhAvg)}
                                </span>
                            </td>
                            <td class="text-center">
                                <span class="badge badge-sm {badge.class}"
                                    >{badge.label}
                                </span>
                            </td>
                            <td>
                                <a
                                    href="/evaluacion/9x9/competencias/{emp.id}"
                                    class="btn btn-ghost btn-square btn-xs"
                                    aria-label="Ver competencias de {emp.name}"
                                >
                                    <ChevronRight class="w-4 h-4" />
                                </a>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    </div>
</div>
