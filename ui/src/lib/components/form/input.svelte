<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLInputAttributes } from 'svelte/elements';

	type HTMLInputAttributesWithRef = HTMLInputAttributes & { ref?: HTMLInputElement | null };

	let {
		value = $bindable(''),
		ref = $bindable(null),
		class: containerClass = '',
		...restProps
	}: HTMLInputAttributesWithRef = $props();
</script>

<input
	bind:value
	bind:this={ref}
	class={cn(
		'bg-background-alt-3 placeholder:text-foreground-alt-2 w-full rounded-md border border-transparent p-2.5 ring-0 duration-250 ease-in-out placeholder:text-sm placeholder:tracking-wide focus:brightness-110 focus:outline-none',
		containerClass
	)}
	{...restProps}
/>
