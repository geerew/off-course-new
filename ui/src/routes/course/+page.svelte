<script lang="ts">
	import { page } from '$app/stores';
	import { Error, Loading } from '$components';
	import { Badge } from '$components/ui/badge';
	import Button from '$components/ui/button/button.svelte';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { Video } from '$components/video';
	import { ATTACHMENT_API, ErrorMessage, GetAllCourseAssets, UpdateAsset } from '$lib/api';
	import * as Accordion from '$lib/components/ui/accordion';
	import { addToast } from '$lib/stores/addToast';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import {
		GetCourseFromParams,
		NO_CHAPTER,
		buildChapterStructure,
		cn,
		isBrowser
	} from '$lib/utils';
	import { CheckCircle2, ChevronRight, CircleDotDashed, Dot, Download } from 'lucide-svelte';
	import { onMount, tick } from 'svelte';

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
	let selectedChapter: string[] = [];
	let selectedAsset: Asset | null;
	let prevAsset: Asset | null;
	let nextAsset: Asset | null;

	// The title element
	let titleEl: HTMLDivElement;

	// scrollY and innerWidth are bound to the window scroll and resize events
	let scrollY: number;
	let innerWidth: number;

	// The menu offset is used to determine the top position of the menu when the user scrolls
	let topOffset = 0;
	let menuPx = topOffset;

	// ----------------------
	// Functions
	// ----------------------

	// Lookup the course based upon the search params
	const getCourse = async () => {
		if (!isBrowser) return false;

		return await GetCourseFromParams($page.url.searchParams)
			.then((resp) => {
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

	// Gets the assets + attachments for the given course. It will build a chapter structure and
	// set the selected asset based upon the query params. If there is no asset query param, it
	// will set the first unfinished asset as the selected asset
	const getAssets = async (courseId: string) => {
		if (!isBrowser) return false;

		const params = $page.url.searchParams;

		return await GetAllCourseAssets(courseId, { orderBy: 'chapter asc, prefix asc' })
			.then(async (resp) => {
				if (!resp) return false;

				chapters = buildChapterStructure(resp);

				// Set ?a=xxx as the selected asset
				const assetId = params && params.get('a');
				if (assetId) selectedAsset = findAsset(assetId, chapters);

				// If there is no selected asset, set it as the first unfinished asset
				if (!selectedAsset) selectedAsset = findFirstUnfinishedAsset(course, chapters);

				assets = resp;
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

	// When the asset is a video, this will be called as the video is played and the time
	// progresses. The data is stored in the backend and used when the asset is reloaded
	async function updateAsset(asset: Asset) {
		if (!assets) return;

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
	}

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
		url.searchParams.set('a', assetId);
		history.pushState({}, '', url.toString());
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Calculate the menu offset based on the scroll position
	function calculateMenuOffset() {
		if (scrollY > topOffset) {
			menuPx = 0;
			return;
		}

		menuPx = topOffset - scrollY;
	}

	function scrollToSelected() {
		const selectedElement = document.querySelector('[data-selected="true"]');
		if (selectedElement) {
			selectedElement.scrollIntoView({
				behavior: 'smooth',
				block: 'start',
				inline: 'nearest'
			});
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Called when the selected asset changes. It sets the  previous and next assets and updates
	function updateSelectedAsset() {
		if (!selectedAsset) return;

		const assetChapter = !selectedAsset.chapter ? NO_CHAPTER : selectedAsset.chapter;

		// Set ?a=xxx
		updateQueryParam(selectedAsset.id);

		// Set the previous and next assets
		const { prev, next } = findAdjacentAssets(selectedAsset, chapters);
		prevAsset = prev;
		nextAsset = next;

		selectedChapter = [assetChapter];
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When the window is resized, recalculate the menu offset
	$: if (titleEl && innerWidth > 0) {
		topOffset = 64 + titleEl.getBoundingClientRect().height;
		calculateMenuOffset();
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As the user scrolls, recalculate the menu offset
	$: if (scrollY > 0) {
		calculateMenuOffset();
	}

	// When the selected asset changes, call updateSelectedAsset()
	$: {
		if (selectedAsset) {
			updateSelectedAsset();
		}
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

		if (selectedAsset) {
			await tick();
			scrollToSelected();
		}
	});
</script>

<svelte:window bind:scrollY bind:innerWidth />

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
		<!-- Heading -->
		<div class="w-full border-b" bind:this={titleEl}>
			<div class="container flex items-center gap-2.5 py-2 md:py-4">
				<span class="grow text-base font-semibold md:text-lg">{course.title}</span>
				<Button
					variant="outline"
					href="/settings/courses/details?id={course.id}"
					class="h-8 md:h-10"
				>
					Details
				</Button>
			</div>
		</div>

		<div class="container flex flex-1 flex-row gap-2.5 pt-2">
			<div
				class="bg-muted/70 border-muted/70 sticky left-8 top-2 z-10 hidden w-72 shrink-0 overflow-hidden rounded-lg border lg:flex"
				style="max-height: calc(100vh - {menuPx}px - 18px)"
			>
				<div class="flex flex-grow flex-col justify-between overflow-y-auto overflow-x-hidden">
					<Accordion.Root bind:value={selectedChapter} multiple class="w-full rounded-lg">
						{#each Object.keys(chapters) as chapter, i}
							{@const numAssets = chapters[chapter].length}
							{@const completedAssets = chapters[chapter].filter((a) => a.completed).length}
							{@const lastChapter = Object.keys(chapters).length - 1 == i}

							<Accordion.Item
								value={chapter}
								class={cn('border-background', lastChapter && 'border-transparent')}
							>
								<Accordion.Trigger class="hover:bg-muted px-3 py-4 text-start hover:no-underline">
									<div class="flex w-full flex-col gap-2.5 pr-4">
										<span class="text-sm font-semibold">{chapter}</span>
										<div>
											<Badge
												class={cn(
													'bg-muted text-success-foreground hover:bg-success items-center rounded-sm px-1',
													numAssets === completedAssets && 'bg-success'
												)}
											>
												{completedAssets}/{numAssets}
											</Badge>
										</div>
									</div>
								</Accordion.Trigger>

								<Accordion.Content class="bg-background flex flex-col">
									{#each chapters[chapter] as asset, i}
										{@const lastAsset = chapters[chapter].length - 1 == i}
										<div class={cn(!lastAsset && 'border-muted/70 border-b')}>
											<Button
												variant="ghost"
												class={cn(
													'flex h-auto w-full flex-row gap-2.5 whitespace-normal rounded-none px-3 py-4 text-start',
													asset.id === selectedAsset?.id && 'bg-muted bg-opacity-40'
												)}
												data-selected={asset.id === selectedAsset?.id ? 'true' : 'false'}
												on:click={() => {
													selectedAsset = asset;
												}}
											>
												<div class="flex grow flex-col gap-2.5">
													<span class={cn(asset.id === selectedAsset?.id && 'text-primary')}
														>{asset.prefix}. {asset.title}</span
													>
													<div
														class="text-muted-foreground flex select-none flex-row flex-wrap items-center gap-y-2 text-xs"
													>
														<!-- Type -->
														<span>{asset.assetType}</span>
														{#if asset.attachments && asset.attachments.length > 0}
															<Dot class="h-5 w-5" />

															<DropdownMenu.Root closeOnItemClick={false}>
																<DropdownMenu.Trigger asChild let:builder>
																	<Button
																		builders={[builder]}
																		variant="ghost"
																		class="group relative flex h-auto items-center gap-1 px-0 py-0 text-xs hover:bg-transparent"
																		on:click={(e) => {
																			e.stopPropagation();
																		}}
																	>
																		attachments

																		<ChevronRight
																			class="h-3 w-3 duration-200 group-data-[state=open]:rotate-90"
																		/>
																	</Button>
																</DropdownMenu.Trigger>

																<DropdownMenu.Content
																	class="flex max-h-[10rem] w-auto max-w-xs flex-col overflow-y-scroll md:max-w-sm"
																	fitViewport={true}
																>
																	{#each asset.attachments as attachment, i}
																		{@const lastAttachment = asset.attachments.length - 1 == i}
																		<DropdownMenu.Item
																			class="cursor-pointer justify-between gap-3 text-xs"
																			href={ATTACHMENT_API + '/' + attachment.id + '/serve'}
																			download
																		>
																			<div class="flex flex-row gap-1.5">
																				<span class="shrink-0">{i + 1}.</span>
																				<span class="grow">{attachment.title}</span>
																			</div>

																			<Download class="flex h-3 w-3 shrink-0" />
																		</DropdownMenu.Item>

																		{#if !lastAttachment}
																			<DropdownMenu.Separator
																				class="bg-muted my-1 -ml-1 -mr-1 block h-px"
																			/>
																		{/if}
																	{/each}
																</DropdownMenu.Content>
															</DropdownMenu.Root>
														{/if}
													</div>
												</div>

												<div class="mt-0.5 flex h-full w-4">
													{#if asset.completed}
														<CheckCircle2
															class="stroke-success [&>:nth-child(2)]:stroke-success-foreground fill-success h-4 w-4"
														/>
													{:else if asset.assetType === 'video' && asset.videoPos > 0}
														<CircleDotDashed class="stroke-secondary fill-secondary h-4 w-4" />
													{/if}
												</div>
											</Button>
										</div>
									{/each}
								</Accordion.Content>
							</Accordion.Item>
						{/each}
					</Accordion.Root>
				</div>
			</div>

			<div class="flex h-full w-full flex-col">
				{#if selectedAsset && selectedAsset.assetType === 'video'}
					<Video
						title={selectedAsset.title}
						src={selectedAsset.id}
						startTime={selectedAsset.videoPos}
						{nextAsset}
						on:progress={(e) => {
							if (!selectedAsset) return;
							selectedAsset.videoPos = e.detail;
							updateAsset(selectedAsset);
						}}
						on:finished={(e) => {
							if (!selectedAsset) return;

							selectedAsset.videoPos = e.detail;
							selectedAsset.completed = true;
							updateAsset(selectedAsset);

							// This is needed to update the n/n in the chapter title
							chapters = { ...chapters };
						}}
						on:next={(e) => {
							if (!nextAsset) return;
							selectedAsset = nextAsset;
						}}
					/>
				{/if}
			</div>
		</div>
	{/if}
</div>
