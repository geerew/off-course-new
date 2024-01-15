<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { CourseCard, Error, Loading } from '$components';
	import { Icons } from '$components/icons';
	import { DeleteCourseDialog } from '$components/settings';
	import CourseAssetRow from '$components/settings/CourseAssetRow.svelte';
	import TableDate from '$components/table/TableDate.svelte';
	import { addToast } from '$lib/stores/addToast';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import {
		AddScan,
		ErrorMessage,
		GetAllCourseAssets,
		GetCourse,
		GetScanByCourseId
	} from '$lib/utils/api';
	import { buildChapterStructure, cn, isBrowser } from '$lib/utils/general';
	import { createAccordion, createDialog } from '@melt-ui/svelte';
	import type { AxiosError } from 'axios';
	import { onDestroy, onMount } from 'svelte';
	import { slide } from 'svelte/transition';

	// ----------------------
	// Variables
	// ----------------------

	// True while the page is loading
	let loadingPage = true;

	// True if the course id search param missing or not a valid course id
	let invalidCourseId = false;

	// Holds the information about the course being viewed
	let course: Course;

	// Holds the assets + attachments for this course
	let assets: Asset[];

	// Holds the course assets in a chapter structure. Populated when getCourse is called in
	// onMount
	let chapters: CourseChapters = {};

	// On mount, if the course has a scan status of either waiting or processing, start polling
	// for updates. This variable will be set the first time the function is called and is used to
	// stop the polling on destroy
	let scanPoll = -1;

	// Create a delete dialog
	const deleteDialog = createDialog({
		role: 'alertdialog'
	});

	// Get the open state of the delete dialog
	const openDeleteDialog = deleteDialog.states.open;

	// Accordion to show the course chapters
	const {
		elements: { content, item, trigger },
		helpers: { isSelected }
	} = createAccordion({
		multiple: true
	});

	// ----------------------
	// Functions
	// ----------------------

	// Gets the course id from the search params and queries the api for the course
	const getCourse = async () => {
		if (!isBrowser) return false;

		const params = isBrowser && $page.url.searchParams;
		const id = params && params.get('id');
		if (!id) return false;

		return await GetCourse(id)
			.then(async (resp) => {
				if (!resp) return false;
				course = { ...resp };
				return true;
			})
			.catch((err) => {
				const errMsg = ErrorMessage(err);
				console.error(errMsg);
				$addToast({
					data: {
						message: errMsg,
						status: 'error'
					}
				});
				return false;
			});
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Gets the assets + attachments for the given course. It will then build a chapter structure
	// for the assets and selected the first asset that is not completed. If the course itself is
	// completed, the first asset will be selected
	const getAssets = async (courseId: string) => {
		if (!isBrowser) return false;

		return await GetAllCourseAssets(courseId)
			.then(async (resp) => {
				if (!resp) return false;
				chapters = buildChapterStructure(resp);
				return true;
			})
			.catch((err) => {
				const errMsg = ErrorMessage(err);
				console.error(errMsg);
				$addToast({
					data: {
						message: errMsg,
						status: 'error'
					}
				});

				return false;
			});
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the scan status is set to either waiting or processing, start polling for updates.
	// When the scan finishes, clear the interval and set the status to an empty string.
	const startPolling = () => {
		scanPoll = setInterval(async () => {
			await GetScanByCourseId(course.id)
				.then((resp) => {
					if (resp && resp.status !== course.scanStatus) {
						course.scanStatus = resp.status;
					}
				})
				.catch(async (err: AxiosError) => {
					if (err.response?.status === 404) {
						// Scan is not longer found which means it completed (or was stopped). Get
						// the latest course details
						await getCourse();
					} else {
						const errMsg = ErrorMessage(err);
						console.error(errMsg);
						$addToast({
							data: {
								message: ErrorMessage(errMsg),
								status: 'error'
							}
						});
					}

					course.scanStatus = '';
					clearInterval(scanPoll);
					scanPoll = -1;
				});
		}, 1500);
	};

	// ----------------------
	// Reactive
	// ----------------------

	// Update the iProcessing variable when the scan status changes
	$: isProcessing = course && course.scanStatus === 'processing';

	// Start a poll when the scan status changes and scanPoll is -1
	$: course && course.scanStatus && scanPoll === -1 && startPolling();

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		if (!(await getCourse())) {
			loadingPage = false;
			invalidCourseId = true;
			return;
		}

		await getAssets(course.id);

		// Start polling if the scan status is either waiting or processing
		if (course && course.scanStatus && scanPoll === -1) {
			startPolling();
		}

		loadingPage = false;
	});

	onDestroy(() => {
		if (scanPoll !== -1) {
			clearInterval(scanPoll);
			scanPoll = -1;
		}
	});
</script>

<div class="flex w-full flex-col gap-4 pb-10">
	{#if loadingPage}
		<div
			class="flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center p-10"
		>
			<Loading class="border-primary" />
		</div>
	{:else if invalidCourseId}
		<Error />
	{:else}
		<div class="w-full border-b">
			<!-- Header -->
			<div class="container flex items-center py-4 md:py-6">
				<span class="grow text-base font-semibold md:text-lg">{course.title}</span>
				<div class="flex flex-row items-center gap-5">
					<!-- Open -->
					<a class="action hover:bg-accent-1 border" href="/course?id={course.id}"> Open </a>

					<!-- Scan -->
					<button
						disabled={course.scanStatus !== ''}
						class="action bg-success text-white enabled:hover:brightness-110 disabled:opacity-60"
						on:click={async () => {
							await AddScan({ courseId: course.id })
								.then(() => {
									course.scanStatus = 'waiting';
								})
								.catch((err) => {
									console.error(err);
								});
						}}
					>
						<Icons.search class="h-4 w-4" />
						Scan
					</button>

					<!-- Delete -->
					<button
						class="action bg-error text-white hover:brightness-110"
						on:click={() => {
							openDeleteDialog.set(true);
						}}
					>
						<Icons.delete class="h-4 w-4" />
						Delete
					</button>
				</div>
			</div>
		</div>

		<div class="container flex flex-col gap-4">
			<!-- Course Details -->
			<div class="flex flex-col gap-8 lg:flex-row">
				<!-- Card -->
				<CourseCard
					{course}
					class="order-1 flex h-48 w-full shrink-0 place-content-center lg:order-2 lg:w-[20rem]"
				/>

				<div class="order-2 flex w-full flex-col gap-5 lg:order-1">
					<!-- Path -->
					<div class="card">
						<div class="title">
							<Icons.path class="icon fill-foreground-muted" />
							<span>Path</span>
						</div>
						<span class="text-sm">{course.path}</span>
					</div>

					<div class="grid grid-cols-3 gap-2.5">
						<!-- Added -->
						<div class="card">
							<div class="title">
								<Icons.calendarPlus class="icon" />
								<span>Added</span>
							</div>
							<TableDate date={course.createdAt} class="text-foreground" />
						</div>

						<!-- Updated -->
						<div class="card">
							<div class="title">
								<Icons.calendarSearch class="icon" />
								<span>Updated</span>
							</div>
							<TableDate date={course.updatedAt} class="text-foreground" />
						</div>

						<!-- Scan status -->
						<div class="card">
							<div class="title">
								<Icons.search class="icon" />
								<span>Scan Status</span>
							</div>
							<span
								class={cn(
									'text-foreground',
									course.scanStatus && isProcessing && 'text-success animate-pulse',
									course.scanStatus && !isProcessing && 'text-foreground-muted animate-pulse'
								)}
							>
								{course.scanStatus ? (isProcessing ? 'scanning' : 'queued') : '-'}
							</span>
						</div>
					</div>
				</div>
			</div>

			<!--  Course Assets -->
			<div class="card !border-0 !px-0">
				<div class="title">
					<Icons.files class="icon" />
					<span>Course Assets</span>
				</div>

				<div class="flex flex-col rounded-md border">
					{#if Object.keys(chapters).length === 0}
						<div class="flex flex-col items-center justify-center py-6">
							<span class="text-foreground-muted">No assets found</span>
						</div>
					{:else}
						{#each Object.keys(chapters) as chapter, i}
							{@const lastChapter = Object.keys(chapters).length - 1 == i}

							<div
								{...$item(chapter)}
								use:item
								class={cn(
									'flex flex-col',
									!lastChapter && 'border-b',
									$isSelected(chapter) && 'bg-accent-1'
								)}
							>
								<!-- Chapter title (button) -->
								<button
									{...$trigger(chapter)}
									use:trigger
									class="hover:bg-accent-1 flex w-full items-center px-4 py-5"
								>
									<span class="grow text-start text-base">{chapter}</span>
									<span class="text-foreground-muted shrink-0 px-2.5"
										>({chapters[chapter].length})</span
									>
									<Icons.chevronRight
										class={cn('h-4 w-4 duration-200', $isSelected(chapter) && 'rotate-90')}
									/>
								</button>

								{#if $isSelected(chapter)}
									<div
										{...$content(chapter)}
										use:content
										transition:slide
										class="bg-background border-t"
									>
										{#each chapters[chapter] as asset, i}
											{@const lastAsset = chapters[chapter].length - 1 == i}
											<div
												class={cn(
													'full-border group relative flex flex-col px-5 text-sm md:px-8',
													lastAsset && 'hover:after:!-bottom-px'
												)}
											>
												<div class={cn('flex py-4 ', !lastAsset && 'border-b')}>
													<CourseAssetRow {asset} />
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>

		<!-- Delete Dialog -->
		<DeleteCourseDialog
			dialog={deleteDialog}
			bind:id={course.id}
			on:confirmed={() => {
				goto('/settings/courses');
			}}
		/>
	{/if}
</div>

<style lang="postcss">
	.action {
		@apply inline-flex items-center justify-center gap-2 whitespace-nowrap rounded px-3 py-1.5 text-center text-sm duration-200;
	}

	.card {
		@apply shrink-0 overflow-hidden rounded-md border px-4 py-3 text-sm;

		> .title {
			@apply text-foreground-muted flex select-none items-center gap-2.5 pb-2.5 text-xs tracking-wide;

			& > :global(.icon) {
				@apply h-4 w-4 stroke-[1.5];
			}
		}
	}

	.full-border {
		@apply hover:before:bg-border hover:before:absolute hover:before:-top-px hover:before:left-0 hover:before:z-10 hover:before:h-px hover:before:w-full;
		@apply hover:after:bg-border hover:after:absolute hover:after:bottom-0 hover:after:left-0 hover:after:z-10 hover:after:h-px hover:after:w-full;
	}
</style>
