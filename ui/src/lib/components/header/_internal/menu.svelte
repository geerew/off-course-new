<script lang="ts">
	import { page } from '$app/stores';
	import { Github } from '$components/icons';
	import Button from '$components/ui/button/button.svelte';
	import { Separator } from '$components/ui/separator';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { header } from '$lib/config/navs';
	import { site } from '$lib/config/site';
	import { FlyAndScale, cn } from '$lib/utils';
	import Burger from './burger.svelte';

	// ----------------------
	// Variables
	// ----------------------
	let isOpen = false;

	$: console.log('isOpen', isOpen);
</script>

<!-- Shown on sm- -->
<Sheet.Root bind:open={isOpen} closeOnOutsideClick={false}>
	<Sheet.Trigger asChild>
		<Burger bind:open={isOpen} />
	</Sheet.Trigger>

	<Sheet.Content
		overlay={false}
		side="top"
		class="top-[calc(var(--header-height)+1px)] h-screen [&>button[data-dialog-close]]:hidden"
		transition={FlyAndScale}
		transitionConfig={{ y: -30, duration: 300 }}
	>
		<div class="flex w-full flex-col items-center pt-10">
			<div class="flex w-80 flex-col">
				{#each header as navItem}
					<a
						href={navItem.href}
						class={cn(
							'hover:text-primary border-alt-1/60 w-full border-b px-1 py-4 text-base font-semibold duration-200',
							$page.url.pathname.startsWith(navItem.href) && 'text-primary'
						)}
						on:click={() => {
							isOpen = false;
						}}
					>
						{navItem.title}
					</a>
				{/each}

				<!-- 
				<div class="flex w-64 select-none pt-5">
					<div class="bg-accent-1 flex w-full items-center justify-between rounded-md px-2 py-1.5">
						<span class="pl-1 text-sm">Appearance</span>
						<Theme />
					</div>
				</div> 
				-->

				<div class="flex select-none items-center justify-center pt-5">
					<Button
						variant="ghost"
						href={site.links.github}
						class="group"
						rel="noreferrer"
						target="_blank"
					>
						<Github
							class="fill-muted-foreground group-hover:fill-foreground size-6 stroke-none duration-200"
						/>
						<span class="sr-only">GitHub</span>
					</Button>
				</div>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>

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
		<Button variant="ghost" href={site.links.github} class="group" rel="noreferrer" target="_blank">
			<Github
				class="fill-muted-foreground group-hover:fill-foreground size-5 stroke-none duration-200"
			/>
			<span class="sr-only">GitHub</span>
		</Button>

		<!-- <Theme /> -->
	</div>
</nav>
