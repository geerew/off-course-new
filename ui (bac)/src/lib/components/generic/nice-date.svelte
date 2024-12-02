<script lang="ts">
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import ago from 's-ago';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export let date: string;
	export let prefix: string = '';
	export { className as class };

	// ----------------------
	// Variables
	// ----------------------
	const d = new Date(date);
	const utcDate = new Date(d.getTime() + d.getTimezoneOffset() * 60000);

	const formattedDate = utcDate.toLocaleDateString(undefined, {
		weekday: 'long',
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		second: 'numeric'
	});
</script>

<div class={cn('text-muted-foreground', className)}>
	<span title={prefix !== '' ? `${prefix} ${formattedDate}` : formattedDate}>
		{ago(utcDate)}
	</span>
</div>
