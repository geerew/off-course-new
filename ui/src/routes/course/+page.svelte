<script lang="ts">
	import { page } from '$app/stores';
	import { Error, Loading, Video } from '$components';
	import { ErrorMessage, GetAllCourseAssets, GetCourse, UpdateAsset } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import { NO_CHAPTER, buildChapterStructure, cn, isBrowser } from '$lib/utils';
	import { createAccordion } from '@melt-ui/svelte';
	import { ChevronRight, FileCode, FileText, FileVideo, Info } from 'lucide-svelte';
	import { onMount } from 'svelte';
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

	// Hold the assets + attachments for this course
	let assets: Asset[];

	// Holds the course assets in a chapter structure. Populated when getCourse is called in
	// onMount
	let chapters: CourseChapters = {};

	// When an asset is selected, these will be populated
	let selectedAsset: Asset | null;
	let prevAsset: Asset | null;
	let nextAsset: Asset | null;

	// Accordion to show the course chapters
	const {
		elements: { content, item, trigger },
		helpers: { isSelected },
		states: { value }
	} = createAccordion({
		multiple: true
	});

	// ----------------------
	// Functions
	// ----------------------

	// Gets the course id from the search params and queries the api for the course
	const getCourse = async () => {
		if (!isBrowser) return false;

		const params = $page.url.searchParams;
		const courseId = params && params.get('id');
		if (!courseId) return false;

		return await GetCourse(courseId)
			.then(async (resp) => {
				if (!resp) return false;
				course = resp;
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

		const params = $page.url.searchParams;

		return await GetAllCourseAssets(courseId)
			.then(async (resp) => {
				if (!resp) return false;

				// Build the assets chapter structure
				let tmpChapters: CourseChapters = {};
				if (resp) tmpChapters = buildChapterStructure(resp);

				// If an asset id is provided as a query param, lookup the asset and set it as
				// the selected
				const assetId = params && params.get('a_id');
				if (assetId) selectedAsset = findAsset(assetId, tmpChapters);

				// If there is still no selected asset, set it as the first unfinished asset
				if (!selectedAsset) selectedAsset = findFirstUnfinishedAsset(course, tmpChapters);

				// Set the chapter, update the query params for the selected asset and find the
				// previous/next assets
				//
				// Note: At this point, the selected asset will be null if this course does not
				// contain any assets
				if (selectedAsset) {
					!selectedAsset.chapter ? value.set([NO_CHAPTER]) : value.set([selectedAsset.chapter]);

					// Update the query param. This is used in the event the user refreshes the
					// page. They will reload the same asset
					updateQueryParam(selectedAsset.id);

					// Set the previous and next assets
					const { prev, next } = findAdjacentAssets(selectedAsset, tmpChapters);
					prevAsset = prev;
					nextAsset = next;
				}

				assets = resp;
				chapters = tmpChapters;
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

	// Finds the first asset that is not finished, or if the course is completed, find the first
	// asset
	const findFirstUnfinishedAsset = (course: Course, chapters: CourseChapters): Asset | null => {
		for (const chapterKey in chapters) {
			const chapterAssets = chapters[chapterKey];

			// Find the first asset in the current chapter that's not completed
			for (const asset of chapterAssets) {
				// When the course has been completed, return the first asset regardless of whether it
				// is finished or not
				if (course.percent === 100) return asset;

				// If the asset is not completed, return it
				if (!asset.completed) return asset;
			}
		}

		// There are no assets within this course
		return null;
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// It flips the current `completed` state. So if the asset is completed, it will be marked as
	// not completed, and vice versa. It then determines if the course is completed or not and
	// updates the course accordingly
	const flipAssetCompleted = async (assetId: string) => {
		if (!assets) return;

		// Get the asset
		const asset = assets.find((a) => a.id === assetId);
		if (!asset) return;

		// Flip
		asset.completed ? (asset.completed = false) : (asset.completed = true);

		// Update the finished state for the asset
		await UpdateAsset(asset).catch((err) => {
			const errMsg = ErrorMessage(err);
			console.error(errMsg);
			$addToast({
				data: {
					message: errMsg,
					status: 'error'
				}
			});

			return;
		});

		// Update the chapters (for reactivity)
		chapters = { ...chapters };
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the asset is a video, this will be called as the video is played and the time
	// progresses. The data is stored in the backend and used when the asset is reloaded
	const updateAssetVideoPos = async (assetId: string, position: number) => {
		if (!assets) return;

		// Get the asset
		const asset = assets.find((a) => a.id === assetId);
		if (!asset || asset.assetType !== 'video') return;

		// Set the progress
		asset.videoPos = position;

		// Update the asset
		await UpdateAsset(asset).catch((err) => {
			const errMsg = ErrorMessage(err);
			console.error(errMsg);
			$addToast({
				data: {
					message: errMsg,
					status: 'error'
				}
			});

			return;
		});
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find an asset by id
	const findAsset = (id: string, chapters: CourseChapters): Asset | null => {
		const allAssets = Object.values(chapters).flat();
		const index = allAssets.findIndex((a) => a.id === id);
		if (index === -1) return null;
		return allAssets[index];
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find the previous and next assets within chapters based upon a current asset. This is used
	// when a video is finished to determine which asset to play next. It is also passed to the
	// video player to generate the previous and next buttons
	const findAdjacentAssets = (
		currentAsset: Asset,
		chapters: CourseChapters
	): { prev: Asset | null; next: Asset | null } => {
		// Flatten the assets and find the index of the current asset
		const allAssets = Object.values(chapters).flat();
		const index = allAssets.findIndex((a) => a.id === currentAsset.id);

		// Determine the previous and next assets based on the index
		const prevAsset = index > 0 ? allAssets[index - 1] : null;
		const nextAsset = index < allAssets.length - 1 ? allAssets[index + 1] : null;

		return { prev: prevAsset, next: nextAsset };
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function updateQueryParam(assetId: string) {
		if (typeof window === 'undefined') return;

		const url = new URL(window.location.href);
		url.searchParams.set('a_id', assetId);
		history.pushState({}, '', url.toString());
	}

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		if (!(await getCourse())) {
			loadingPage = false;
			invalidCourseId = true;
			return;
		}

		if (!(await getAssets(course.id))) {
			loadingPage = false;
			return;
		}

		loadingPage = false;
	});
</script>

<div class="flex w-full flex-col">
	{#if loadingPage}
		<div
			class="flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center p-10"
		>
			<Loading />
		</div>
	{:else if invalidCourseId || assets.length === 0}
		<Error />
	{:else}
		<div class="w-full border-b">
			<div class="container flex h-[var(--course-header-height)] items-center">
				<span class="grow text-base font-semibold md:text-lg">{course.title}</span>
				<a
					href="/settings/courses/details?id={course.id}"
					class="hover:bg-accent-1 inline-flex items-center justify-center gap-2 whitespace-nowrap rounded border px-3 py-1.5 text-center text-sm duration-200"
				>
					<Info class="h-4 w-4" />
					<span>Details</span>
				</a>
			</div>
		</div>

		<div
			class="container flex-1 items-start px-4 md:px-8 lg:grid lg:grid-cols-[300px_minmax(0,1fr)]"
		>
			<!-- Chapters -->
			<div
				class="lg:sticky lg:left-0 lg:top-0 lg:h-[calc(100vh-var(--header-height)-var(--course-header-height)-1px)] lg:overflow-y-auto lg:border-r"
			>
				{#if Object.keys(chapters).length === 0}
					<div class="flex flex-col items-center justify-center py-6">
						<span class="text-muted-foreground">No assets found</span>
					</div>
				{:else}
					{#each Object.keys(chapters) as chapter, i}
						{@const [prefix, ...rest] = chapter.split(' ')}
						{@const title = rest.join(' ')}
						{@const completedCount = chapters[chapter].filter((a) => a.completed).length}

						<div
							{...$item(chapter)}
							use:item
							class={cn('flex flex-col border-b', $isSelected(chapter) && 'bg-accent-1')}
						>
							<!-- Chapter title (button) -->
							<button
								{...$trigger(chapter)}
								use:trigger
								class="hover:bg-accent-1 flex w-full flex-col gap-2 px-4 py-5"
							>
								<div class="flex w-full flex-row gap-1.5">
									<span class="shink-0 text-start text-sm font-semibold">{prefix}</span>
									<span class="grow text-start text-sm font-semibold">{title}</span>
									<ChevronRight
										class={cn(
											'mt-0.5 h-4 w-4 shrink-0 duration-200',
											$isSelected(chapter) && 'rotate-90'
										)}
									/>
								</div>
								<div
									class={cn(
										'flex text-xs',
										completedCount === chapters[chapter].length && 'text-success'
									)}
								>
									{completedCount} of {chapters[chapter].length}
									completed
								</div>
							</button>

							{#if $isSelected(chapter)}
								<div
									{...$content(chapter)}
									use:content
									transition:slide
									class=" bg-background border-t"
								>
									{#each chapters[chapter] as asset, i}
										{@const lastAsset = chapters[chapter].length - 1 == i}

										<div
											class={cn(
												'hover:bg-accent-1/70 hover-full-border relative flex flex-col px-4 text-sm',
												lastAsset && 'after:!-bottom-px hover:after:!-bottom-px',
												selectedAsset && selectedAsset.id == asset.id
													? 'bg-accent-1/70 full-border'
													: ''
											)}
										>
											<div
												class={cn(
													'flex gap-3 py-2.5 text-start text-sm ',
													!lastAsset && 'border-b'
												)}
												on:click={() => {
													// Update the selected asset and find the previous/next
													if (selectedAsset === asset) return;
													selectedAsset = asset;
													const { prev, next } = findAdjacentAssets(selectedAsset, chapters);
													prevAsset = prev;
													nextAsset = next;

													// Update the query param. This is used in the
													// event the user refreshes the page
													updateQueryParam(selectedAsset.id);
												}}
												on:keydown={(e) => {
													if (e.key !== 'Enter') return;

													// Update the selected asset and find the previous/next
													if (selectedAsset === asset) return;
													selectedAsset = asset;
													const { prev, next } = findAdjacentAssets(selectedAsset, chapters);
													prevAsset = prev;
													nextAsset = next;

													// Update the query param. This is used in the
													// event the user refreshes the page
													updateQueryParam(selectedAsset.id);
												}}
												tabindex="0"
												role="button"
											>
												<!-- Finished/unfinished input -->
												<button
													class="flex"
													on:click|stopPropagation={() => {
														flipAssetCompleted(asset.id);
													}}
												>
													<input tabindex="0" type="checkbox" checked={asset.completed} />
												</button>

												<div class="flex select-none flex-col gap-3">
													<!-- Prefix + Title -->
													<span>{asset.prefix} {asset.title}</span>

													<div class="flex items-center gap-2">
														<span
															class="inline-flex select-none items-center justify-center gap-2 whitespace-nowrap rounded border px-2 py-1 text-center text-xs"
														>
															<!-- Asset type -->
															<svelte:component
																this={asset.assetType === 'video'
																	? FileVideo
																	: asset.assetType === 'html'
																		? FileCode
																		: FileText}
																class="h-4 w-4"
															/>
															<span>{asset.assetType}</span>
														</span>

														<!-- Attachments -->
														{#if asset.attachments && asset.attachments.length > 0}
															<!-- <AttachmentsPopover
																attachments={asset.attachments}
																showIcon={false}
																showCount={false}
															/> -->
														{/if}
													</div>
												</div>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				{/if}
			</div>

			<!-- Media -->
			<div class="flex h-full">
				{#if selectedAsset && selectedAsset.assetType === 'video'}
					<Video
						id={selectedAsset.id}
						startTime={selectedAsset.videoPos}
						prevVideo={prevAsset}
						nextVideo={nextAsset}
						on:progress={(e) => {
							if (!selectedAsset) return;

							// Set the course as started manually (This will automatically happen in the backend)
							if (e.detail > 5 && !course.started) {
								course.started = true;
							}

							updateAssetVideoPos(selectedAsset.id, e.detail);
						}}
						on:finished={() => {
							// Video finished. Mark the asset as completed but only if it is not
							// already completed
							if (!selectedAsset || selectedAsset.completed) return;
							flipAssetCompleted(selectedAsset.id);
						}}
						on:previous={() => {
							if (!prevAsset) return;
							selectedAsset = prevAsset;

							// Set the previous and next assets
							const { prev, next } = findAdjacentAssets(selectedAsset, chapters);
							prevAsset = prev;
							nextAsset = next;

							// Update the query param. This is used in the event the user
							// refreshes the page
							updateQueryParam(selectedAsset.id);
						}}
						on:next={() => {
							if (!nextAsset) return;
							selectedAsset = nextAsset;

							// Set the previous and next assets
							const { prev, next } = findAdjacentAssets(selectedAsset, chapters);
							prevAsset = prev;
							nextAsset = next;

							// Update the query param. This is used in the event the user
							// refreshes the page
							updateQueryParam(selectedAsset.id);
						}}
					/>
				{:else}
					TODO: handle type
				{/if}
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	.full-border {
		@apply before:bg-border before:absolute before:-top-px before:left-0 before:z-10 before:h-px before:w-full;
		@apply after:bg-border after:absolute after:bottom-0 after:left-0 after:z-10 after:h-px after:w-full;
	}

	.hover-full-border {
		@apply hover:before:bg-border hover:before:absolute hover:before:-top-px hover:before:left-0 hover:before:z-10 hover:before:h-px hover:before:w-full;
		@apply hover:after:bg-border hover:after:absolute hover:after:bottom-0 hover:after:left-0 hover:after:z-10 hover:after:h-px hover:after:w-full;
	}

	input {
		@apply bg-background pointer-events-none cursor-pointer rounded border-2 p-2 duration-150;
		@apply border-foreground/40;
		@apply checked:bg-primary checked:hover:bg-primary checked:border-transparent;
		@apply indeterminate:bg-muted/50 indeterminate:border-transparent;
		@apply outline-none focus:ring-0 focus:ring-offset-0;
	}
</style>
