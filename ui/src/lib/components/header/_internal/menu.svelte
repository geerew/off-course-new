<script lang="ts">
	import { page } from '$app/stores';
	import Theme from '$components/generic/theme.svelte';
	import { Github } from '$components/icons';
	import Button from '$components/ui/button/button.svelte';
	import { Separator } from '$components/ui/separator';
	import { header } from '$lib/config/navs';
	import { site } from '$lib/config/site';
	import { cn, isBrowser } from '$lib/utils';
	import { createCollapsible } from '@melt-ui/svelte';
	import { slide } from 'svelte/transition';
	import Burger from './burger.svelte';

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
<nav class="md:hidden" {...$root}>
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
					<a href={site.links.github} target="_blank" rel="noreferrer" class="flex items-center">
						<div
							class="hover:bg-accent-1 group rounded-md p-1.5 text-sm font-semibold duration-200"
						>
							<Github
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
<nav class="hidden items-center gap-2.5 md:inline-flex">
	{#each header as navItem}
		<Button
			variant="link"
			href={navItem.href}
			class={cn(
				'text-muted-foreground',
				$page.url.pathname.startsWith(navItem.href) && 'text-foreground'
			)}
		>
			{navItem.title}
		</Button>
	{/each}

	<Separator orientation="vertical" class="h-8" />

	<div class="flex flex-row gap-2.5">
		<Button variant="ghost" href={site.links.github} size="icon" class="fill-foreground">
			<Github class="h-[1.2rem] w-[1.2rem]" />
			<span class="sr-only">GitHub</span>
		</Button>

		<Theme />
	</div>
</nav>
