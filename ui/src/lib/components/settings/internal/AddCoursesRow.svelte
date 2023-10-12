<script lang="ts">
	import { Loading } from '$components';
	import Separator from '$components/Separator.svelte';
	import type { FileInfo } from '$lib/types/fileSystem';
	import { cn } from '$lib/utils/general';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	// Row info, namely title, path, whether this directory is a parent of another course, or if
	// this course was previously add, etc
	export let data: FileInfo;

	// True when this or another path is loading. It is used to disable clicking on this and other
	// paths
	export let loadingPath: boolean;

	// ----------------------
	// Variables
	// ----------------------

	// True when this path was the clicked path
	let clickedThis = false;

	// Used to dispatch events. Currently fires add, remove and click events
	const dispatch = createEventDispatcher();

	// ----------------------
	// Reactive
	// ----------------------

	// True when this course was previously added or it is currently selected
	$: checked = (data.isSelected || data.isExistingCourse) ?? false;
</script>

<div class="flex h-14 flex-row items-center border-b">
	<button
		tabindex="-1"
		disabled={loadingPath || clickedThis || data.isExistingCourse || checked}
		class={cn(
			'enabled:hover:bg-accent-1 group flex h-full grow items-center gap-4 px-2 sm:px-5',
			checked && !data.isExistingCourse && 'text-primary',
			data.isExistingCourse && 'text-foreground-muted'
		)}
		on:click={(e) => {
			clickedThis = true;
			dispatch('click', e);
		}}
	>
		<span class="flex grow text-sm">{data.title}</span>

		{#if data.isExistingCourse}
			<span
				class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded border px-1.5 py-1 text-center text-xs"
				>added</span
			>
		{/if}
	</button>

	{#if !data.isExistingCourse}
		<Separator class="h-14" />

		{#if loadingPath && clickedThis}
			<!-- Show loading icon -->
			<div class="flex h-full w-14 shrink-0 place-content-center items-center duration-200 sm:w-20">
				<Loading class="border-primary h-5 w-5 border-[3px]" />
			</div>
		{:else}
			<button
				disabled={data.isParent ?? false}
				class="enabled:hover:bg-accent-1 group flex h-full w-14 shrink-0 place-content-center items-center duration-200 sm:w-20"
				on:click={() => {
					checked = !checked;
					checked ? dispatch('add') : dispatch('remove');
				}}
			>
				<input tabindex="0" bind:checked type="checkbox" indeterminate={data.isParent ?? false} />
			</button>
		{/if}
	{/if}
</div>

<style lang="postcss">
	input {
		@apply bg-background pointer-events-none cursor-pointer rounded border-2 p-2 duration-150;
		@apply border-foreground/40;
		@apply checked:bg-primary checked:hover:bg-primary checked:border-transparent;
		@apply indeterminate:bg-foreground-muted/50 indeterminate:border-transparent;
		@apply outline-none focus:ring-0 focus:ring-offset-0;
	}
</style>
