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
	export let errorMessage = '';
</script>

<div
	class={cn(
		'flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center gap-5 p-10 text-xl font-bold',
		className
	)}
>
	<CircleOff class={cn('text-destructive h-12 w-12', imgClass)} />

	<div class="flex flex-col gap-2.5 text-center">
		<span class={cn('text-foreground', textClass)}>Oh no! An error occurred</span>
		{#if errorMessage}
			<span class={cn('text-muted-foreground text-sm', textClass)}>{errorMessage}</span>
		{/if}
	</div>
</div>
