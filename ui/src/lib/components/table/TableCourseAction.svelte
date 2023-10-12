<script lang="ts">
	import { Icons } from '$components/icons';
	import type { Course } from '$lib/types/models';
	import { createDropdownMenu } from '@melt-ui/svelte';
	import { createEventDispatcher } from 'svelte';
	import { fly } from 'svelte/transition';

	// ----------------------
	// Exports
	// ----------------------

	export let course: Course;

	// ----------------------
	// Variables
	// ----------------------

	const {
		elements: { trigger, menu, item, separator },
		states: { open }
	} = createDropdownMenu({
		forceVisible: true
	});

	const dispatch = createEventDispatcher();
</script>

<button
	{...$trigger}
	use:trigger
	class="hover:bg-accent-1 rounded-md p-1.5 text-sm font-semibold duration-200"
	aria-label="Menu"
>
	<Icons.moreHorizontal class="h-4 w-4" />
	<span class="sr-only">Open Menu</span>
</button>

{#if $open}
	<div class="menu" {...$menu} use:menu transition:fly={{ duration: 150, y: -10 }}>
		<a class="item hover:bg-accent-1" {...$item} use:item href="/courses/course?id={course.id}">
			Go to ...
		</a>

		<div class="bg-border/50 h-px" {...$separator} use:separator />

		<a
			class="item hover:bg-accent-1"
			{...$item}
			use:item
			href="/settings/courses/course?id={course.id}"
		>
			Details
		</a>

		<button
			class="item"
			disabled={course.scanStatus !== ''}
			{...$item}
			use:item
			on:click={() => {
				dispatch('scan', { id: course.id });
			}}>Scan</button
		>
		<button
			class="item data-[highlighted]:!bg-error text-error data-[highlighted]:text-white"
			{...$item}
			use:item
			on:click={() => {
				dispatch('delete', { id: course.id });
			}}
		>
			Delete
		</button>
	</div>
{/if}

<style lang="postcss">
	.menu {
		@apply z-10 flex w-[7rem] flex-col gap-1 rounded-md border py-1 shadow-md;
		@apply bg-background shadow-foreground/30 dark:shadow-foreground/20;
		@apply outline-none focus:outline-none;
	}

	.item {
		@apply mx-1 flex h-8 select-none items-center rounded-md py-2 pl-2 pr-1 text-sm leading-none;
		@apply enabled:data-[highlighted]:bg-accent-1;
		@apply disabled:text-foreground-muted;
	}
</style>
