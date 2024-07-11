<script lang="ts">
	import { Icons } from '$components/icons';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import Separator from '$components/ui/separator/separator.svelte';
	import { CourseProgress } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import { CoursesProgressFilter, CoursesTagsFilter, CoursesTitleFilter } from '.';

	// ----------------------
	// Variables
	// ----------------------
	let filterTitles: string[] = [];
	let filterProgress: CourseProgress | undefined;
	let filterTags: Record<string, string> = {};

	const dispatchEvent = createEventDispatcher();

	// ----------------------
	// Reactive
	// ----------------------
	$: isFiltering =
		filterTitles.length > 0 || Object.values(filterTags).length > 0 || filterProgress;
</script>

<!-- Filters -->
<div class="flex w-full flex-col gap-5 border-b border-alt-1/60 pb-5 md:flex-row">
	<CoursesTitleFilter
		on:change={(e) => {
			filterTitles = [...filterTitles, e.detail];
			dispatchEvent('titleFilter', filterTitles);
		}}
	/>

	<div class="flex gap-2.5 md:gap-5">
		<CoursesProgressFilter
			bind:progress={filterProgress}
			on:change={() => {
				dispatchEvent('progressFilter', filterProgress);
			}}
		/>

		<CoursesTagsFilter
			bind:filterTags
			on:change={() => {
				dispatchEvent('tagsFilter', Object.values(filterTags));
			}}
		/>
	</div>
</div>

{#if isFiltering}
	<div class="flex flex-col gap-4 border-b border-alt-1/60 pb-5">
		<div class="flex flex-row items-center gap-2.5 text-sm text-primary">
			<Icons.Filter class="size-4" />
			<span class="tracking-wide">ACTIVE FILTERS</span>
		</div>

		<div class="flex flex-row items-center gap-2">
			<!-- Titles -->
			{#each filterTitles as title}
				<div class="flex flex-row" data-title={title}>
					<Badge
						class={cn(
							'h-6 min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1/60 text-foreground hover:bg-alt-1/60'
						)}
					>
						<Icons.Text class="size-3" />
						<span>{title}</span>
					</Badge>

					<Button
						class={cn(
							'inline-flex h-6 items-center rounded-l-none rounded-r-sm border-l bg-alt-1/60 px-1.5 py-0.5 duration-200 hover:bg-destructive'
						)}
						on:click={() => {
							filterTitles = filterTitles.filter((t) => t !== title);
							filterTitles = [...filterTitles];
							dispatchEvent('titleFilter', filterTitles);
						}}
					>
						<Icons.X class="size-3" />
					</Button>
				</div>
			{/each}

			<!-- Progress -->
			{#if filterProgress}
				{#if filterTitles.length > 0}
					<Separator orientation="vertical" class="h-6 bg-alt-1" />
				{/if}

				<div class="flex flex-row" data-progress={filterProgress}>
					<Badge
						class={cn(
							'h-6 min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1/60 text-foreground hover:bg-alt-1/60'
						)}
					>
						<Icons.Hourglass class="size-3" />
						<span>{filterProgress}</span>
					</Badge>

					<Button
						class={cn(
							'inline-flex h-6 items-center rounded-l-none rounded-r-sm border-l bg-alt-1/60 px-1.5 py-0.5 duration-200 hover:bg-destructive'
						)}
						on:click={() => {
							filterProgress = undefined;
							dispatchEvent('progressFilter', filterProgress);
						}}
					>
						<Icons.X class="size-3" />
					</Button>
				</div>
			{/if}

			<!-- Tags -->
			{#if Object.keys(filterTags).length > 0}
				{#if filterTitles.length > 0 || filterProgress}
					<Separator orientation="vertical" class="h-6 bg-alt-1" />
				{/if}

				{#each Object.keys(filterTags) as id}
					<div class="flex flex-row" data-tag={filterTags[id]}>
						<Badge
							class={cn(
								'h-6 min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1/60 text-foreground hover:bg-alt-1/60'
							)}
						>
							<Icons.Tag class="size-3" />
							<span>{filterTags[id]}</span>
						</Badge>

						<Button
							class={cn(
								'inline-flex h-6 items-center rounded-l-none rounded-r-sm border-l bg-alt-1/60 px-1.5 py-0.5 duration-200 hover:bg-destructive'
							)}
							on:click={() => {
								delete filterTags[id];
								filterTags = { ...filterTags };
								dispatchEvent('tagsFilter', Object.values(filterTags));
							}}
						>
							<Icons.X class="size-3" />
						</Button>
					</div>
				{/each}
			{/if}

			<Button
				class={cn(
					'inline-flex h-6 items-center rounded-lg bg-primary px-2.5 py-0.5 duration-200 hover:bg-primary hover:brightness-110'
				)}
				on:click={() => {
					filterTitles = [];
					filterTags = {};
					filterProgress = undefined;
					dispatchEvent('clear');
				}}
			>
				Clear all
			</Button>
		</div>
	</div>
{/if}
