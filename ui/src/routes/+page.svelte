<script lang="ts">
	import AddCourses from '$components/dialogs/add-courses.svelte';
	import { Err, Loading } from '$components/generic';
	import { Carousel } from '$components/pages/home';
	import { Button } from '$components/ui/button';
	import { GetCourses } from '$lib/api';
	import { IsBrowser } from '$lib/utils';
	import { ArrowRight, BookPlus } from 'lucide-svelte';

	// ----------------------
	// Variables
	// ----------------------

	// A boolean promise that initially fetches the first course. It is used in an `await`
	// block and determines if we should show a landing page or not
	let getFirstCourse = getCourse();

	// Whether to show the landing page or not
	let showLandingPage = true;

	// ----------------------
	// Functions
	// ----------------------

	// Get the first course
	async function getCourse(): Promise<boolean> {
		if (!IsBrowser) return false;

		try {
			const response = await GetCourses({ page: 1, perPage: 1 });

			if (response.totalItems > 0) {
				showLandingPage = false;
			}
			return true;
		} catch (error) {
			throw error;
		}
	}

	let open = false;
</script>

{#await getFirstCourse}
	<Loading class="max-h-96" />
{:then _}
	{#if showLandingPage}
		<div class="container flex flex-col gap-6 py-6">
			<div class="flex flex-row pt-10">
				<div class="flex flex-col items-center gap-4 sm:items-start md:basis-4/5 lg:basis-3/5">
					<h1 class=" text-center font-sans text-2xl font-bold sm:text-start sm:text-4xl">
						View and manage course material locally
					</h1>
					<p
						class="text-muted-foreground w-11/12 pb-4 text-center text-sm sm:pb-7 sm:text-start sm:text-base"
					>
						Effortlessly view and organize your course content locally. Dive into learning with ease
						and let’s get started by adding some courses!
					</p>

					<AddCourses
						on:added={() => {
							showLandingPage = false;
						}}
					>
						<Button
							let:open
							slot="trigger"
							variant="outline"
							class="bg-primary hover:bg-primary group flex h-10 w-56 justify-between gap-1.5 hover:brightness-110 sm:w-44"
							on:click={async () => {
								open();
							}}
						>
							<div class="flex flex-row items-center gap-1.5">
								<BookPlus class="size-4" />
								<span>Add Courses</span>
							</div>

							<ArrowRight class="size-4" />
						</Button>
					</AddCourses>
				</div>
			</div>
		</div>
	{:else}
		<div class="container flex flex-col gap-6 py-6">
			<Carousel variant="ongoing" />
			<Carousel variant="latest" />
		</div>
	{/if}
{:catch error}
	<Err class="text-muted min-h-[6rem] p-5 text-sm" imgClass="size-6" errorMessage={error} />
{/await}
