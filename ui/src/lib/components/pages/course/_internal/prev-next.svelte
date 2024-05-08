<script lang="ts">
	import { Button } from '$components/ui/button';
	import type { Asset } from '$lib/types/models';
	import { ArrowLeft, ArrowRight } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let prevAsset: Asset | null;
	export let nextAsset: Asset | null;

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();
</script>

<div class="flex w-full flex-row gap-3 pt-5">
	{#if prevAsset}
		<Button
			variant="outline"
			class="text-muted-foreground hover:text-foreground hover:bg-background hover:border-alt-1 flex h-auto basis-1/2 flex-row items-center justify-start gap-4 whitespace-normal rounded-sm border p-3 text-start"
			on:click={() => {
				dispatch('prev');
			}}
		>
			<span class="text-start">
				<ArrowLeft class="size-5" />
			</span>
			{prevAsset.prefix}. {prevAsset.title}
		</Button>
	{:else}
		<div class="basis-1/2"></div>
	{/if}

	{#if nextAsset}
		<Button
			variant="outline"
			class="text-muted-foreground hover:text-foreground hover:bg-background hover:border-alt-1 flex h-auto basis-1/2 flex-row items-center justify-end gap-4 whitespace-normal rounded-sm border p-3 text-end"
			on:click={() => {
				dispatch('next');
			}}
		>
			<span class="text-start">
				{nextAsset.prefix}. {nextAsset.title}
			</span>
			<ArrowRight class="size-5" />
		</Button>
	{:else}
		<div class="basis-1/2"></div>
	{/if}
</div>
