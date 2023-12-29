<script lang="ts">
	import { page } from '$app/stores';
	import { Error, Loading, Video } from '$components';
	import { Icons } from '$components/icons';
	import AttachmentsPopover from '$components/settings/internal/AttachmentsPopover.svelte';
	import { addToast } from '$lib/stores/addToast';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import { ErrorMessage, GetCourse, UpdateAsset, UpdateCourse } from '$lib/utils/api';
	import { NO_CHAPTER, buildChapterStructure, cn, isBrowser } from '$lib/utils/general';
	import { createAccordion } from '@melt-ui/svelte';
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

	// Gets the course id from the search params and queries the api for the course. It finds the
	// assets and attachments associated with this course and builds a chapter structure. It also
	// finds the first asset that is not finished, or if the course is completed, finds the first
	// asset and sets it as the selected asset
	const getCourse = async () => {
		if (!isBrowser) return false;

		const params = isBrowser && $page.url.searchParams;
		const courseId = params && params.get('id');
		if (!courseId) return false;

		return await GetCourse(courseId, { expand: true })
			.then((resp) => {
				if (!resp) return false;

				// build the assets chapter structure
				let tmpChapters: CourseChapters = {};
				if (resp.assets) tmpChapters = buildChapterStructure(resp.assets);

				// If an asset id is provided as a query param, lookup the asset and set it as
				// the selected
				const assetId = params && params.get('a_id');
				if (assetId) selectedAsset = findAsset(assetId, tmpChapters);

				// If there is still no selected asset, set it as the first unfinished asset
				if (!selectedAsset) selectedAsset = findFirstUnfinishedAsset(resp, tmpChapters);

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

				course = resp;
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

	// ----------------------

	// Finds the first asset that is not finished, or if the course is completed, find the first
	// asset
	const findFirstUnfinishedAsset = (course: Course, chapters: CourseChapters): Asset | null => {
		for (const chapterKey in chapters) {
			const chapterAssets = chapters[chapterKey];

			// Find the first asset in the current chapter that's not finished
			for (const asset of chapterAssets) {
				// When the course has been completed, return the first asset regardless of whether it
				// is finished or not
				if (course.finished) return asset;

				// If the asset is not finished, return it
				if (!asset.finished) return asset;
			}
		}

		// There are no assets within this course
		return null;
	};

	// ----------------------

	// It flips the current `finished`  state. So if the asset is finished, it will be set to
	// unfinished, and vice versa. It then determines if the course is completed or not and updates
	// the course accordingly
	const flipAssetFinishedState = async (assetId: string) => {
		if (!course.assets) return;

		// Get the asset
		const asset = course.assets.find((a) => a.id === assetId);
		if (!asset) return;

		// Flip the finished state
		asset.finished ? (asset.finished = false) : (asset.finished = true);

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

		// Count the number of assets that are finished for this course and update the course
		// started/finished accordingly
		const numAssets = course.assets.length;
		const finishedAssets = course.assets.filter((a) => a.finished).length;

		if (finishedAssets === numAssets) {
			// If all assets are finished, update the course to finished
			course.finished = true;
		} else if (finishedAssets === 0) {
			// If no assets are finished, update the course to not started
			course.finished = false;
			course.started = false;
		} else {
			// If some assets are finished, update the course to started
			course.finished = false;
			course.started = true;
		}

		// Update the course
		await updateCourse();
	};

	// ----------------------

	// When the asset is a video, this will be called as the video is played and the time
	// progresses. The data is stored in the backend and used when the asset is reloaded
	const updateAssetProgress = async (assetId: string, progress: number) => {
		if (!course.assets) return;

		// Get the asset
		const asset = course.assets.find((a) => a.id === assetId);
		if (!asset) return;

		// Set the progress
		asset.progress = progress;

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

	// ----------------------

	// Update the course. Called when information about the asset changes, such as the video
	// finishes
	const updateCourse = async () => {
		await UpdateCourse(course).catch((err) => {
			const errMsg = ErrorMessage(err);
			console.error(errMsg);
			$addToast({
				data: {
					message: errMsg,
					status: 'error'
				}
			});
		});
	};

	// ----------------------

	// Find an asset by id
	const findAsset = (id: string, chapters: CourseChapters): Asset | null => {
		const allAssets = Object.values(chapters).flat();
		const index = allAssets.findIndex((a) => a.id === id);
		if (index === -1) return null;
		return allAssets[index];
	};

	// ----------------------

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

		loadingPage = false;
	});
</script>

<div class="flex w-full flex-col">
	{#if loadingPage}
		<div
			class="flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center p-10"
		>
			<Loading class="border-primary" />
		</div>
	{:else if invalidCourseId}
		<Error />
	{:else}
		<div class="bg-background-muted w-full border-b">
			<div class="container flex h-[var(--course-header-height)] items-center">
				<span class="grow text-base font-semibold md:text-lg">{course.title}</span>
			</div>
		</div>

		<div class="container flex-1 items-start !pl-4 lg:grid lg:grid-cols-[300px_minmax(0,1fr)]">
			<!-- Chapters -->
			<div
				class="lg:sticky lg:left-0 lg:top-0 lg:h-[calc(100vh-var(--header-height)-var(--course-header-height)-1px)] lg:overflow-y-auto lg:border-r"
			>
				{#if Object.keys(chapters).length === 0}
					<div class="flex flex-col items-center justify-center py-6">
						<span class="text-foreground-muted">No assets found</span>
					</div>
				{:else}
					{#each Object.keys(chapters) as chapter, i}
						{@const [prefix, ...rest] = chapter.split(' ')}
						{@const title = rest.join(' ')}
						{@const finishedCount = chapters[chapter].filter((a) => a.finished).length}

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
									<Icons.chevronRight
										class={cn(
											'mt-0.5 h-4 w-4 shrink-0 duration-200',
											$isSelected(chapter) && 'rotate-90'
										)}
									/>
								</div>
								<div
									class={cn(
										'flex text-xs',
										finishedCount === chapters[chapter].length && 'text-success'
									)}
								>
									{finishedCount} of {chapters[chapter].length}
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
														flipAssetFinishedState(asset.id);
													}}
												>
													<input tabindex="0" type="checkbox" checked={asset.finished} />
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
																	? Icons.fileVideo
																	: asset.assetType === 'html'
																	? Icons.fileHtml
																	: Icons.filePdf}
																class="h-4 w-4"
															/>
															<span>{asset.assetType}</span>
														</span>

														<!-- Attachments -->
														{#if asset.attachments && asset.attachments.length > 0}
															<AttachmentsPopover
																attachments={asset.attachments}
																showIcon={false}
																showCount={false}
															/>
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
						startTime={selectedAsset.progress}
						prevVideo={prevAsset}
						nextVideo={nextAsset}
						on:started={() => {
							// Video started. Mark the course as started
							if (!selectedAsset || course.started) return;
							updateCourse();
						}}
						on:progress={(e) => {
							// Update the asset progress
							if (!selectedAsset) return;
							updateAssetProgress(selectedAsset.id, e.detail);
						}}
						on:finished={() => {
							// Video finished. Mark the asset as finished
							if (!selectedAsset || selectedAsset.finished) return;
							flipAssetFinishedState(selectedAsset.id);
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
		@apply indeterminate:bg-foreground-muted/50 indeterminate:border-transparent;
		@apply outline-none focus:ring-0 focus:ring-offset-0;
	}
</style>
