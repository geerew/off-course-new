<script lang="ts">
	import { page } from '$app/stores';
	import { Err, Loading } from '$components/generic';
	import { CourseContent, CourseMenu } from '$components/pages/course';
	import { GetAllCourseAssets, GetCourseFromParams, UpdateAsset } from '$lib/api';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import { BuildChapterStructure, IsBrowser, UpdateQueryParam } from '$lib/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Variables
	// ----------------------

	// Used during the #await. It is initially set to a promise that never resolves to prevent
	// the page from rendering before the course is fetched, which occurs during onMount. This
	// is because the site is pre-rendered and as such we can only get the search params after
	// the page is mounted
	let coursePromise: Promise<Course> = new Promise(() => {});

	let pageParams: URLSearchParams;

	// Hold the assets + attachments for this course
	let assets: Asset[];

	// Holds the course assets in a chapter structure. Populated when getCourse is called in
	// onMount
	let chapters: CourseChapters = {};

	// When an asset is selected, these will be populated
	let selectedAsset: Asset | null = null;
	let prevAsset: Asset | null = null;
	let nextAsset: Asset | null = null;

	// ----------------------
	// Functions
	// ----------------------

	// Lookup the course based on the `id` query param. Then build a chapter structure from the assets
	// and attachment
	//
	// If the `a` query param is not set, find the first unfinished asset and set the `a` query param,
	// triggering a reactive statement, which sets the selected asset
	async function getCourse(): Promise<Course> {
		if (!IsBrowser) return {} as Course;

		try {
			const course = await GetCourseFromParams(pageParams);
			if (!course) throw new Error('Course not found');

			// Get the assets
			assets = await GetAllCourseAssets(course.id, {
				orderBy: 'chapter asc, prefix asc',
				expand: true
			});
			if (!assets) throw new Error('Failed to get course assets');

			chapters = BuildChapterStructure(assets);

			// If no asset was found in the query params, find the first unfinished asset and
			// update the query params. This will trigger the reactive statement that will find
			// and set the selected asset
			if (!pageParams || !pageParams.get('a')) {
				const found = findFirstUnfinishedAsset(course, chapters);
				if (found) {
					UpdateQueryParam('a', found.id, true);
				}
			}

			return course;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As the query param `a` changes, this will be reactively called to
	function updateSelectedAsset(id: string) {
		selectedAsset = findAsset(id, chapters);
		if (!selectedAsset) return;

		// Set the previous and next assets
		const { prev, next } = findAdjacentAssets(selectedAsset, chapters);
		prevAsset = prev;
		nextAsset = next;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Finds the first asset that is not finished, or if the course is completed, find the first
	// asset
	function findFirstUnfinishedAsset(course: Course, chapters: CourseChapters): Asset | null {
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
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the asset is a video, this will be called as the video is played and the time
	// progresses. The data is stored in the backend and used when the asset is reloaded
	async function updateAsset(asset: Asset) {
		if (!assets) return;

		try {
			await UpdateAsset(asset);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find an asset by id
	function findAsset(id: string, chapters: CourseChapters): Asset | null {
		const allAssets = Object.values(chapters).flat();
		const index = allAssets.findIndex((a) => a.id === id);
		if (index === -1) return null;
		return allAssets[index];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Find the previous and next assets within chapters based upon a current asset. This is used
	// when a video is finished to determine which asset to play next. It is also passed to the
	// video player to generate the previous and next buttons
	function findAdjacentAssets(
		currentAsset: Asset,
		chapters: CourseChapters
	): { prev: Asset | null; next: Asset | null } {
		// Flatten the assets and find the index of the current asset
		const allAssets = Object.values(chapters).flat();
		const index = allAssets.findIndex((a) => a.id === currentAsset.id);

		// Determine the previous and next assets based on the index
		const prevAsset = index > 0 ? allAssets[index - 1] : null;
		const nextAsset = index < allAssets.length - 1 ? allAssets[index + 1] : null;

		return { prev: prevAsset, next: nextAsset };
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When the query param `a` changes, update the selected asset
	$: {
		if (IsBrowser) {
			const assetId = $page.url.searchParams.get('a');
			if (assetId && chapters && selectedAsset?.id !== assetId) {
				updateSelectedAsset(assetId);
			}
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	// Due to the site being pre-rendered, we need to wait for the page to be mounted before we
	// can get the search params
	onMount(() => {
		pageParams = $page.url.searchParams;
		coursePromise = getCourse();
	});
</script>

<div class="flex w-full flex-col">
	{#await coursePromise}
		<Loading />
	{:then data}
		<div class="flex h-full flex-col lg:container lg:flex-row lg:gap-10">
			<CourseMenu title={data.title} id={data.id} {chapters} {selectedAsset} />

			<CourseContent
				{selectedAsset}
				{prevAsset}
				{nextAsset}
				on:update={() => {
					if (!selectedAsset) return;
					updateAsset(selectedAsset);
					chapters = { ...chapters };
				}}
			/>
		</div>
	{:catch error}
		<Err errorMessage={error} />
	{/await}
</div>
