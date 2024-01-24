<script lang="ts">
	import { CourseCard, Error, Loading } from '$components';
	import { Button } from '$components/ui/button';
	import { ErrorMessage, GetCourses } from '$lib/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Carousel from '$lib/components/ui/carousel';
	import { addToast } from '$lib/stores/addToast';
	import type { Course, CoursesGetParams } from '$lib/types/models';
	import { isBrowser } from '$lib/utils';
	import { ArrowLeft, ArrowRight } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import theme from 'tailwindcss/defaultTheme';
	import { NiceDate } from './table/renderers';
	import type { CarouselAPI } from './ui/carousel/context';

	// ----------------------
	// Exports
	// ----------------------

	export let variant: 'ongoing' | 'latest';

	// ----------------------
	// Variables
	// ----------------------

	// True while the page is loading
	let loadingCourses = true;

	// True when an error occurred
	let loadingCoursesError = false;

	// Holds the courses
	let courses: Course[] = [];

	// The current page
	let currentPage = 1;

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

	// ----------------------
	// Functions
	// ----------------------

	// Get courses (paginated)
	async function getCourses(page: number) {
		if (!isBrowser) return false;

		const params: CoursesGetParams = {
			page,
			perPage: 8
		};

		if (variant === 'ongoing') {
			params.started = true;
			params.orderBy = 'progress_updated_at desc';
		}

		return await GetCourses(params)
			.then((resp) => {
				if (!resp) return false;

				// If the current page is 1, then we can just set the courses to the response, or
				// else append the response to the current courses
				courses.length === 0
					? (courses = resp.items as Course[])
					: (courses = [...courses, ...(resp.items as Course[])]);

				// Are there more courses to get?
				moreToGet = courses.length < resp.totalItems;

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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Load more courses
	async function loadMoreCourses() {
		if (!moreToGet) return;

		currentPage++;

		if (!(await getCourses(currentPage))) {
			loadingCoursesError = true;
			return;
		}
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
	// Lifecycle
	// ----------------------
	onMount(async () => {
		if (!(await getCourses(currentPage))) {
			loadingCourses = false;
			loadingCoursesError = true;
		}

		loadingCourses = false;
		return;
	});
</script>

<div class="flex flex-col">
	<div class="flex flex-row items-center justify-between pb-5">
		<h2 class="text-lg font-bold">{variant === 'latest' ? 'New Courses' : 'Ongoing Courses'}</h2>
		<div class="flex gap-1">
			<Button
				variant="ghost"
				disabled={!$canScrollPrev}
				on:click={() => {
					currentSlide -= scrollBy;
					api.scrollTo(currentSlide);
				}}
				class="hover:text-primary px-3"
			>
				<ArrowLeft class="h-6 w-6" />
			</Button>

			<Button
				variant="ghost"
				disabled={!$canScrollNext}
				on:click={() => {
					currentSlide += scrollBy;
					api.scrollTo(currentSlide);
				}}
				class="hover:text-primary px-3"
			>
				<ArrowRight class="h-6 w-6" />
			</Button>
		</div>
	</div>
	{#if loadingCourses}
		<div class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10">
			<Loading />
		</div>
	{:else if loadingCoursesError}
		<Error class="text-muted min-h-[6rem] p-5 text-sm" imgClass="h-6 w-6" />
	{:else if courses.length === 0}
		<div class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10">
			<span class="text-muted-foreground">No courses have been added.</span>
		</div>
	{:else}
		<Carousel.Root bind:api opts={{ watchSlides: true, align: 'start' }}>
			<Carousel.Content class="flex select-none">
				<!-- Courses -->
				{#each courses as course}
					<Carousel.Item class="group basis-1/2 md:basis-1/3 lg:basis-1/4 xl:basis-1/5">
						<Card.Root class="relative h-full">
							{#if !course.available}
								<span
									class="bg-destructive absolute right-0 top-0 z-10 flex h-1 w-1 items-center justify-center rounded-bl-lg rounded-tr-lg p-3 text-center text-sm"
								>
									!
								</span>
							{/if}

							<a href="/course?id={course.id}">
								<Card.Content class="bg-muted flex h-full flex-col overflow-hidden rounded-lg p-0">
									<CourseCard {course} />

									<div class="flex h-full flex-col justify-between p-3 text-sm md:p-3">
										<h3 class="group-hover:text-primary font-semibold">
											{course.title}
										</h3>

										<div class="flex flex-row justify-between">
											<NiceDate
												date={variant === 'latest' ? course.createdAt : course.progressUpdatedAt}
												prefix={variant === 'latest' ? 'Added:' : 'Last Viewed:'}
												class="shrink-0 pt-3 text-xs"
											/>

											<span class="flex w-full justify-end pt-3 text-xs">{course.percent}%</span>
										</div>
									</div>
								</Card.Content>
							</a>
						</Card.Root>
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
</div>
