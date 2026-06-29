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
    import { ArrowUpRight, Network, Users } from "@lucide/svelte";

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

    // ─── Node selection ────────────────────────────────────────────────────

    let selectedNodeId = $state<string>("");
    let selectedNode = $state<OrgNode | null>(null);
    let nodeEmployees = $state<
        Array<{
            id: string;
            firstName: string;
            lastName: string;
            jobTitle?: string;
            profileDescription?: string;
        }>
    >([]);
    let employeesLoading = $state(false);

    function handleNodeSelect(node: OrgNode) {
        selectedNodeId = node.id;
        selectedNode = node;
        loadNodeEmployees(node.id);
    }

    async function loadNodeEmployees(nodeId: string): Promise<void> {
        employeesLoading = true;
        nodeEmployees = [];

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
                    profileDescription?: string;
                }>;
            };
            nodeEmployees = (raw?.data ?? []).map((e) => ({
                id: e.id,
                firstName: e.firstName ?? "",
                lastName: e.lastName ?? "",
                jobTitle: e.jobTitle,
                profileDescription: e.profileDescription,
            }));
        } catch (e) {
            console.error("Error loading node employees:", e);
            nodeEmployees = [];
        } finally {
            employeesLoading = false;
        }
    }

    function handleJefeSelect(_id: string) {
        // ponytail: detail panel for selected evaluatee — add when needed
    }
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
        <div class="flex flex-col lg:flex-row gap-8">
            <div class="lg:w-1/2 xl:w-2/5">
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
            <div class="lg:w-1/2 xl:w-3/5">
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
        <div class="flex flex-col lg:flex-row gap-8">
            <!-- Tree -->
            <div class="lg:w-1/2 xl:w-2/5">
                <div class="card bg-base-100 border border-base-300">
                    <div class="card-body p-0">
                        <h2
                            class="card-title text-xs font-semibold text-base-content/50 tracking-wide"
                        >
                            Organigrama
                        </h2>
                        <OrgHierarchyTree
                            node={treeRoot}
                            onNodeSelect={handleNodeSelect}
                            {selectedNodeId}
                        />
                    </div>
                </div>
            </div>

            <!-- Node summary panel -->
            <div class="lg:w-1/2 xl:w-3/5">
                {#if selectedNode}
                    <div class="card bg-base-100 border border-base-300">
                        <div class="card-body p-0">
                            <div
                                class="flex items-center justify-between px-4 pt-4 pb-2 border-b border-base-200"
                            >
                                <h2 class="card-title text-lg">
                                    {selectedNode.headEmployee
                                        ? `${selectedNode.headEmployee.firstName} ${selectedNode.headEmployee.lastName}`
                                        : selectedNode.name}
                                </h2>
                                {#if selectedNode.headEmployee?.jobTitle}
                                    <span
                                        class="badge badge-ghost badge-sm capitalize"
                                    >
                                        {selectedNode.headEmployee.jobTitle}
                                    </span>
                                {/if}
                            </div>

                            <!-- Employees section -->
                            <div class="px-4 py-3">
                                {#if employeesLoading}
                                    <div
                                        class="py-4 text-center text-sm text-base-content/40"
                                    >
                                        Cargando empleados…
                                    </div>
                                {:else if nodeEmployees.length === 0}
                                    <div
                                        class="py-4 text-center text-sm text-base-content/40"
                                    >
                                        Sin empleados en esta área
                                    </div>
                                {:else}
                                    <div class="overflow-x-auto">
                                        <table class="table table-sm">
                                            <thead>
                                                <tr
                                                    class="text-xs text-base-content/50"
                                                >
                                                    <th>Nombre</th>
                                                    <th>Puesto</th>
                                                    <th>Perfil</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {#each nodeEmployees as emp (emp.id)}
                                                    <tr class="hover">
                                                        <td class="font-medium"
                                                            >{emp.firstName}
                                                            {emp.lastName}</td
                                                        >
                                                        <td
                                                            class="text-xs text-base-content/50"
                                                            >{emp.jobTitle ??
                                                                "—"}</td
                                                        >
                                                        <td>
                                                            <span
                                                                class="text-xs text-base-content/50"
                                                            >
                                                                {emp.profileDescription ??
                                                                    "—"}
                                                            </span>
                                                        </td>
                                                    </tr>
                                                {/each}
                                            </tbody>
                                        </table>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    </div>
                {:else}
                    <div class="card bg-base-100 border border-base-300">
                        <div class="card-body p-8 text-center">
                            <Users
                                class="w-10 h-10 text-base-content/20 mx-auto mb-3"
                            />
                            <p class="text-sm text-base-content/40">
                                Selecciona un nodo del organigrama para ver sus
                                empleados.
                            </p>
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>
