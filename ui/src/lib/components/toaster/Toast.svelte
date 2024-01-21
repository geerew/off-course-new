<script lang="ts">
	import type { ToastData } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import { createProgress, type Toast, type ToastsElements } from '@melt-ui/svelte';
	import { AlertCircle, CheckCircle2, Info, X, XCircle } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import { fly } from 'svelte/transition';

	// ----------------------
	// Exports
	// ----------------------
	export let elements: ToastsElements;
	$: ({ content, description, close } = elements);

	export let toast: Toast<ToastData>;
	$: ({ data, id, getPercentage } = toast);

	// ----------------------
	// Variables
	// ----------------------
	const percentage = writable(0);
	const {
		elements: { root: progress },
		options: { max }
	} = createProgress({
		max: 100,
		value: percentage
	});

	// ----------------------
	// Reactive
	// ----------------------
	$: timerStyle = cn(
		data && data.status === 'info' && 'bg-primary',
		data && data.status === 'success' && 'bg-success',
		data && data.status === 'warning' && 'bg-orange-500',
		data && data.status === 'error' && 'bg-error'
	);

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(() => {
		let frame: number;
		const updatePercentage = () => {
			percentage.set(getPercentage());
			frame = requestAnimationFrame(updatePercentage);
		};
		frame = requestAnimationFrame(updatePercentage);

		return () => cancelAnimationFrame(frame);
	});
</script>

<div
	{...$content(id)}
	use:content
	in:fly={{ duration: 150, y: -50 }}
	out:fly={{ duration: 150, y: -50 }}
	class={cn('bg-accent-1 relative overflow-hidden rounded-md border text-white shadow-md')}
>
	<!-- Status & close button -->
	<div class="flex flex-row items-center justify-between">
		<div class="py-1 pl-2">
			{#if data.status === 'info'}
				<Info class="text-primary mr-2 h-6 w-6" />
			{:else if data.status === 'success'}
				<CheckCircle2 class="text-success mr-2 h-6 w-6" />
			{:else if data.status === 'warning'}
				<AlertCircle class="mr-2 h-6 w-6 text-orange-500" />
			{:else if data.status === 'error'}
				<XCircle class="text-error mr-2 h-6 w-6" />
			{/if}
		</div>

		<button
			{...$close(id)}
			use:close
			class="group z-50 rounded-md p-1 text-sm font-semibold duration-200"
		>
			<X class="group-hover:text-foreground text-muted-foreground h-4 w-4 duration-200" />
		</button>
	</div>

	<!-- Message -->
	<div
		class="relative flex min-w-[15rem] max-w-[23rem] items-center justify-between gap-4 px-2.5 pb-2.5 pt-1.5"
	>
		<div>
			<div class="text-foreground" {...$description(id)} use:description>
				{@html data.message}
			</div>
		</div>
	</div>

	<!-- Progress bar -->
	<div {...$progress} use:progress class="bg-accent-1 h-1.5 overflow-hidden">
		<div
			class={cn('h-full w-full rounded-full', timerStyle)}
			style={`transform: translateX(-${100 - (100 * ($percentage ?? 0)) / ($max ?? 1)}%)`}
		/>
	</div>
</div>
