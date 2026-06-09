<script lang="ts">
    import { getProfile } from "$lib/stores/devContext.svelte";
    import {
        PROFILE_LABELS,
        type EvaluationProfile,
    } from "$lib/types/evaluation";
    import { PROFILE_USERS } from "$lib/dev/profileUsers";
    import {
        User,
        LogOut,
        Clock,
        ClipboardCheck,
        Target,
        MessageSquare,
        LogIn,
        Eye,
        Star,
        Award,
        FileText,
        Download,
        CheckCircle,
    } from "@lucide/svelte";
    import activityLogs from "$lib/fixtures/activity/activity-logs.json";

    const profile = $derived(getProfile());
    const user = $derived(PROFILE_USERS[profile]);
    const profileLabel = $derived(PROFILE_LABELS[profile]);
    const userInitial = $derived(user.name.charAt(0).toUpperCase());

    const areaMap: Record<EvaluationProfile, string> = {
        colaborador: "Operaciones · Sucursal Centro",
        jefe: "Servicio al Cliente",
        vendedor: "Ventas · Tienda Polanco",
        "gerente-tienda": "Tienda Polanco",
        divisional: "División Comercial",
        regional: "Región Centro",
        director: "Dirección General",
        "director-general": "Dirección General Corporativa",
        rh: "Recursos Humanos",
    };
    const myArea = $derived(areaMap[profile]);

    const filteredLogs = $derived(
        activityLogs
            .filter((log) => log.profileId === profile)
            .sort(
                (a, b) =>
                    new Date(b.timestamp).getTime() -
                    new Date(a.timestamp).getTime(),
            ),
    );

    const ACTION_ICONS: Record<string, typeof Clock> = {
        evaluation_started: ClipboardCheck,
        evaluation_completed: CheckCircle,
        goal_viewed: Eye,
        goal_approved: CheckCircle,
        goal_progress: Target,
        comment_added: MessageSquare,
        login: LogIn,
        profile_viewed: User,
        competency_viewed: Star,
        hierarchy_viewed: Eye,
        evaluation_reviewed: Eye,
        cycle_configured: ClipboardCheck,
        pillar_edited: Award,
        scale_modified: FileText,
        report_exported: Download,
    };

    function getActionIcon(action: string) {
        return ACTION_ICONS[action] ?? Clock;
    }

	function formatTimeLabel(timestamp: string): string {
        const now = new Date();
        const date = new Date(timestamp);
        const diffMs = now.getTime() - date.getTime();
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffDays < 1) {
            const diffMins = Math.floor(diffMs / 60000);
            const diffHours = Math.floor(diffMins / 60);
            if (diffMins < 1) return "Ahora mismo";
            if (diffMins < 60) return `Hace ${diffMins} min`;
            if (diffHours < 24) return `Hace ${diffHours}h`;
            return "Hoy";
        }

        return date.toLocaleDateString("es-AR", {
            day: "numeric",
            month: "long",
            year: "numeric",
        });
    }

    function handleLogout() {
        alert("Cerrar sesión (mock): en producción redirigiría a /login");
    }
</script>

<svelte:head>
    <title>Mi perfil — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
    <!-- User info -->
    <div class="flex items-start justify-between">
        <div class="flex flex-col items-start gap-3">
            <div
                class="w-16 h-16 rounded-full bg-primary flex items-center justify-center shrink-0"
            >
                <span class="text-primary-content font-bold text-2xl"
                    >{userInitial}</span
                >
            </div>
            <div class="flex flex-col">
                <div class="flex items-center gap-2">
                    <h2 class="text-lg font-semibold">{user.name}</h2>
                </div>
                <div class="flex flex-wrap items-center gap-2 mt-1">
                    <span
                        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium"
                    >
                        {profileLabel}
                    </span>
                    <span
                        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-base-200 text-base-content/60 text-xs font-medium"
                    >
                        {myArea}
                    </span>
                </div>
                <p class="text-sm text-base-content/50 mt-1.5">{user.email}</p>
            </div>
        </div>
        <button
            class="btn btn-outline btn-error btn-sm gap-2 mt-2"
            onclick={handleLogout}
        >
            <LogOut class="w-4 h-4" />
            Cerrar sesión
        </button>
    </div>

    <!-- Activity log -->
    <div>
        <h2
            class="text-xs font-semibold text-base-content/50 tracking-wide mb-3"
        >
            Actividad reciente
        </h2>

        {#if filteredLogs.length === 0}
            <div
                class="flex flex-col items-center text-center py-8 text-base-content/40"
            >
                <Clock class="w-10 h-10 text-base-content/20" />
                <p class="text-sm mt-2">No hay actividad registrada</p>
            </div>
        {:else}
            <section class="flex items-start flex-col">
                <ul class="timeline timeline-vertical timeline-left">
                    {#each filteredLogs as log (log.id)}
                        {@const Icon = getActionIcon(log.action)}
                        <li>
                            <hr class="bg-neutral" />
                            <div
                                class="timeline-middle bg-neutral text-neutral-content rounded-full p-1"
                            >
                                <Icon class="w-4 h-4" />
                            </div>
                            <div class="timeline-end timeline-box">
                                <p class="text-sm">{log.description}</p>
                                <p class="text-xs text-base-content/50 mt-1">
                                    {formatTimeLabel(log.timestamp)}
                                </p>
                            </div>
                            <hr class="bg-neutral" />
                        </li>
                    {/each}
                </ul>
            </section>
        {/if}
    </div>
</div>
