<script lang="ts">
	import { page } from '$app/stores';
	import Separator from '$components/Separator.svelte';
	import Theme from '$components/Theme.svelte';
	import { Icons } from '$components/icons';
	import { header } from '$lib/config/navs';
	import { site } from '$lib/config/site';
	import { cn, isBrowser } from '$lib/utils/general';
	import { createCollapsible } from '@melt-ui/svelte';
	import { useClickOutside } from '@melt-ui/svelte/internal/actions';
	import { slide } from 'svelte/transition';
	import Burger from './Burger.svelte';

	// ----------------------
	// Variables
	// ----------------------
	const {
		elements: { root, content, trigger },
		states: { open }
	} = createCollapsible();

	// ----------------------
	// Reactive
	// ----------------------
	$: {
		if (isBrowser) {
			if ($open) {
				document.body.style.overflow = 'hidden';
				document.body.style.height = '100%';
			} else {
				document.body.style.overflow = 'auto';
				document.body.style.height = 'auto';
			}
		}
	}
</script>

<!-- Shown on sm- -->
<nav
	class="md:hidden"
	{...$root}
	use:useClickOutside={{
		handler: () => {
			open.set(false);
		}
	}}
>
	<Burger {trigger} {open} />

	{#if $open}
		<div
			{...$content}
			transition:slide={{ duration: 200 }}
			class="bg-background fixed inset-0 top-16 z-10 h-screen w-screen"
		>
			<div class="flex flex-col place-content-center items-center gap-2 pt-10">
				{#each header as navItem}
					<a
						href={navItem.href}
						class={cn(
							'hover:text-primary w-64 border-b px-1 py-2 text-base font-semibold duration-200',
							$page.url.pathname.startsWith(navItem.href) && 'text-primary'
						)}
						on:click={() => {
							open.set(false);
						}}
					>
						{navItem.title}
					</a>
				{/each}

				<div class="flex w-64 select-none pt-5">
					<div class="bg-accent-1 flex w-full items-center justify-between rounded-md px-2 py-1.5">
						<span class="pl-1 text-sm">Appearance</span>
						<Theme />
					</div>
				</div>

				<div class="flex w-64 select-none items-center justify-center pt-2.5">
					<a
						href={site.links.github}
						target="_blank"
						rel="noreferrer"
						class="flex items-center px-1"
					>
						<div
							class="hover:bg-accent-1 group rounded-md p-1.5 text-sm font-semibold duration-200"
						>
							<Icons.gitHub
								class="fill-foreground-muted group-hover:fill-foreground h-7 w-7 stroke-none duration-200"
							/>
							<span class="sr-only">GitHub</span>
						</div>
					</a>
				</div>
			</div>
		</div>
	{/if}
</nav>

<!-- Showing on md+ -->
<nav class="hidden items-center gap-1.5 md:inline-flex">
	{#each header as navItem}
		<a
			href={navItem.href}
			class={cn(
				'hover:text-primary px-3 py-2 text-sm font-semibold duration-200',
				$page.url.pathname.startsWith(navItem.href) && 'text-primary'
			)}
		>
			{navItem.title}
		</a>
	{/each}

	<Separator orientation="vertical" class="" />

	<div class="flex flex-row">
		<a href={site.links.github} target="_blank" rel="noreferrer" class="flex items-center px-1">
			<div class="hover:bg-accent-1 group rounded-md p-1.5 text-sm font-semibold duration-200">
				<Icons.gitHub
					class="fill-foreground-muted group-hover:fill-foreground h-5 w-5 stroke-none duration-200"
				/>
				<span class="sr-only">GitHub</span>
			</div>
		</a>

		<Theme />
	</div>
</nav>
