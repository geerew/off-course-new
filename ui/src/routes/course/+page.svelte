<script lang="ts">
	import { page } from '$app/stores';
	import { Err, Loading } from '$components/generic';
	import { CourseContent, CourseMenu } from '$components/pages/course';
	import { GetAllCourseAssets, GetCourseFromParams, UpdateAsset } from '$lib/api';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import { NO_CHAPTER, buildChapterStructure } from '$lib/utils';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Variables
	// ----------------------
	// Hold the assets + attachments for this course
	let assets: Asset[];

	// Holds the course assets in a chapter structure. Populated when getCourse is called in
	// onMount
	let chapters: CourseChapters = {};

	// When an asset is selected, these will be populated
	let selectedAsset: Asset | null;
	let prevAsset: Asset | null;
	let nextAsset: Asset | null;

	// ----------------------
	// Functions
	// ----------------------

	// Lookup the course based upon the search params
	const getCourse = async () => {
		try {
			const course = await GetCourseFromParams($page.url.searchParams);
			if (!course) throw new Error('Course not found');

			// Get the assets
			assets = await GetAllCourseAssets(course.id, {
				orderBy: 'chapter asc, prefix asc',
				expand: true
			});
			if (!assets) throw new Error('Failed to get course assets');

			chapters = buildChapterStructure(assets);

			// Set ?a=xxx as the selected asset
			const assetId = $page.url.searchParams && $page.url.searchParams.get('a');
			if (assetId) selectedAsset = findAsset(assetId, chapters);

			// If there is no selected asset, set it as the first unfinished asset
			if (!selectedAsset) selectedAsset = findFirstUnfinishedAsset(course, chapters);

			return course;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
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

		try {
			await UpdateAsset(asset);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
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
</script>

<div class="flex w-full flex-col">
	{#await getCourse()}
		<Loading />
	{:then data}
		<div class="flex h-full flex-row">
			<CourseMenu title={data.title} id={data.id} {chapters} bind:selectedAsset />

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
	{:catch error}
		<Err errorMessage={error} />
	{/await}
</div>
