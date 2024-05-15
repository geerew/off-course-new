<script lang="ts">
	import { CourseCard, Err, Loading, NiceDate, Pagination } from '$components/generic';
	import { CoursesFilter } from '$components/pages/courses';
	import * as Card from '$components/ui/card';
	import { GetCourses } from '$lib/api';
	import type { Course, CourseProgress, CoursesGetParams } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { toast } from 'svelte-sonner';

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
	async function getCourses(): Promise<boolean> {
		const params: CoursesGetParams = {
			page: pagination.page,
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
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	}
</script>

<div class="container flex flex-col gap-6 py-6">
	<div class="flex h-full w-full flex-col gap-5">
		<CoursesFilter
			on:titleFilter={(ev) => {
				filterTitles = ev.detail;
				courses = getCourses();
			}}
			on:tagsFilter={(ev) => {
				filterTags = ev.detail;
				courses = getCourses();
			}}
			on:progressFilter={(ev) => {
				filterProgress = ev.detail;
				courses = getCourses();
			}}
			on:clear={() => {
				filterTitles = [];
				filterTags = [];
				filterProgress = undefined;
				courses = getCourses();
			}}
		/>

		<!-- Courses -->
		<div class="flex h-full w-full">
			{#await courses}
				<Loading />
			{:then _}
				{#if fetchedCourses.length === 0}
					<div class="flex min-h-[6rem] w-full flex-grow flex-col items-center p-10">
						<span class="text-muted-foreground">No courses.</span>
					</div>
				{:else}
					<div class="flex flex-col gap-5 pb-5">
						<div class="grid grid-cols-1 gap-5 md:grid-cols-3 xl:grid-cols-4">
							{#each fetchedCourses as course}
								<Card.Root class="group relative h-full">
									{#if !course.available}
										<span
											class="bg-destructive absolute right-0 top-0 z-10 flex h-1 w-1 items-center justify-center rounded-bl-lg rounded-tr-lg p-3 text-center text-sm"
										>
											!
										</span>
									{/if}

									<a href="/course?id={course.id}">
										<Card.Content
											class="bg-muted flex h-full flex-col overflow-hidden rounded-lg p-0"
										>
											<CourseCard {course} />

											<div class="flex h-full flex-col justify-between p-3 text-sm md:p-3">
												<h3 class="group-hover:text-secondary font-semibold">
													{course.title}
												</h3>

												<div class="flex flex-row justify-between">
													<NiceDate date={course.progressUpdatedAt} class="shrink-0 pt-3 text-xs" />

													<span class="flex w-full justify-end pt-3 text-xs">{course.percent}%</span
													>
												</div>
											</div>
										</Card.Content>
									</a>
								</Card.Root>
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
				<Err class="text-muted min-h-[6rem] p-5 text-sm" imgClass="size-6" errorMessage={error} />
			{/await}
		</div>
	</div>
</div>
