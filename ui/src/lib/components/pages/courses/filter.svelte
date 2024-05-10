<script lang="ts">
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import type { CourseProgress } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { Filter, Loader2, Tag, X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { CoursesProgressFilter, CoursesTagsFilter } from '.';

	// ----------------------
	// Variables
	// ----------------------
	let selectedTags: Record<string, string> = {};
	let selectedProgress: CourseProgress | undefined;

	const dispatchEvent = createEventDispatcher();
</script>

<!-- Filters -->
<div class="border-alt-1/60 flex w-full flex-row gap-5 border-b pb-5">
	<CoursesTagsFilter
		bind:selectedTags
		on:change={() => {
			dispatchEvent('tagsFilter', Object.values(selectedTags));
		}}
	/>

	<CoursesProgressFilter
		bind:progress={selectedProgress}
		on:change={() => {
			dispatchEvent('progressFilter', selectedProgress);
		}}
	/>
</div>

{#if Object.values(selectedTags).length > 0 || selectedProgress}
	<div class="border-alt-1/60 flex flex-col gap-4 border-b pb-5">
		<div class="text-primary flex flex-row items-center gap-2.5 text-sm">
			<Filter class="size-4" />
			<span class="tracking-wide">ACTIVE FILTERS</span>
		</div>

		<div class="flex flex-row gap-2">
			<!-- Progress -->

			{#if selectedProgress}
				<div class="flex flex-row" data-progress={selectedProgress}>
					<Badge
						class={cn(
							'bg-alt-1/60 hover:bg-alt-1/60 text-foreground min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none'
						)}
					>
						<Loader2 class="size-3" />
						<span>{selectedProgress}</span>
					</Badge>

					<Button
						class={cn(
							'bg-alt-1/60 hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5 duration-200'
						)}
						on:click={() => {
							selectedProgress = undefined;
							dispatchEvent('progressFilter', selectedProgress);
						}}
					>
						<X class="size-3" />
					</Button>
				</div>
			{/if}

			<!-- Tags -->
			{#each Object.keys(selectedTags) as id}
				<div class="flex flex-row" data-tag={selectedTags[id]}>
					<Badge
						class={cn(
							'bg-alt-1/60 hover:bg-alt-1/60 text-foreground min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none'
						)}
					>
						<Tag class="size-3" />
						<span>{selectedTags[id]}</span>
					</Badge>

					<Button
						class={cn(
							'bg-alt-1/60 hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5 duration-200'
						)}
						on:click={() => {
							delete selectedTags[id];
							selectedTags = { ...selectedTags };
							dispatchEvent('tagsFilter', Object.values(selectedTags));
						}}
					>
						<X class="size-3" />
					</Button>
				</div>
			{/each}

			<Button
				class={cn(
					'bg-primary hover:bg-primary inline-flex h-auto items-center rounded-lg px-2.5 py-0.5 duration-200 hover:brightness-110'
				)}
				on:click={() => {
					selectedTags = {};
					selectedProgress = undefined;
					dispatchEvent('clear');
				}}
			>
				Clear all
			</Button>
		</div>
	</div>
{/if}
