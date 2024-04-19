<script lang="ts">
	import { cn } from '$lib/utils';
	import type { CollapsibleElements, CollapsibleStates } from '@melt-ui/svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let trigger: CollapsibleElements['trigger'];
	export let open: CollapsibleStates['open'];
</script>

<button
	{...$trigger}
	use:trigger
	class="hover:bg-accent-1 group flex items-center justify-center rounded-md p-3"
	aria-label="mobile navigation"
	aria-expanded={$open}
>
	<span class="relative h-[16px] w-4 overflow-hidden">
		<span
			class={cn(
				'bg-foreground absolute left-0 h-0.5 w-4 rounded-sm duration-200',
				$open
					? 'top-1.5 translate-x-0 rotate-[225deg]'
					: 'top-0 translate-x-0 group-hover:translate-x-1'
			)}
		></span>
		<span
			class={cn(
				'bg-foreground absolute left-0 top-1.5 h-[1.5px] w-4 rounded-sm duration-200',
				$open ? 'translate-x-4' : 'translate-x-2 group-hover:translate-x-0'
			)}
		></span>
		<span
			class={cn(
				'bg-foreground absolute left-0 h-0.5 w-4 rounded-sm duration-200',
				$open
					? 'top-1.5 translate-x-0 rotate-[135deg]'
					: 'top-3 translate-x-1 group-hover:translate-x-2'
			)}
		></span>
	</span>
</button>
