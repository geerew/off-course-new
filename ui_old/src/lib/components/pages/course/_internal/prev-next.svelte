<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import type { Asset } from '$lib/types/models';
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

<div class="flex w-full flex-col gap-3 pt-5 md:flex-row">
	{#if prevAsset}
		{#key prevAsset.id}
			<Button
				variant="outline"
				class="flex h-auto flex-row items-center justify-start gap-4 whitespace-normal rounded-sm border p-3 text-start text-muted-foreground hover:border-alt-1 hover:bg-background hover:text-foreground md:basis-1/2"
				on:click={() => {
					dispatch('prev');
				}}
			>
				<span class="text-start">
					<Icons.ArrowLeft class="size-5" />
				</span>
				{prevAsset.prefix}. {prevAsset.title}
			</Button>
		{/key}
	{:else}
		<div class="basis-1/2"></div>
	{/if}

	{#if nextAsset}
		{#key nextAsset.id}
			<Button
				variant="outline"
				class="flex h-auto flex-row place-content-end items-center justify-end gap-4 whitespace-normal rounded-sm border p-3 text-end text-muted-foreground hover:border-alt-1 hover:bg-background hover:text-foreground md:basis-1/2"
				on:click={() => {
					dispatch('next');
				}}
			>
				<span class="text-start">
					{nextAsset.prefix}. {nextAsset.title}
				</span>
				<Icons.ArrowRight class="size-5" />
			</Button>
		{/key}
	{:else}
		<div class="basis-1/2"></div>
	{/if}
</div>
