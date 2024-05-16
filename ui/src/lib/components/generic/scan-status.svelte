<script lang="ts">
	import { GetCourse } from '$lib/api';
	import type { ClassName } from '$lib/types/general';
	import { cn } from '$lib/utils';
	import { createEventDispatcher, onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;
	export let waitingText = 'queued';
	export let processingText = 'processing';
	export { className as class };

	// ----------------------
	// Variables
	// ----------------------

	// Current scan status
	let scanStatus = '';

	// When the scanStatus is not empty, this will set. It is used to determine if we should stop
	// polling during onDestroy()
	let scanPoll = -1;

	// Dispatch events to as the status changes
	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	// When the scan status is set to either waiting or processing, start polling for updates. As
	// the status changes, we dispatch an event to the parent component to update the courses list.
	// When the scan finishes, we clear the interval and set the status to an empty string.
	function startPolling() {
		scanPoll = setInterval(async () => {
			try {
				const response = await GetCourse(courseId);

				if (!response) throw new Error('Course not found');

				scanStatus = response.scanStatus;

				if (scanStatus === 'waiting') {
					dispatch('waiting', response);
				} else if (scanStatus === 'processing') {
					dispatch('processing', response);
				} else {
					dispatch('empty', response);
					clearInterval(scanPoll);
					scanPoll = -1;
				}
			} catch (error) {
				toast.error(error instanceof Error ? error.message : (error as string));

				scanStatus = '';
				clearInterval(scanPoll);
				scanPoll = -1;
			}
		}, 1000);
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		if (scanPoll === -1) {
			startPolling();
		}

		return () => {
			if (scanPoll !== -1) {
				clearInterval(scanPoll);
				scanPoll = -1;
			}
		};
	});
</script>

<div class={cn('text-muted-foreground flex items-center justify-center', className)}>
	{#if !scanStatus}
		-
	{:else if scanStatus === 'waiting'}
		{waitingText}
	{:else}
		<span class="text-secondary">{processingText}</span>
	{/if}
</div>
