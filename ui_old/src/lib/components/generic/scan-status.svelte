<script lang="ts">
	import { GetCourse } from '$lib/api';
	import type { ClassName } from '$lib/types/general';
	import type { ScanStatus } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { createEventDispatcher, onDestroy } from 'svelte';
	import { toast } from 'svelte-sonner';
	import type { Writable } from 'svelte/store';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;

	export let initialStatus: ScanStatus;

	export let poll: Writable<boolean>;

	export let waitingText = 'queued';
	export let processingText = 'scanning';
	export { className as class };

	// ----------------------
	// Variables
	// ----------------------

	// Current scan status
	let scanStatus: ScanStatus = '';

	// When the scanStatus is not empty, this will set. It is used to determine if we should stop
	// polling during onDestroy()
	let pollInterval = -1;

	// Dispatch events to as the status changes
	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	// When the scan status is set to either waiting or processing, start polling for updates. As
	// the status changes, we dispatch an event to the parent component to update the courses list.
	// When the scan finishes, we clear the interval and set the status to an empty string.
	function startPoll() {
		pollInterval = setInterval(async () => {
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
					poll.set(false);
				}
			} catch (error) {
				toast.error(error instanceof Error ? error.message : (error as string));

				scanStatus = '';
				poll.set(false);
			}
		}, 1000);
	}

	// ----------------------
	// Reactive
	// ----------------------

	// Start polling when the poll store is set to true and we are currently not polling. It
	// will set the scanStatus to 'waiting' so there is instant feedback to the user
	$: if ($poll && pollInterval === -1) {
		scanStatus = 'waiting';
		startPoll();
	}

	// Stop polling when the poll store is set to false
	$: if (!$poll) {
		clearInterval(pollInterval);
		pollInterval = -1;
	}

	$: scanStatus = initialStatus;

	// ----------------------
	// Lifecycle
	// ----------------------

	onDestroy(() => {
		if (pollInterval !== -1) {
			clearInterval(pollInterval);
			pollInterval = -1;
		}
	});
</script>

<div class={cn('flex items-center justify-center text-muted-foreground', className)}>
	{#if !scanStatus}
		-
	{:else if scanStatus === 'waiting'}
		{waitingText}
	{:else}
		<span class="text-secondary">{processingText}</span>
	{/if}
</div>
