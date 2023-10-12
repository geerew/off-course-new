<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils/general';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
</script>

<div class="flex w-full place-content-center items-center">
	<div
		class={cn(
			'loader border-success box-border inline-block h-8 w-8 animate-spin rounded-full border-[4px]',
			className,
			'border-b-transparent'
		)}
	/>
</div>
