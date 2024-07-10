<script lang="ts">
	import { CourseCard, Err, Loading, NiceDate, Pagination } from '$components/generic';
	import { CoursesFilter } from '$components/pages/courses';
	import { GetCourses } from '$lib/api';
	import type { Course, CourseProgress, CoursesGetParams } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { IsBrowser } from '$lib/utils';

	// ----------------------
	// Variables
	// ----------------------

	// The current fetched courses
	let fetchedCourses: Course[] = [];

	// The titles to filter on
	let filterTitles: string[] = [];

	// The tags to filter on
	let filterTags: string[] = [];

	// The progress to filter on
	let filterProgress: CourseProgress | undefined;

	// Pagination for courses
	let pagination: PaginationParams = {
		page: 1,
		perPage: 12,
		perPages: [], // not used,
		totalItems: -1,
		totalPages: -1
	};

	// A boolean promise that initially fetches the courses. It is used in an `await` block
	let courses = getCourses();

	// ----------------------
	// Functions
	// ----------------------

	// Get courses (paginated)
	async function getCourses(resetPage = false): Promise<boolean> {
		if (!IsBrowser) return false;

		const params: CoursesGetParams = {
			page: resetPage ? 1 : pagination.page,
			perPage: pagination.perPage
		};

		if (filterTitles && filterTitles.length > 0) {
			params.titles = filterTitles.join(',');
		}

		if (filterTags && filterTags.length > 0) {
			params.tags = filterTags.join(',');
		}

		params.progress = filterProgress;

		try {
			const response = await GetCourses(params);

			fetchedCourses = response.items as Course[];

			pagination = {
				...pagination,
				totalItems: response.totalItems,
				totalPages: response.totalPages
			};

			return true;
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="container flex flex-col gap-6 py-6">
	<div class="flex h-full w-full flex-col gap-5">
		<CoursesFilter
			on:titleFilter={(ev) => {
				filterTitles = ev.detail;
				courses = getCourses(true);
			}}
			on:tagsFilter={(ev) => {
				filterTags = ev.detail;
				courses = getCourses(true);
			}}
			on:progressFilter={(ev) => {
				filterProgress = ev.detail;
				courses = getCourses(true);
			}}
			on:clear={() => {
				filterTitles = [];
				filterTags = [];
				filterProgress = undefined;
				courses = getCourses(true);
			}}
		/>

		<!-- Courses -->
		<div class="flex h-full w-full">
			{#await courses}
				<Loading class="max-h-96" />
			{:then _}
				{#if fetchedCourses && fetchedCourses.length === 0}
					<div class="flex min-h-[6rem] w-full flex-grow flex-col items-center p-10">
						{#if filterTitles.length > 0 || filterTags.length > 0 || filterProgress}
							<span class="text-muted-foreground">No courses found with the selected filters.</span>
						{:else}
							<span class="text-muted-foreground">No courses.</span>
						{/if}
					</div>
				{:else}
					<div class="flex w-full flex-col gap-5 overflow-hidden pb-5">
						<div
							class="grid w-full auto-cols-fr grid-cols-[repeat(auto-fill,minmax(17.5rem,1fr))] gap-4"
						>
							{#each fetchedCourses as course}
								<a
									class="group relative grid h-full min-h-36 cursor-pointer grid-cols-2 gap-4 overflow-hidden whitespace-normal rounded-lg bg-muted p-2 sm:flex sm:flex-col sm:gap-0 sm:p-0"
									href={`/course/?id=${course.id}`}
								>
									{#if !course.available}
										<span
											class="absolute right-0 top-0 z-10 flex h-1 w-1 items-center justify-center rounded-bl-lg rounded-tr-lg bg-destructive p-3 text-center text-sm"
										>
											!
										</span>
									{/if}

									<CourseCard
										courseId={course.id}
										hasCard={course.hasCard}
										class="aspect-h-7 aspect-w-16 sm:aspect-h-7 sm:aspect-w-16"
										imgClass="rounded-lg object-cover object-center sm:rounded-b-none md:object-top"
										fallbackClass="bg-alt-1 inline-flex grow place-content-center items-center rounded-lg sm:rounded-b-none"
									/>

									<div
										class="flex h-full flex-grow flex-col justify-between text-base sm:p-3 sm:text-sm"
									>
										<h3 class="font-semibold group-hover:text-secondary">
											{course.title}
										</h3>

										<div class="flex flex-row justify-between">
											<NiceDate date={course.progressUpdatedAt} class="shrink-0 pt-3 text-xs" />

											<span class="flex w-full justify-end pt-3 text-xs">{course.percent}%</span>
										</div>
									</div>
								</a>
							{/each}
						</div>

						<Pagination
							type={'course'}
							{pagination}
							showPerPage={false}
							on:pageChange={(ev) => {
								pagination.page = ev.detail;
								courses = getCourses();
							}}
						/>
					</div>
				{/if}
			{:catch error}
				<Err class="min-h-[6rem] p-5 text-sm text-muted" imgClass="size-6" errorMessage={error} />
			{/await}
		</div>
	</div>
</div>
