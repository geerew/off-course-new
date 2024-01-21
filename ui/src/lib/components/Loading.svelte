<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import { Loader2 } from 'lucide-svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export { className as class };
</script>

<div class="flex w-full place-content-center items-center">
	<Loader2 class={cn('text-primary h-12 w-16 animate-spin', className)} />
</div>
