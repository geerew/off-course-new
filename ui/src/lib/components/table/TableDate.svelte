<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils/general';
	import ago from 's-ago';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export let date: string;
	export { className as class };

	// ----------------------
	// Variables
	// ----------------------
	const formattedDate = new Date(date).toLocaleDateString(undefined, {
		weekday: 'long',
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		second: 'numeric'
	});
</script>

<div class={cn('text-foreground-muted', className)}>
	<span title={formattedDate}>
		{ago(new Date(date))}
	</span>
</div>
