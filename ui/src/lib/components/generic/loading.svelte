<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import { Loader2 } from 'lucide-svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
	export let loaderClass: ClassName = undefined;
</script>

<!-- <div
	class={cn(
		'flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center gap-5 p-10',
		className
	)}>
</div> -->

<div
	class={cn(
		'flex min-h-40 w-full flex-grow flex-col place-content-center items-center gap-5 p-10',
		className
	)}
>
	<Loader2 class={cn('text-primary h-12 w-16 animate-spin', loaderClass)} />
</div>
