<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils/general';
	import { Icons } from './icons';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
	export let imgClass: ClassName = undefined;
</script>

<div
	class={cn(
		'flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center gap-5 p-10 text-xl font-bold',
		className
	)}
>
	<Icons.circleOff class={cn('text-error h-12 w-12', imgClass)} />
	<span> Oops, an error occurred</span>
</div>
