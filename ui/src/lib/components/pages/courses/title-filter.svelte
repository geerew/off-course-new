<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Variables
	// ----------------------
	let titleValue = '';
	let inputEl: HTMLInputElement;
	const dispatchEvent = createEventDispatcher();
</script>

<div
	class="group relative w-64 rounded-md border border-alt-1/60 focus-within:border-alt-1/100 md:w-56"
>
	<label for="tags-input">
		<Icons.Search class="absolute left-2 top-1/2 size-3 -translate-y-1/2 text-muted-foreground" />
	</label>

	<input
		id="tags-input"
		bind:this={inputEl}
		class="w-full rounded-md border border-none border-alt-1/60 bg-inherit px-7 text-sm text-foreground placeholder-muted-foreground/60 focus-visible:outline-none focus-visible:ring-0"
		placeholder="Filter titles"
		bind:value={titleValue}
		on:keydown={(e) => {
			if (e.key === 'Enter' && titleValue.trim().length > 0) {
				dispatchEvent('change', titleValue.trim());
				titleValue = '';
			}
		}}
	/>

	{#if titleValue.length > 0}
		<Button
			class="absolute right-1 top-1/2 h-auto -translate-y-1/2 transform px-2 py-1 text-muted-foreground hover:bg-inherit hover:text-foreground"
			variant="ghost"
			on:click={() => {
				titleValue = '';
				inputEl.focus();
			}}
		>
			<Icons.X class="size-3" />
		</Button>
	{/if}
</div>
