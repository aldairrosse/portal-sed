<script lang="ts">
    import { LayoutGrid, Edit3 } from "@lucide/svelte";
    import type { EvaluationProfile } from "$lib/types/evaluation";
    import { PROFILE_LABELS } from "$lib/types/evaluation";
    import {
        getProfiles,
        getPillars,
        getCompetencies,
        getLevelDefinitions,
        getCompetencyAcceptanceLevel,
        setCompetencyAcceptanceLevelOptimistic,
        reload,
    } from "$lib/stores/competencyStore.svelte";
    import { error as toastError } from "$lib/stores/notifications.svelte";
    import LevelDefinitionModal from "./LevelDefinitionModal.svelte";
    import AcceptanceLevelSummaryModal from "./AcceptanceLevelSummaryModal.svelte";
    import CustomSelect from "$lib/components/ui/CustomSelect.svelte";

    const profiles = $derived(getProfiles());
    let selectedProfile = $state<EvaluationProfile>("colaborador");
    let showLevelDefModal = $state(false);
    let showSummary = $state(false);

    const levels = [1, 2, 3, 4, 5] as const;

    const pillars = $derived(getPillars());
    const competencies = $derived(getCompetencies());
    const levelDefs = $derived(getLevelDefinitions());

    const levelOptions = $derived(
        levels.map((l) => {
            const def = levelDefs.find((d) => d.level === l);
            return {
                value: String(l),
                label: `N${l} - ${def?.label ?? "Nivel " + l}`,
            };
        }),
    );

    function getCompetenciesByPillar(pillarId: string) {
        return competencies.filter((c) => c.pillarId === pillarId);
    }

    function getLevelForCompetency(competencyId: string): string {
        const cal = getCompetencyAcceptanceLevel(competencyId, selectedProfile);
        return String(cal?.level ?? 3);
    }

    function handleLevelChange(competencyId: string, newLevel: string) {
        const level = Number(newLevel) as 1 | 2 | 3 | 4 | 5;
        // Defer to next microtask so the popover closes before the store
        // mutation triggers a re-render of every CustomSelect.
        queueMicrotask(() => {
            const { revert, commit } = setCompetencyAcceptanceLevelOptimistic(
                competencyId,
                selectedProfile,
                level,
            );
            commit().catch(() => {
                revert();
                toastError("Error al guardar el nivel. Se revirtió el cambio.");
            });
        });
    }
</script>

<div>
    <!-- Action buttons -->
    <div class="flex items-center justify-end gap-2 mb-3">
        <button
            class="btn btn-ghost btn-sm"
            onclick={() => (showLevelDefModal = true)}
            aria-label="Editar definiciones de nivel"
        >
            <Edit3 class="w-4 h-4" />
            Editar definiciones de nivel
        </button>
        <button
            class="btn btn-ghost btn-sm"
            onclick={() => (showSummary = true)}
            aria-label="Vista resumen"
        >
            <LayoutGrid class="w-4 h-4" />
            Vista resumen
        </button>
    </div>

    <!-- Tabs -->
    <div
        class="tabs tabs-lift tabs-sm mb-4"
        role="tablist"
        aria-label="Perfiles de evaluación"
    >
        {#each profiles as profile (profile.name)}
            <button
                role="tab"
                class="tab"
                class:tab-active={selectedProfile === profile.name}
                onclick={() => {
                    selectedProfile = profile.name;
                }}
                aria-selected={selectedProfile === profile.name}
            >
                {PROFILE_LABELS[profile.name] ?? profile.name}
            </button>
        {/each}
    </div>

    <!-- Selected profile and description -->
    <div class="mb-6">
        <h4 class="font-semibold text-base mb-2">
            {PROFILE_LABELS[selectedProfile]}
        </h4>
        <p class="text-sm text-base-content/60">
            Niveles de aceptación para el perfil. Asigne un nivel de aceptación
            a cada competencia.
        </p>
    </div>

    <!-- Competencies grouped by pillar -->
    <div class="space-y-6">
        {#each pillars as pillar (pillar.id)}
            {@const pillarComps = getCompetenciesByPillar(pillar.id)}
            <fieldset>
                <legend
                    class="text-xs tracking-wide font-semibold text-base-content/50 mb-3"
                >
                    {pillar.name}
                </legend>
                <div class="space-y-2">
                    {#each pillarComps as competency (competency.id)}
                        <div
                            class="flex items-center justify-between gap-4 p-3 rounded-lg bg-base-200/50 hover:bg-base-200 transition-colors w-full"
                        >
                            <div class="flex-1 min-w-0">
                                <span class="text-sm font-medium block"
                                    >{competency.name}</span
                                >
                                <p
                                    class="text-xs text-base-content/50 break-words"
                                >
                                    {competency.description}
                                </p>
                            </div>
                            <CustomSelect
                                options={levelOptions}
                                value={getLevelForCompetency(competency.id)}
                                onChange={(v) =>
                                    handleLevelChange(competency.id, v)}
                                ariaLabel="Nivel para {competency.name}"
                            />
                        </div>
                    {/each}
                </div>
            </fieldset>
        {/each}
    </div>

    <!-- Empty state -->
    {#if competencies.length === 0}
        <div class="text-center py-12 text-base-content/50 text-sm">
            No hay competencias registradas. Agregue competencias para asignar
            niveles de aceptación.
        </div>
    {/if}
</div>

<LevelDefinitionModal
    open={showLevelDefModal}
    onClose={() => (showLevelDefModal = false)}
    onSaved={() => { showLevelDefModal = false; reload(); }}
/>
<AcceptanceLevelSummaryModal
    open={showSummary}
    onClose={() => (showSummary = false)}
/>
