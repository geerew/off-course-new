<script lang="ts">
	import { Checkbox } from '$components/ui/checkbox';
	import { cn } from '$lib/utils';
	import { type Writable } from 'svelte/store';

	//----------------------
	// Exports
	//----------------------

	export let selectedTagsCount: Writable<number>;
	export let totalItems: number;

	//----------------------
	// Variables
	//----------------------

	let checkedOrIndeterminate: boolean | 'indeterminate' = 'indeterminate';

	//----------------------
	// Reactive
	//----------------------

	// Determine if the checkbox should be unchecked, checked or indeterminate
	$: checkedOrIndeterminate =
		$selectedTagsCount === totalItems ? true : $selectedTagsCount > 0 ? 'indeterminate' : false;
</script>

<Checkbox
	on:click
	class={cn('border-alt-1', totalItems === 0 && 'opacity-0')}
	bind:checked={checkedOrIndeterminate}
/>
