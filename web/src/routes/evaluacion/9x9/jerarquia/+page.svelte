<script lang="ts">
    import { onMount } from "svelte";
    import OrgHierarchyTree from "$lib/components/org-hierarchy/OrgHierarchyTree.svelte";
    import JefeEvaluateesTable from "$lib/components/org-hierarchy/JefeEvaluateesTable.svelte";
    import {
        getRoot,
        getSubtree,
        isLoading,
        getError,
        reload,
        load,
    } from "$lib/stores/orgHierarchyStore.svelte";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { getSession } from "$lib/api/session.svelte";
    import { client } from "$lib/api/client";
    import { type EvaluationProfile } from "$lib/types/evaluation";
    import type { OrgNode } from "$lib/types/org-hierarchy";
    import EmptyState from "$lib/components/ui/EmptyState.svelte";
    import PageSkeleton from "$lib/components/ui/PageSkeleton.svelte";
    import ErrorState from "$lib/components/ui/ErrorState.svelte";
    import { ArrowUpRight, Briefcase, Network, Target, Users } from "@lucide/svelte";
    import { titleCase } from "$lib/utils/text";

    // ─── Profile guard ─────────────────────────────────────────────────────

    const profile = $derived(getProfile());
    const ALLOWED_PROFILES: EvaluationProfile[] = [
        "director",
        "director-general",
        "jefe",
    ];
    const isAuthorized = $derived(ALLOWED_PROFILES.includes(profile));
    const isJefe = $derived(profile === "jefe");

    // ─── Jefe evaluatees state ──────────────────────────────────────────────

    let jefeEvaluatees = $state<
        Array<{
            id: string;
            firstName: string;
            lastName: string;
            employeeNumber?: string;
            profileId?: string;
        }>
    >([]);
    let jefeLoading = $state(false);
    let jefeError = $state<string | null>(null);

    onMount(() => {
        if (isAuthorized) {
            if (isJefe) {
                loadJefeEvaluatees();
            } else {
                load().then(() => {
                    // Auto-select the tree root (user's subtree or full tree)
                    const root = treeRoot;
                    if (root) handleNodeSelect(root);
                });
            }
        }
    });

    async function loadJefeEvaluatees(): Promise<void> {
        const user = getSession().user;
        if (!user?.employeeId) {
            jefeError = "No se pudo determinar el usuario";
            return;
        }

        jefeLoading = true;
        jefeError = null;

        try {
            const res = await client.GET("/employees/{empId}/evaluatees", {
                params: { path: { empId: user.employeeId } },
            });
            if (res.error) {
                throw new Error("Error al cargar evaluados");
            }
            const raw = res.data as {
                data?: Array<{
                    id: string;
                    firstName?: string;
                    lastName?: string;
                    employeeNumber?: string;
                    profileId?: string;
                }>;
            };
            jefeEvaluatees = (raw?.data ?? []).map((e) => ({
                id: e.id,
                firstName: e.firstName ?? "",
                lastName: e.lastName ?? "",
                employeeNumber: e.employeeNumber,
                profileId: e.profileId,
            }));
        } catch (e) {
            jefeError =
                e instanceof Error ? e.message : "Error al cargar evaluados";
        } finally {
            jefeLoading = false;
        }
    }

    // ─── Tree root (non-jefe) ──────────────────────────────────────────────

    const user = $derived(getSession().user);
    const userOrgNodeId = $derived(user?.orgNodeId ?? "");

    const treeRoot: OrgNode | null = $derived(
        isAuthorized && !isJefe
            ? userOrgNodeId
                ? (getSubtree(userOrgNodeId) ?? getRoot())
                : getRoot()
            : null,
    );

    // ─── Employee leaf type ───────────────────────────────────────────────

    interface EmployeeLeaf {
        id: string;
        firstName: string;
        lastName: string;
        jobTitle?: string;
        profileName?: string;
        profileDescription?: string;
    }

    // ─── Node selection ────────────────────────────────────────────────────

    let selectedNodeId = $state<string>("");
    let selectedNode = $state<OrgNode | null>(null);

    // Employees indexed by nodeId for tree leaf rendering
    let employeeLeaves = $state<Record<string, EmployeeLeaf[]>>({});

    // ─── Employee selection ────────────────────────────────────────────────

    let selectedEmployee = $state<EmployeeLeaf | null>(null);
    let selectedEmployeeScore = $state<number | null>(null);
    let employeeScoreLoading = $state(false);

    function handleNodeSelect(node: OrgNode) {
        selectedNodeId = node.id;
        selectedNode = node;
        // Open card for head employee if available
        if (node.headEmployee) {
            selectedEmployee = {
                id: node.headEmployee.id,
                firstName: node.headEmployee.firstName,
                lastName: node.headEmployee.lastName,
                jobTitle: node.headEmployee.jobTitle,
                profileName: node.headEmployee.profileName,
                profileDescription: node.headEmployee.profileDescription,
            };
            loadEmployeeScore(node.headEmployee.id);
        } else {
            selectedEmployee = null;
            selectedEmployeeScore = null;
        }
        loadNodeEmployees(node.id);
    }

    function handleEmployeeSelect(emp: EmployeeLeaf) {
        selectedEmployee = emp;
        selectedNodeId = "";
        selectedNode = null;
        loadEmployeeScore(emp.id);
    }

    async function loadNodeEmployees(nodeId: string): Promise<void> {
        try {
            const res = await client.GET("/employees", {
                params: { query: { nodeId, limit: 50 } },
            });
            if (res.error) {
                throw new Error("Error al cargar empleados");
            }
            const raw = res.data as {
                data?: Array<{
                    id: string;
                    firstName?: string;
                    lastName?: string;
                    jobTitle?: string;
                    profileName?: string;
                    profileDescription?: string;
                }>;
            };
            const employees = (raw?.data ?? []).map((e) => ({
                id: e.id,
                firstName: e.firstName ?? "",
                lastName: e.lastName ?? "",
                jobTitle: e.jobTitle,
                profileName: e.profileName,
                profileDescription: e.profileDescription,
            }));
            employeeLeaves = { ...employeeLeaves, [nodeId]: employees };
        } catch (e) {
            console.error("Error loading node employees:", e);
        }
    }

    async function loadEmployeeScore(empId: string): Promise<void> {
        employeeScoreLoading = true;
        selectedEmployeeScore = null;
        try {
            const res = await client.GET("/employees/{empId}/score", {
                params: { path: { empId } },
            });
            if (!res.error) {
                const data = res.data as { score?: number };
                selectedEmployeeScore = data?.score ?? null;
            }
        } catch (e) {
            console.error("Error loading employee score:", e);
        } finally {
            employeeScoreLoading = false;
        }
    }

    function handleJefeSelect(_id: string) {
        // ponytail: detail panel for selected evaluatee — add when needed
    }

    // Hide progress/competencies/goals for director-general or root node head
    const isHeadOfRoot = $derived(
        selectedNode && treeRoot && selectedNode.id === treeRoot.id,
    );
    const hideDetailGrid = $derived(
        selectedEmployee?.profileName === "director-general" || isHeadOfRoot,
    );
</script>

<svelte:head>
    <title>Jerarquía organizacional — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
        <div>
            <h1
                class="text-2xl font-bold text-base-content flex items-center gap-2"
            >
                <Network class="w-6 h-6" />
                Jerarquía organizacional
            </h1>
            <p class="text-sm text-base-content/50 mt-1">
                Explora la estructura de tu organización
            </p>
        </div>
        <a href="/evaluacion/9x9" class="btn btn-ghost btn-sm gap-1.5">
            <ArrowUpRight class="w-4 h-4" />
            Matriz 9×9
        </a>
    </div>

    {#if !isAuthorized}
        <EmptyState
            title="Sin acceso"
            message="Solo directores, director general y jefes pueden ver la jerarquía organizacional."
            actionLabel="Volver al inicio"
            actionHref="/"
        />
    {:else if isJefe}
        <!-- Jefe view: table of direct reports -->
        <div class="card bg-base-100 border border-base-300">
            <div class="card-body p-0">
                <h2
                    class="card-title text-xs font-semibold text-base-content/50 tracking-wide px-4 pt-4"
                >
                    Mis evaluados
                </h2>
                <JefeEvaluateesTable
                    evaluatees={jefeEvaluatees}
                    loading={jefeLoading}
                    error={jefeError}
                    onSelect={handleJefeSelect}
                />
            </div>
        </div>
    {:else if isLoading()}
        <div class="flex flex-col md:flex-row gap-6">
            <div class="md:flex-[2]">
                <div class="card bg-base-100 border border-base-300">
                    <div class="card-body p-0">
                        <h2
                            class="card-title text-xs font-semibold text-base-content/50 tracking-wide px-4 pt-4"
                        >
                            Organigrama
                        </h2>
                        <OrgHierarchyTree loading={true} />
                    </div>
                </div>
            </div>
            <div class="md:flex-[1]">
                <div class="card bg-base-100 border border-base-300">
                    <div class="card-body p-8 text-center">
                        <PageSkeleton variant="default" rows={5} />
                    </div>
                </div>
            </div>
        </div>
    {:else if getError()}
        <ErrorState message={getError() ?? undefined} onretry={reload} />
    {:else if !treeRoot}
        <EmptyState
            title="Sin datos"
            message="No se encontró la jerarquía para tu perfil."
        />
    {:else}
        <div class="flex flex-col md:flex-row gap-6">
            <!-- Tree -->
            <div class="md:flex-[2]">
                <div class="card bg-base-100 border border-base-300">
                    <div class="card-body p-0 overflow-x-auto">
                        <h2
                            class="card-title text-xs font-semibold text-base-content/50 tracking-wide"
                        >
                            Organigrama
                        </h2>
                        <OrgHierarchyTree
                            node={treeRoot}
                            onNodeSelect={handleNodeSelect}
                            {selectedNodeId}
                            {employeeLeaves}
                            onEmployeeSelect={handleEmployeeSelect}
                            selectedEmployeeId={selectedEmployee?.id ?? ""}
                        />
                    </div>
                </div>
            </div>

            <!-- Node/Employee detail panel -->
            <div class="md:flex-[1] md:sticky md:top-4 md:self-start">
                {#if selectedEmployee}
                    <div class="card bg-base-100 border border-base-300">
                        <div class="card-body p-0">
                            <h2 class="card-title text-lg px-4 pt-4 mb-0">
                                {selectedEmployee.firstName}
                                {selectedEmployee.lastName}
                            </h2>

                            <div class="p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
                                {#if selectedEmployee.jobTitle}
                                    <div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
                                        <Briefcase class="w-5 h-5 text-base-content/40 shrink-0" />
                                        <div>
                                            <p class="text-xs text-base-content/40">Puesto</p>
                                            <p class="text-xs font-medium capitalize">{selectedEmployee.jobTitle}</p>
                                        </div>
                                    </div>
                                {/if}

                                {#if selectedEmployee.profileName}
                                    <div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
                                        <Users class="w-5 h-5 text-base-content/40 shrink-0" />
                                        <div>
                                            <p class="text-xs text-base-content/40">Perfil</p>
                                            <p class="text-xs font-medium">{titleCase(selectedEmployee.profileName)}</p>
                                        </div>
                                    </div>
                                {/if}

                                {#if !hideDetailGrid}
                                <div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
                                    <Target class="w-5 h-5 text-base-content/40 shrink-0" />
                                    <div>
                                        <p class="text-xs text-base-content/40">Progreso global</p>
                                        {#if employeeScoreLoading}
                                            <p class="text-sm text-base-content/40">Cargando…</p>
                                        {:else if selectedEmployeeScore !== null}
                                            <p class="text-sm font-medium">{Math.round(selectedEmployeeScore)}%</p>
                                        {:else}
                                            <p class="text-sm text-base-content/40">Sin datos</p>
                                        {/if}
                                    </div>
                                </div>

                                <div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
                                    <ArrowUpRight class="w-5 h-5 text-base-content/40 shrink-0" />
                                    <div>
                                        <p class="text-xs text-base-content/40">Competencias</p>
                                        <a
                                            href="/evaluacion/9x9/competencias/{selectedEmployee.id}"
                                            class="link link-primary text-sm font-medium"
                                        >
                                            Ver red
                                        </a>
                                    </div>
                                </div>

                                <div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
                                    <ArrowUpRight class="w-5 h-5 text-base-content/40 shrink-0" />
                                    <div>
                                        <p class="text-xs text-base-content/40">Metas</p>
                                        <a
                                            href="/mis-evaluados"
                                            class="link link-primary text-sm font-medium"
                                        >
                                            Ver metas
                                        </a>
                                    </div>
                                </div>
                                {/if}
                            </div>
                        </div>
                    </div>
                {:else if selectedNode}
                    <div class="card bg-base-100 border border-base-300">
                        <div class="card-body p-8 text-center">
                            <Users
                                class="w-10 h-10 text-base-content/20 mx-auto mb-3"
                            />
                            <p class="text-sm text-base-content/40">
                                {selectedNode.name} — sin jefe asignado
                            </p>
                        </div>
                    </div>
                {:else}
                    <div class="card bg-base-100 border border-base-300">
                        <div class="card-body p-8 text-center">
                            <Users
                                class="w-10 h-10 text-base-content/20 mx-auto mb-3"
                            />
                            <p class="text-sm text-base-content/40">
                                Selecciona un nodo o colaborador del
                                organigrama para ver sus detalles.
                            </p>
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>
