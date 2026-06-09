<script lang="ts">
    import { page } from "$app/stores";
    import logoBlack from "$lib/assets/logo_black.png";
    import logoWhite from "$lib/assets/logo_white.png";
    import { getVisibleMenuItems } from "$lib/nav/menuConfig";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { PROFILE_LABELS } from "$lib/types/evaluation";
    import { PROFILE_USERS } from "$lib/dev/profileUsers";
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
    } from "@lucide/svelte";
    import { version } from "../../../package.json";

    interface Props {
        onclose?: () => void;
    }

    let { onclose }: Props = $props();

    const profile = $derived(getProfile());
    const visibleItems = $derived(getVisibleMenuItems(profile, "inicio-anio"));
    const profileLabel = $derived(PROFILE_LABELS[profile]);
    const year = new Date().getFullYear();

    const user = $derived(PROFILE_USERS[profile]);
    const userName = $derived(user.name);
    const userInitial = $derived(userName.charAt(0).toUpperCase());

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
    };

    function getIcon(name: string) {
        return iconMap[name] ?? Home;
    }

    function handleNav() {
        onclose?.();
    }
</script>

<aside class="bg-base-100 h-full min-h-screen w-64 flex flex-col">
    <!-- Logo / Brand -->
    <div class="px-5 pt-5 pb-4">
        <div class="flex items-center gap-2.5">
            <img src={logoBlack} alt="SED" class="logo-theme-light h-6 w-auto max-w-full object-contain" />
            <img src={logoWhite} alt="SED" class="logo-theme-dark h-6 w-auto max-w-full object-contain" />
        </div>
    </div>

    <!-- Profile badge -->
    <div class="px-3 mb-4">
        <a
            class="px-3 py-2 rounded-lg flex items-center gap-3 hover:bg-base-200 transition-colors"
            href="/perfil"
            onclick={handleNav}
        >
            <div
                class="w-8 h-8 rounded-full bg-primary flex items-center justify-center shrink-0"
            >
                <span class="text-primary-content font-bold text-sm"
                    >{userInitial}</span
                >
            </div>
            <div class="min-w-0">
                <p class="text-xs text-base-content/60 truncate">{profileLabel}</p>
                <p class="text-xs font-medium truncate">{userName}</p>
            </div>
        </a>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 px-3" aria-label="Navegación principal">
        <ul class="flex flex-col gap-0.5">
            {#each visibleItems as item (item.href)}
                {@const isActive = $page.url.pathname === item.href}
                {@const Icon = getIcon(item.icon)}
                <li>
                    <a
                        href={item.href}
                        class="flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-lg transition-colors
							{isActive
                            ? 'bg-primary/10 text-primary'
                            : 'text-base-content/60 hover:bg-base-200 hover:text-base-content'}"
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

    <!-- Footer -->
    <div class="px-5 py-4">
        <p class="text-[11px] text-base-content/30 text-center">
            Portal SED {year} v{version}
        </p>
    </div>
</aside>
