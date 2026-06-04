<script lang="ts">
    import { Save, LayoutGrid, Edit3 } from "@lucide/svelte";
    import type { EvaluationProfile } from "$lib/types/evaluation";
    import { EVALUATION_PROFILES, PROFILE_LABELS } from "$lib/types/evaluation";
    import {
        getPillars,
        getCompetencies,
        getLevelDefinitions,
        getCompetencyAcceptanceLevel,
        setCompetencyAcceptanceLevel,
    } from "$lib/stores/competencyStore.svelte";
    import LevelDefinitionModal from "./LevelDefinitionModal.svelte";
    import AcceptanceLevelSummaryModal from "./AcceptanceLevelSummaryModal.svelte";
    import CustomSelect from "$lib/components/ui/CustomSelect.svelte";

    let selectedProfile = $state<EvaluationProfile>("colaborador");
    let successMsg = $state("");
    let hasChanges = $state(false);
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
        setCompetencyAcceptanceLevel(
            competencyId,
            selectedProfile,
            Number(newLevel) as 1 | 2 | 3 | 4 | 5,
        );
        hasChanges = true;
    }

    function handleSave() {
        hasChanges = false;
        successMsg = "Niveles de aceptación guardados correctamente.";
        setTimeout(() => (successMsg = ""), 3000);
    }
</script>

<div>
    <!-- Header with tabs and action buttons -->
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
        <div
            class="tabs tabs-lift tabs-sm"
            role="tablist"
            aria-label="Perfiles de evaluación"
        >
            {#each EVALUATION_PROFILES as profile (profile)}
                <button
                    role="tab"
                    class="tab"
                    class:tab-active={selectedProfile === profile}
                    onclick={() => {
                        selectedProfile = profile;
                    }}
                    aria-selected={selectedProfile === profile}
                >
                    {PROFILE_LABELS[profile]}
                </button>
            {/each}
        </div>
        <div class="flex items-center gap-2">
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
    </div>

    <!-- Success alert -->
    {#if successMsg}
        <div class="alert alert-success mb-4 text-sm" role="status">
            <span>{successMsg}</span>
        </div>
    {/if}

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
                    class="text-sm font-semibold text-base-content mb-3 flex items-center gap-2"
                >
                    <span class="w-1.5 h-5 rounded bg-primary"></span>
                    {pillar.name}
                </legend>
                <div class="space-y-2">
                    {#each pillarComps as competency (competency.id)}
                        <div
                            class="flex items-center justify-between gap-4 p-3 rounded-lg bg-base-200/50 hover:bg-base-200 transition-colors"
                        >
                            <div class="flex-1 min-w-0">
                                <span class="text-sm font-medium"
                                    >{competency.name}</span
                                >
                                <p
                                    class="text-xs text-base-content/50 truncate"
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

    <!-- Save button -->
    <div class="mt-6 flex justify-end">
        <button
            class="btn btn-primary btn-sm"
            onclick={handleSave}
            disabled={!hasChanges}
        >
            <Save class="w-4 h-4" />
            Guardar cambios
        </button>
    </div>
</div>

<LevelDefinitionModal
    open={showLevelDefModal}
    onClose={() => (showLevelDefModal = false)}
/>
<AcceptanceLevelSummaryModal
    open={showSummary}
    onClose={() => (showSummary = false)}
/>
