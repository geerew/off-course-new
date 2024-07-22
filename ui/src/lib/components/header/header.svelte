<script lang="ts">
	import { page } from '$app/stores';
	import { Icons } from '$components/icons';
	import { site } from '$lib/config/site';
	import Menu from './_internal/menu.svelte';
	import SettingsMenu from './_internal/settings-menu.svelte';
</script>

<header class="sticky top-0 z-20 w-full border-b bg-background">
	<div
		class="container flex h-[var(--header-height)] items-center justify-between gap-8 px-4 md:px-8"
	>
		<div class="mr-3">
			<a href="/" class="mr-6 flex items-center gap-2">
				<Icons.Logo
					class="size-8 [&>:nth-child(1)]:stroke-primary [&>:nth-child(2)]:stroke-secondary"
				/>
				<span class="hidden font-semibold md:inline-block">{site.name}</span>
			</a>
		</div>

		<div class="flex items-center justify-between">
			<Menu />
		</div>
	</div>

	<!-- /settings/... -->
	{#if $page.url.pathname.startsWith('/settings')}
		<div class="container flex items-center px-4 md:px-8">
			<SettingsMenu />
		</div>
	{/if}
</header>
