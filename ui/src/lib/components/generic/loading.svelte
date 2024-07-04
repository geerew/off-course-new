<script lang="ts">
	import { Icons } from '$components/icons';
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
	export let loaderClass: ClassName = undefined;
</script>

<div
	class={cn(
		'flex min-h-40 w-full flex-grow flex-col place-content-center items-center gap-5 p-10',
		className
	)}
>
	<Icons.Loader class={cn('text-primary size-12 animate-spin', loaderClass)} />
</div>
