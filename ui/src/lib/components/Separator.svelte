<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils/general';
	import { createSeparator, type CreateSeparatorProps } from '@melt-ui/svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
	export let orientation: CreateSeparatorProps['orientation'] = 'vertical';

	// ----------------------
	// Variables
	// ----------------------

	const {
		elements: { root }
	} = createSeparator({
		orientation
	});
</script>

{#if orientation === 'vertical'}
	<div {...$root} use:root class={cn('bg-border h-8 w-[1px]', className)} />
{:else}
	<div {...$root} use:root class={cn('bg-border h-1 w-full', className)} />
{/if}
