<script lang="ts">
	import { CourseCard, Error, Loading } from '$components';
	import { addToast } from '$lib/stores/addToast';
	import type { Course } from '$lib/types/models';
	import { ErrorMessage, GetCourses } from '$lib/utils/api';
	import { isBrowser } from '$lib/utils/general';
	import { onMount } from 'svelte';
	import { TableDate } from './table';

	// ----------------------
	// Variables
	// ----------------------

	// True while the page is loading
	let loadingLatestCourses = true;

	// True when an error occurred
	let loadingLatestCoursesError = false;

	let latestCourses: Course[] = [];

	// The current page to get
	let currentPage = 1;

	let moreToGet = false;

	// ----------------------
	// Functions
	// ----------------------

	// Gets  (started) courses
	async function getLatestCourses(page: number) {
		if (!isBrowser) return false;

		console.log('loading page ', page);
		return await GetCourses({ page, perPage: 8 })
			.then((resp) => {
				if (!resp) return false;

				// If the current page is 1, then we can just set the courses to the response, or
				// else append the response to the current courses
				latestCourses.length === 0
					? (latestCourses = resp.items)
					: (latestCourses = [...latestCourses, ...resp.items]);

				// Are there more courses to get?
				moreToGet = latestCourses.length < resp.totalItems;

				loadingLatestCourses = false;
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
	}

	async function loadMoreCourses() {
		if (!moreToGet) return;
		currentPage++;
		if (!(await getLatestCourses(currentPage))) {
			loadingLatestCoursesError = true;
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		if (!(await getLatestCourses(currentPage))) {
			loadingLatestCourses = false;
			loadingLatestCoursesError = true;
		}
	});
</script>

<div class="flex flex-col gap-3 lg:gap-5">
	<h2 class="text-lg font-bold dark:text-white md:text-xl md:leading-tight">Latest Courses</h2>
	{#if loadingLatestCourses}
		<div class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10">
			<Loading class="border-primary" />
		</div>
	{:else if loadingLatestCoursesError}
		<Error class="text-muted min-h-[6rem] p-5 text-sm" imgClass="h-6 w-6" />
	{:else if latestCourses.length === 0}
		<div class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10">
			<span class="text-foreground-muted">No courses have been added.</span>
		</div>
	{:else}
		<div class="grid gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
			{#each latestCourses as course}
				<a
					class="group flex flex-col rounded-xl border shadow-sm transition hover:shadow-md"
					href="/course?id={course.id}"
				>
					<CourseCard
						{course}
						class="aspect-w-16 aspect-h-9"
						imgClass="rounded-none w-full rounded-t-xl object-cover"
						fallbackClass="w-auto h-auto rounded-none rounded-t-xl"
					/>

					<div class="flex h-full flex-col justify-between p-4 md:p-5">
						<h3
							class="text-muted group-hover:text-primary mt-2 font-semibold dark:group-hover:text-white"
						>
							{course.title}
						</h3>

						<TableDate date={course.createdAt} class="text-muted pt-3 text-xs" />
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>

{#if moreToGet}
	<button
		class="border-border text-foreground hover:bg-accent-1 inline-flex items-center justify-center rounded-lg border px-4 py-3 text-sm font-medium"
		on:click={loadMoreCourses}>Load More</button
	>
{/if}
