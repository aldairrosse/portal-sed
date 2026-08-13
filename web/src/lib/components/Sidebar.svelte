<script lang="ts">
    import { page } from "$app/stores";
    import logoWhite from "$lib/assets/logo_white.png";
    import { getVisibleMenuItems } from "$lib/nav/menuConfig";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { getSession, logout } from "$lib/api/session.svelte";
    import { LogOut } from "@lucide/svelte";
    import {
        Home,
        Target,
        TrendingUp,
        ClipboardCheck,
        Users,
        Grid3x3,
        Award,
        FileText,
        ClipboardList,
        Star,
        Network,
        Calendar,
    } from "@lucide/svelte";
    import { version } from "../../../package.json";

    interface Props {
        onclose?: () => void;
    }

    let { onclose }: Props = $props();

    const session = $derived(getSession());
    const profile = $derived(getProfile());
    const visibleItems = $derived(getVisibleMenuItems(profile, "inicio-anio"));
    const year = new Date().getFullYear();

    const user = $derived(session.user);
    const userName = $derived(user?.name ?? '');

    const iconMap: Record<string, typeof Home> = {
        Home,
        Target,
        TrendingUp,
        ClipboardCheck,
        Users,
        Grid3x3,
        Award,
        FileText,
        ClipboardList,
        Star,
        Network,
        Calendar,
    };

    function getIcon(name: string) {
        return iconMap[name] ?? Home;
    }

    function handleNav() {
        onclose?.();
    }
</script>

<aside class="text-(--color-sidebar-content) bg-(--color-sidebar) h-full min-h-screen w-64 flex flex-col overflow-y-auto">
    <!-- Logo / Brand (sidebar bg is always dark → white logo) -->
    <div class="mt-8 px-8 pt-5 pb-4">
        <div class="flex items-center gap-2.5">
            <img src={logoWhite} alt="SED" class="h-6 w-auto max-w-full object-contain" />
        </div>
    </div>

    <!-- Profile badge -->
    <div class="mb-4">
        <a
            class="py-2 px-8 rounded-0 block hover:bg-white/10 transition-colors"
            href="/perfil"
            onclick={handleNav}
        >
            <div class="min-w-0">
                <p class="text-xs font-medium truncate">{userName}</p>
                <p class="text-xs text-(--color-sidebar-content)/70 truncate">{user?.jobTitle}</p>
            </div>
        </a>
    </div>

    <!-- Navigation -->
    <nav class="flex-1" aria-label="Navegación principal">
        <ul class="flex flex-col gap-0.5">
            {#each visibleItems as item (item.href)}
                {@const isActive = $page.url.pathname === item.href}
                {@const Icon = getIcon(item.icon)}
                <li class="w-full">
                    <a
                        href={item.href}
                        class="flex w-full items-center gap-3 px-8 py-2.5 text-sm font-medium transition-colors
							{isActive
                            ? 'bg-primary/50 text-white'
                            : 'text-(--color-sidebar-content)/60 hover:bg-white/10 hover:text-(--color-sidebar-content)'}"
                        onclick={handleNav}
                        aria-current={isActive ? "page" : undefined}
                    >
                        <Icon
                            class="w-[18px] h-[18px] flex-shrink-0"
                            strokeWidth={isActive ? 2.2 : 1.8}
                        />
                        {item.label}
                    </a>
                </li>
            {/each}
        </ul>
    </nav>

    <!-- Logout -->
    <div class="mb-1">
        <button
            class="flex w-full items-center gap-3 px-8 py-2.5 text-sm font-medium transition-colors text-(--color-sidebar-content)/60 hover:bg-white/10 hover:text-error disabled:opacity-60 disabled:hover:bg-transparent disabled:hover:text-(--color-sidebar-content)/60 cursor-pointer"
            onclick={logout}
            disabled={session.loggingOut}
            aria-busy={session.loggingOut}
        >
            {#if session.loggingOut}
                <span class="loading loading-spinner loading-xs flex-shrink-0"></span>
                <span>Cerrando sesión…</span>
            {:else}
                <LogOut class="w-[18px] h-[18px] flex-shrink-0" />
                <span>Cerrar sesión</span>
            {/if}
        </button>
    </div>

    <!-- Footer -->
    <div class="px-8 py-4">
        <p class="text-[11px] text-(--color-sidebar-content)/30 text-center">
            Portal SED {year} v{version}
        </p>
    </div>
</aside>
