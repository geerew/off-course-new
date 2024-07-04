<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Variables
	// ----------------------
	let messageValue = '';
	let inputEl: HTMLInputElement;
	const dispatchEvent = createEventDispatcher();
</script>

<div
	class="border-alt-1/60 focus-within:border-alt-1/100 group relative w-64 rounded-md border md:w-56"
>
	<label for="tags-input">
		<Icons.Search class="text-muted-foreground absolute left-2 top-1/2 size-4 -translate-y-1/2" />
	</label>

	<input
		id="tags-input"
		bind:this={inputEl}
		class="placeholder-muted-foreground/60 text-foreground border-alt-1/60 w-full rounded-md border border-none bg-inherit pl-8 pr-7 text-sm focus-visible:outline-none focus-visible:ring-0"
		placeholder="Filter messages..."
		bind:value={messageValue}
		on:keydown={(e) => {
			if (e.key === 'Enter' && messageValue.trim().length > 0) {
				dispatchEvent('change', messageValue.trim());
				messageValue = '';
			}
		}}
	/>

	{#if messageValue.length > 0}
		<Button
			class="text-muted-foreground hover:text-foreground absolute right-1 top-1/2 h-auto -translate-y-1/2 transform px-2 py-1 hover:bg-inherit"
			variant="ghost"
			on:click={() => {
				messageValue = '';
				inputEl.focus();
			}}
		>
			<Icons.X class="size-3" />
		</Button>
	{/if}
</div>
