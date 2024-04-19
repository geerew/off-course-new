<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import { CircleOff } from 'lucide-svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
	export let imgClass: ClassName = undefined;
	export let textClass: ClassName = undefined;
	export let errorMessage = 'Oh no! An error occurred.';
</script>

<div
	class={cn(
		'flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center gap-5 p-10 text-xl font-bold',
		className
	)}
>
	<CircleOff class={cn('text-destructive h-12 w-12', imgClass)} />
	<span class={cn('text-destructive-foreground', textClass)}>{errorMessage}</span>
</div>
