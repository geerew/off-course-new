<script lang="ts">
	import type { Course } from '$lib/types/models';
	import { GetScanByCourseId } from '$lib/utils/api';
	import { cn } from '$lib/utils/general';
	import { createEventDispatcher, onDestroy, onMount } from 'svelte';
	import TableCourseAction from './TableCourseAction.svelte';
	import TableDate from './TableDate.svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let course: Course;

	// ----------------------
	// Variables
	// ----------------------

	// On mount, if the course has a scan status of either waiting or processing, start polling
	// for updates. This variable will be set the first time the function is called and is used to
	// stop the polling on destroy
	let scanPoll = -1;

	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	// When the scan status is set to either waiting or processing, start polling for updates. As
	// the status changes, we dispatch an event to the parent component to update the courses list.
	// When the scan finishes, we clear the interval and set the status to an empty string.
	const startPolling = () => {
		scanPoll = setInterval(async () => {
			await GetScanByCourseId(course.id, true)
				.then((resp) => {
					if (resp && resp.status !== course.scanStatus) {
						course.scanStatus = resp.status;
						dispatch('change');
					}
				})
				.catch(() => {
					// Either the scan finished or there was an error
					// TODO: handle this better
					course.scanStatus = '';
					clearInterval(scanPoll);
					scanPoll = -1;
					dispatch('change');
				});
		}, 1500);
	};

	// ----------------------
	// Reactive
	// ----------------------
	$: isProcessing = course.scanStatus === 'processing';

	$: course.scanStatus && scanPoll === -1 && startPolling();

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		if (course.scanStatus && scanPoll === -1) startPolling();
	});

	onDestroy(() => {
		if (scanPoll !== -1) {
			clearInterval(scanPoll);
			scanPoll = -1;
		}
	});
</script>

<div class="flex flex-row items-start gap-2">
	<!-- Shown on md+ -->
	<div class="flex w-full flex-col gap-3 md:flex-row md:items-center md:gap-2.5">
		<div class="flex w-full items-center gap-4">
			<span class="grow">{course.title}</span>

			{#if course.scanStatus}
				<span
					title={isProcessing ? 'scanning course' : 'waiting for scan to start'}
					class={cn(
						'leading hidden items-center justify-center gap-2 whitespace-nowrap rounded border px-1.5 py-0.5 text-center text-xs md:inline-flex',
						isProcessing ? 'text-success animate-pulse' : 'text-foreground-muted'
					)}>{isProcessing ? 'scanning' : 'queued'}</span
				>
			{/if}
		</div>

		<!-- Shown on sm- -->
		<div
			class={cn(
				'grid grid-cols-2 grid-rows-2 gap-2.5 text-xs sm:grid-cols-3 sm:grid-rows-1 md:hidden',
				!course.scanStatus && 'grid-rows-1'
			)}
		>
			<div class="flex flex-row gap-1">
				<span class="text-foreground-muted font-semibold">Added:</span>
				<TableDate date={course.createdAt} />
			</div>

			<div class="flex flex-row gap-1">
				<span class="text-foreground-muted font-semibold">Updated:</span>
				<TableDate date={course.updatedAt} />
			</div>

			{#if course.scanStatus}
				<div class="flex flex-row items-center gap-1">
					<span class="text-foreground-muted font-semibold">Scan status:</span>
					<span
						title={isProcessing ? 'scanning course' : 'waiting for scan to start'}
						class={cn(isProcessing ? 'text-success animate-pulse' : 'text-foreground-muted')}
					>
						{isProcessing ? 'scanning' : 'queued'}
					</span>
				</div>
			{/if}
		</div>
	</div>

	<div class="md:hidden">
		<TableCourseAction {course} on:delete on:scan />
	</div>
</div>
