<script lang="ts">
	import { CourseCard, Error, Loading } from '$components';
	import { ErrorMessage, GetAllCourses } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { Course } from '$lib/types/models';
	import { isBrowser } from '$lib/utils';
	import { onMount } from 'svelte';

	// ----------------------
	// Variables
	// ----------------------

	// True while the page is loading
	let loadingOngoingCourses = true;
	let loadingOngoingCoursesError = false;

	let ongoingCourses: Course[] = [];

	// ----------------------
	// Functions
	// ----------------------

	// Gets all the ongoing (started) courses
	const getOngoingCourses = async () => {
		if (!isBrowser) return false;

		return await GetAllCourses({ started: true })
			.then((resp) => {
				if (!resp) return false;
				ongoingCourses = resp;
				loadingOngoingCourses = false;
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
	// Lifecycle
	// ----------------------

	onMount(async () => {
		if (!(await getOngoingCourses())) {
			loadingOngoingCourses = false;
			loadingOngoingCoursesError = true;
		}
	});
</script>

<div class="flex flex-col gap-3 lg:gap-5">
	<h2 class="text-lg font-bold md:text-xl md:leading-tight dark:text-white">Ongoing Courses</h2>
	{#if loadingOngoingCourses}
		<div class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10">
			<Loading />
		</div>
	{:else if loadingOngoingCoursesError}
		<Error class="text-muted min-h-[6rem] p-5 text-sm" imgClass="h-6 w-6" />
	{:else if ongoingCourses.length === 0}
		<div class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10">
			<span class="text-muted-foreground">No courses have been started.</span>
		</div>
	{:else}
		<div class="grid gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
			{#each ongoingCourses as course}
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

						<span class="text-muted pt-3 text-xs">{course.percent}%</span>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
