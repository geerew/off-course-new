<script lang="ts">
	import { CourseCard, Err, Loading, NiceDate } from '$components/generic';
	import { Button } from '$components/ui/button';
	import { GetCourses } from '$lib/api';
	import * as Carousel from '$lib/components/ui/carousel';
	import { CourseProgress, type Course, type CoursesGetParams } from '$lib/types/models';
	import { ArrowLeft, ArrowRight } from 'lucide-svelte';
	import { writable } from 'svelte/store';
	import theme from 'tailwindcss/defaultTheme';
	import type { CarouselAPI } from '../../ui/carousel/context';

	// ----------------------
	// Exports
	// ----------------------

	export let variant: 'ongoing' | 'latest';

	// ----------------------
	// Functions
	// ----------------------

	// Get courses (paginated)
	async function getCourses(page: number, numCoursesToFetch: number): Promise<boolean> {
		const params: CoursesGetParams = {
			page,
			perPage: numCoursesToFetch
		};

		if (variant === 'ongoing') {
			params.progress = CourseProgress.Started;
			params.orderBy = 'progress_updated_at desc';
		}

		try {
			const response = await GetCourses(params);
			if (!response) throw new Error('Failed to get courses');

			// If the current page is 1, then we can just set the courses to the response, or
			// else append the response to the current courses
			fetchedCourses.length === 0
				? (fetchedCourses = response.items as Course[])
				: (fetchedCourses = [...fetchedCourses, ...(response.items as Course[])]);

			// Are there more courses to get?
			moreToGet = fetchedCourses.length < response.totalItems;

			return true;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load more courses
	async function loadMoreCourses() {
		if (!moreToGet) return;
		currentPage++;
		await getCourses(currentPage, numCoursesToFetch);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the scroll by value based on the screen size
	function setScrollBy() {
		if (window.innerWidth < mdPx) {
			scrollBy = 1;
		} else if (window.innerWidth < lgPx) {
			scrollBy = 2;
		} else if (window.innerWidth < xlPx) {
			scrollBy = 3;
		} else {
			scrollBy = 4;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Return a random empty message. To be used when no courses are found
	function randomEmptyMessage(): string {
		const phrases = [
			'Not a course in sight',
			'A blank slate',
			'Tumbleweeds...',
			"I've got nothing...",
			'All quiet on the course front',
			'Course-free zone',
			'Silent halls',
			"Do you hear that? It's the sound of no courses",
			'No courses to be found',
			'Nothing to see here',
			'ECHO echo echo',
			'Waiting for liftoff',
			'Uncharted territories',
			'The great course void',
			'No courses, only dreams',
			'Where courses dare not tread',
			'A sea of tranquility',
			'Awaiting ignition'
		];

		const randomIndex = Math.floor(Math.random() * phrases.length);
		return phrases[randomIndex];
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When the carousel API is set, set the scroll variables
	$: if (api) {
		canScrollPrev.set(api.canScrollPrev());
		canScrollNext.set(api.canScrollNext());

		// Calculate the scroll by value based upon the screen size
		setScrollBy();

		api.on('select', () => {
			canScrollPrev.set(api.canScrollPrev());
			canScrollNext.set(api.canScrollNext());

			// Update the current item to be the selected item as this may differ from what we
			// tried to scroll to, for example, if we tried to scroll to an item that doesn't
			// exist
			currentSlide = api.selectedScrollSnap();
		});

		api.on('resize', () => {
			// Update the scroll by value based on the new screen size
			setScrollBy();

			canScrollPrev.set(api.canScrollPrev());
			canScrollNext.set(api.canScrollNext());
		});

		api.on('reInit', () => {
			canScrollPrev.set(api.canScrollPrev());
			canScrollNext.set(api.canScrollNext());
		});
	}

	// ----------------------
	// Variables
	// ----------------------

	// The current page
	let currentPage = 1;

	// The current fetched courses
	let fetchedCourses: Course[] = [];

	// The number of courses to fetch
	let numCoursesToFetch = 8;

	// Holds the courses
	let courses = getCourses(currentPage, numCoursesToFetch);

	// True when there are more courses to get
	let moreToGet = false;

	// Carousel API
	let api: CarouselAPI;

	// True when the user can scroll to the previous/next
	let canScrollPrev = writable(false);
	let canScrollNext = writable(false);

	// How many slides to scroll by
	let scrollBy: number;

	// The currently selected slide
	let currentSlide = 0;

	// Screen sizes
	const mdPx = +theme.screens.md.replace('px', '');
	const lgPx = +theme.screens.lg.replace('px', '');
	const xlPx = +theme.screens.xl.replace('px', '');
</script>

<div class="flex flex-col">
	<div class="flex flex-row items-center justify-between pb-5">
		<h2 class="text-lg font-bold">{variant === 'latest' ? 'New Courses' : 'Ongoing Courses'}</h2>
		<div class="flex gap-1">
			{#await courses then _}
				<Button
					variant="ghost"
					disabled={!$canScrollPrev}
					on:click={() => {
						currentSlide -= scrollBy;
						api.scrollTo(currentSlide);
					}}
					class="hover:text-secondary px-3"
				>
					<ArrowLeft class="size-6" />
				</Button>

				<Button
					variant="ghost"
					disabled={!$canScrollNext}
					on:click={() => {
						currentSlide += scrollBy;
						api.scrollTo(currentSlide);
					}}
					class="hover:text-secondary px-3"
				>
					<ArrowRight class="size-6" />
				</Button>
			{/await}
		</div>
	</div>
	{#await courses}
		<Loading />
	{:then _}
		{#if fetchedCourses.length === 0}
			<div
				class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10"
			>
				<span class="text-muted-foreground">
					{randomEmptyMessage()}
				</span>
			</div>
		{:else}
			<Carousel.Root bind:api opts={{ watchSlides: true, align: 'start' }}>
				<Carousel.Content class="flex select-none">
					<!-- Courses -->
					{#each fetchedCourses as course}
						<Carousel.Item class="group basis-1/2 md:basis-1/3 lg:basis-1/4 xl:basis-1/5">
							<a
								class="bg-muted group relative flex h-full min-h-36 cursor-pointer flex-col gap-4 overflow-hidden whitespace-normal rounded-lg"
								href={`/course/?id=${course.id}`}
							>
								{#if !course.available}
									<span
										class="bg-destructive absolute right-0 top-0 z-10 flex h-1 w-1 items-center justify-center rounded-bl-lg rounded-tr-lg p-3 text-center text-sm"
									>
										!
									</span>
								{/if}

								<CourseCard
									courseId={course.id}
									hasCard={course.hasCard}
									class="aspect-w-16 aspect-h-7 sm:aspect-w-16 sm:aspect-h-7"
									imgClass="rounded-lg object-cover object-center sm:rounded-b-none md:object-top"
									fallbackClass="bg-alt-1 inline-flex grow place-content-center items-center rounded-lg sm:rounded-b-none"
								/>

								<div class="flex h-full flex-grow flex-col justify-between p-2 text-sm">
									<h3 class="group-hover:text-secondary font-semibold">
										{course.title}
									</h3>

									<div class="flex flex-row justify-between">
										<NiceDate date={course.progressUpdatedAt} class="shrink-0 pt-3 text-xs" />

										<span class="flex w-full justify-end pt-3 text-xs">{course.percent}%</span>
									</div>
								</div>
							</a>
						</Carousel.Item>
					{/each}

					<!-- Load more -->
					{#if moreToGet}
						<Carousel.Item class="basis-1/2 md:basis-1/3 lg:basis-1/4 xl:basis-1/5">
							<Button
								variant="outline"
								class="hover:text-primary h-full w-full rounded-lg text-lg"
								on:click={loadMoreCourses}>Load More</Button
							>
						</Carousel.Item>
					{/if}
				</Carousel.Content>
			</Carousel.Root>
		{/if}
	{:catch error}
		<Err class="text-muted min-h-[6rem] p-5 text-sm" imgClass="size-6" errorMessage={error} />
	{/await}
</div>
