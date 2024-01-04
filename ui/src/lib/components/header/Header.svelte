<script lang="ts">
	import { page } from '$app/stores';
	import { site } from '$lib/config/site';
	import { cn } from '$lib/utils/general';
	import { Icons } from '../icons';
	import HeaderMenu from './HeaderMenu.svelte';
	import SettingsMenu from './SettingsMenu.svelte';
</script>

<header
	class={cn(
		'w-full border-b',
		$page.url.pathname.startsWith('/settings') ? '' : 'h-[var(--header-height)]'
	)}
>
	<div
		class={cn(
			'container flex items-center justify-between gap-8',
			$page.url.pathname.startsWith('/settings') ? 'h-16' : 'h-full'
		)}
	>
		<div class="mr-3">
			<a href="/" class="mr-6 flex items-center gap-2">
				<Icons.logo class="fill-primary h-8 w-8 rotate-90 stroke-none" />
				<span class="inline-block font-semibold">{site.name}</span>
			</a>
		</div>

		<div class="flex items-center justify-between">
			<HeaderMenu />
		</div>
	</div>

	<!-- Settings -->
	{#if $page.url.pathname.startsWith('/settings')}
		<div class="container flex items-center px-4 md:px-8">
			<SettingsMenu />
		</div>
	{/if}
</header>
