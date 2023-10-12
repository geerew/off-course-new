<script lang="ts">
	import { page } from '$app/stores';
	import { Error, Loading } from '$components';
	import type { Course } from '$lib/types/models';
	import { GetCourse } from '$lib/utils/api';
	import { isBrowser } from '$lib/utils/general';
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

	// ----------------------
	// Functions
	// ----------------------

	// Gets the course id from the search params and queries the api for the course
	const getCourse = async () => {
		if (!isBrowser) return false;

		const params = isBrowser && $page.url.searchParams;
		const id = params && params.get('id');
		if (!id) return false;

		return await GetCourse(id, { includeAssets: true }, true)
			.then((resp) => {
				if (!resp) return false;

				course = resp;
				return true;
			})
			.catch((err) => {
				console.error(err);
				return false;
			});
	};

	onMount(async () => {
		if (!(await getCourse())) {
			loadingPage = false;
			invalidCourseId = true;
			return;
		}

		loadingPage = false;
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
		<div class="bg-background-muted w-full border-b">
			<div class="container flex items-center py-4 md:py-6">
				<span class="grow text-base font-semibold md:text-lg">{course.title}</span>
			</div>
		</div>
	{/if}
</div>
