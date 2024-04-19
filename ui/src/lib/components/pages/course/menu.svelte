<script lang="ts">
	import Button from '$components/ui/button/button.svelte';
	import * as Tooltip from '$components/ui/tooltip';
	import type { Asset, CourseChapters } from '$lib/types/models';
	import { cn, flyAndScale } from '$lib/utils';
	import { CircleCheck, Info } from 'lucide-svelte';

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
</script>

<div
	class="fixed left-0 hidden h-[calc(100vh-var(--header-height))] overflow-hidden lg:block lg:w-[var(--course-menu-width)] xl:w-[max(var(--course-menu-width),23vw)]"
>
	<nav
		class="bg-muted before:bg-alt-1 relative left-0 top-0 max-h-[calc(100vh-var(--header-height))] min-h-[calc(100vh-var(--header-height))] overflow-y-auto overflow-x-hidden before:absolute before:right-0 before:top-0 before:h-full before:w-px"
	>
		<ul class="ml-auto h-full w-80 columns-1 pl-8 pt-7">
			<!-- Course title -->
			<div class="flex flex-row gap-3 pb-8 pr-3">
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
							transition={flyAndScale}
							transitionConfig={{ y: 8, duration: 100 }}
							side="bottom"
						>
							Details
							<Tooltip.Arrow class="bg-background" />
						</Tooltip.Content>
					</Tooltip.Root>
				</span>
			</div>

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
									on:click={() => {
										selectedAsset = asset;
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
