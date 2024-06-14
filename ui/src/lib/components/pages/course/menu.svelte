<script lang="ts">
	import Button from '$components/ui/button/button.svelte';
	import * as Sheet from '$components/ui/sheet';
	import * as Tooltip from '$components/ui/tooltip';
	import type { Asset, CourseChapters } from '$lib/types/models';
	import { UpdateQueryParam, cn } from '$lib/utils';
	import { CircleCheck, Info, Menu, X } from 'lucide-svelte';
	import { tick } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------

	// Course title
	export let title: string;

	// Course ID
	export let id: string;

	// Course chapters
	export let chapters: CourseChapters;

	// Currently selected asset. This should be used with `bind:`
	export let selectedAsset: Asset | null;

	// ----------------------
	// Variables
	// ----------------------

	let open = false;

	// ----------------------
	// Functions
	// ----------------------
	async function scroll(isMobile: boolean) {
		await tick();
		const selectedButton = document.querySelector(
			isMobile ? '[data-mobile-selected=true]' : '[data-selected=true]'
		);

		if (!selectedButton) return;
		selectedButton.scrollIntoView({ behavior: 'smooth', block: 'center' });
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When the mobile menu is opened, scroll to the currently selected asset
	$: {
		if (open) {
			scroll(true);
		}
	}

	$: {
		if (selectedAsset) {
			scroll(false);
		}
	}
</script>

<!-- xs, sm, md -->
<div class="border-b px-4 py-2 md:px-8 lg:hidden">
	<Sheet.Root openFocus="[data-mobile-selected=true]" bind:open>
		<Sheet.Trigger asChild let:builder>
			<Button
				builders={[builder]}
				variant="ghost"
				class="text-muted-foreground hover:text-foreground gap-2 px-0 hover:bg-inherit lg:hidden"
			>
				<Menu class="size-5" />
				<span>Menu</span>
			</Button>
		</Sheet.Trigger>

		<Sheet.Content
			side="left"
			class="overflow-none w-[calc(22rem+36px)] border-none bg-transparent p-0 shadow-none sm:max-w-[calc(22rem+36px)] [&>button[data-dialog-close]]:hidden"
		>
			<!--  X button -->
			<div
				class="bg-background absolute right-px top-1 z-10 flex place-content-center items-center rounded-r-md border-y border-r"
			>
				<Button
					variant="ghost"
					class="text-muted-foreground hover:text-foreground hover:bg-background h-auto p-2"
					on:click={() => (open = false)}
				>
					<X class="size-5" />
				</Button>
			</div>

			<!-- Overflow area that when clicked will close sheet -->
			<div class="absolute right-0 top-0 z-[1] flex h-full place-content-center bg-transparent">
				<Button
					variant="ghost"
					class="h-auto cursor-auto hover:bg-transparent"
					on:click={() => (open = false)}
				></Button>
			</div>

			<!-- Content -->
			<div class="bg-background relative h-full w-[22rem] border-r pr-0">
				<nav
					class="relative left-0 top-0 max-h-screen min-h-screen overflow-y-auto overflow-x-hidden"
					tabindex="-1"
				>
					<!-- Course title -->
					<div
						class="bg-background sticky top-0 z-[1] flex flex-row items-start gap-3 border-b py-6 pl-6 pr-3"
					>
						<span class="grow text-sm">{title}</span>

						<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
							<Tooltip.Trigger asChild let:builder>
								<Button
									builders={[builder]}
									variant="ghost"
									href="/settings/courses/details?id={id}"
									class="text-muted-foreground hover:text-foreground mt-1 h-auto px-0 py-0 hover:bg-transparent"
								>
									<Info class="size-4 shrink-0" />
								</Button>
							</Tooltip.Trigger>

							<Tooltip.Content
								class="bg-foreground text-background select-none rounded-sm border-none px-1.5 py-1 text-xs"
								transitionConfig={{ y: 8, duration: 100 }}
								side="bottom"
							>
								Details
								<Tooltip.Arrow class="bg-background" />
							</Tooltip.Content>
						</Tooltip.Root>
					</div>

					<ul class="ml-auto h-full columns-1 py-6 pl-6">
						{#each Object.keys(chapters) as chapter}
							<li class="pb-8 leading-5">
								<!-- Chapter heading -->
								<span
									class="after:bg-alt-1 relative flex w-full pr-2 text-base font-semibold tracking-wide after:absolute after:-bottom-1 after:left-0 after:h-px after:w-full"
								>
									{chapter}
								</span>

								<!-- Assets -->
								<ul class="pr-3 pt-3">
									{#each chapters[chapter] as asset}
										<li class="pl-1.5">
											<!-- Asset -->
											<Button
												variant="ghost"
												class={cn(
													'h-auto w-full justify-start whitespace-normal px-0 py-0 text-start hover:bg-transparent hover:underline',
													asset.id === selectedAsset?.id
														? 'decoration-foreground'
														: 'decoration-muted-foreground'
												)}
												data-mobile-selected={asset.id === selectedAsset?.id}
												on:click={() => {
													UpdateQueryParam('a', asset.id, false);
													open = false;
												}}
											>
												<div class={cn('flex w-full flex-row justify-between py-1.5')}>
													<!-- Asset title -->
													<div
														class={cn(
															'text-muted-foreground grow pr-2.5',
															asset.id === selectedAsset?.id && 'text-foreground'
														)}
													>
														<span>{asset.prefix}.</span>
														{asset.title}
													</div>

													<!-- Asset completed -->
													<CircleCheck
														class={cn(
															'text-muted-foreground mt-0.5 size-4 shrink-0',
															asset.completed &&
																'fill-success text-success [&>:nth-child(2)]:text-white'
														)}
													/>
												</div>
											</Button>
										</li>
									{/each}
								</ul>
							</li>
						{/each}
					</ul>
				</nav>
			</div>
		</Sheet.Content>
	</Sheet.Root>
</div>

<!-- lg and up -->
<div
	class="hidden h-[calc(100vh-var(--header-height))] shrink-0 overflow-hidden lg:block lg:w-[20rem]"
>
	<nav
		class="border-alt-1 relative max-h-[calc(100vh-var(--header-height))] min-h-[calc(100vh-var(--header-height))] overflow-y-auto overflow-x-hidden border-r"
	>
		<!-- Course title -->
		<div class="bg-background sticky top-0 z-[1] flex flex-row gap-3 border-b py-5 pl-1.5 pr-3">
			<span class="grow text-sm">{title}</span>

			<span>
				<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
					<Tooltip.Trigger asChild let:builder>
						<Button
							builders={[builder]}
							variant="ghost"
							href="/settings/courses/details?id={id}"
							class="text-muted-foreground hover:text-foreground mt-1 h-auto px-0 py-0 hover:bg-transparent"
						>
							<Info class="size-4 shrink-0" />
						</Button>
					</Tooltip.Trigger>

					<Tooltip.Content
						class="bg-foreground text-background select-none rounded-sm border-none px-1.5 py-1 text-xs"
						transitionConfig={{ y: 8, duration: 100 }}
						side="bottom"
					>
						Details
						<Tooltip.Arrow class="bg-background" />
					</Tooltip.Content>
				</Tooltip.Root>
			</span>
		</div>

		<ul class="ml-auto h-full columns-1 pt-7">
			{#each Object.keys(chapters) as chapter}
				<li class="pb-8 leading-5">
					<!-- Chapter heading -->
					<span
						class="after:bg-alt-1 relative flex w-full pr-2 text-base font-semibold tracking-wide after:absolute after:-bottom-1 after:left-0 after:h-px after:w-full"
					>
						{chapter}
					</span>

					<!-- Assets -->
					<ul class="pr-3 pt-3">
						{#each chapters[chapter] as asset}
							<li class="pl-1.5">
								<!-- Asset -->
								<Button
									variant="ghost"
									class={cn(
										'h-auto w-full justify-start whitespace-normal px-0 py-0 text-start hover:bg-transparent hover:underline',
										asset.id === selectedAsset?.id
											? 'decoration-foreground'
											: 'decoration-muted-foreground'
									)}
									data-selected={asset.id === selectedAsset?.id}
									on:click={() => {
										UpdateQueryParam('a', asset.id, false);
									}}
								>
									<div class={cn('flex w-full flex-row justify-between py-1.5')}>
										<!-- Asset title -->
										<div
											class={cn(
												'text-muted-foreground grow pr-2.5',
												asset.id === selectedAsset?.id && 'text-foreground'
											)}
										>
											<span>{asset.prefix}.</span>
											{asset.title}
										</div>

										<!-- Asset completed -->
										<CircleCheck
											class={cn(
												'text-muted-foreground mt-0.5 size-4 shrink-0',
												asset.completed && 'fill-success text-success [&>:nth-child(2)]:text-white'
											)}
										/>
									</div>
								</Button>
							</li>
						{/each}
					</ul>
				</li>
			{/each}
		</ul>
	</nav>
</div>
