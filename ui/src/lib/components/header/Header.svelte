<script lang="ts">
	import { page } from '$app/stores';
	import { site } from '$lib/config/site';
	import { Hexagon } from 'lucide-svelte';
	import Menu from './_internal/menu.svelte';
	import SettingsMenu from './_internal/settings-menu.svelte';
</script>

<header class="bg-background sticky top-0 z-50 w-full border-b">
	<div class="flex h-[var(--header-height)] items-center justify-between gap-8 px-4 md:px-8">
		<div class="mr-3">
			<a href="/" class="mr-6 flex items-center gap-2">
				<Hexagon class="fill-primary h-8 w-8 rotate-90 stroke-none" />
				<span class="inline-block font-semibold">{site.name}</span>
			</a>
		</div>

		<div class="flex items-center justify-between">
			<Menu />
		</div>
	</div>

	<!-- Settings -->
	{#if $page.url.pathname.startsWith('/settings')}
		<div class="container flex items-center px-4 md:px-8">
			<SettingsMenu />
		</div>
	{/if}
</header>
