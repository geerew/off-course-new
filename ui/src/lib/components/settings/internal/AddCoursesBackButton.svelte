<script lang="ts">
	import { Loading } from '$components';
	import { Icons } from '$components/icons';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------

	// True when a path is loading. It is used to disable this back button
	export let loadingPath: boolean;

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();

	// True when this back button was clicked. It is used to disable this back button and render
	// a loading icon
	let clickedThis = false;
</script>

<button
	class="enabled:hover:bg-accent-1 group flex h-14 flex-row items-center border-b pl-3"
	disabled={loadingPath || clickedThis}
	tabindex="-1"
	on:click={(e) => {
		clickedThis = true;
		dispatch('click', e);
	}}
>
	<div class="flex grow gap-2 text-sm">
		<Icons.cornerUpLeft class="text-foreground-muted group-hover:text-foreground h-4 w-4" />
		<span>..</span>
	</div>

	{#if loadingPath && clickedThis}
		<div class="flex h-full w-14 shrink-0 place-content-center items-center duration-200 sm:w-20">
			<Loading class="border-primary h-5 w-5 border-[3px]" />
		</div>
	{/if}
</button>
