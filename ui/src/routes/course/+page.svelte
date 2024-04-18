<script lang="ts">
	import { page } from '$app/stores';
	import { Error, Loading } from '$components';
	import { CourseContent, CourseMenu } from '$components/course';
	import { ErrorMessage, GetAllCourseAssets, UpdateAsset } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import { GetCourseFromParams, NO_CHAPTER, buildChapterStructure, isBrowser } from '$lib/utils';
	import { onMount } from 'svelte';

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

	// scrollY and innerWidth are bound to the window scroll and resize events
	let scrollY: number;
	let innerWidth: number;

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
		<div class="flex h-full flex-row">
			<CourseMenu title={course.title} id={course.id} {chapters} bind:selectedAsset />

			<CourseContent
				bind:selectedAsset
				bind:prevAsset
				bind:nextAsset
				on:update={() => {
					if (!selectedAsset) return;

					updateAsset(selectedAsset);
					chapters = { ...chapters };
				}}
			/>
		</div>
	{/if}
</div>
