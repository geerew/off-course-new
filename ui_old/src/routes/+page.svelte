<script lang="ts">
	import AddCourses from '$components/dialogs/add-courses.svelte';
	import { Err, Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import { Carousel } from '$components/pages/home';
	import { Button } from '$components/ui/button';
	import { GetCourses } from '$lib/api';
	import { IsBrowser } from '$lib/utils';

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

		const response = await GetCourses({ page: 1, perPage: 1 });

		if (response.totalItems > 0) {
			showLandingPage = false;
		}
		return true;
	}
</script>

{#await getFirstCourse}
	<Loading class="max-h-96" />
{:then _}
	<div class="main container">
		{#if showLandingPage}
			<div class="flex flex-col gap-10 lg:flex-row lg:pt-16">
				<div
					class="order-2 flex flex-col place-items-center gap-4 lg:order-1 lg:basis-3/5 lg:place-content-center lg:place-items-start"
				>
					<h1
						class="w-full items-center text-center font-sans text-2xl font-bold sm:text-4xl lg:text-start"
					>
						View and manage course material locally
					</h1>
					<p
						class="max-w-[35rem] pb-2.5 text-center text-sm text-muted-foreground sm:text-base lg:w-11/12 lg:pb-4 lg:text-start"
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
							class="group flex h-10 w-56 justify-between gap-1.5 bg-primary hover:bg-primary hover:brightness-110 sm:w-44"
							on:click={async () => {
								open();
							}}
						>
							<div class="flex flex-row items-center gap-1.5">
								<Icons.StackPlus class="size-4" />
								<span>Add Courses</span>
							</div>

							<Icons.ArrowRight class="size-4" />
						</Button>
					</AddCourses>
				</div>

				<div class="order-1 flex w-full place-content-center lg:order-2 lg:basis-2/5">
					<Icons.Rocket class="size-96" />
				</div>
			</div>
		{:else}
			<Carousel variant="ongoing" />
			<Carousel variant="latest" />
		{/if}
	</div>
{:catch error}
	<Err class="min-h-[6rem] p-5 text-sm text-muted" imgClass="size-6" errorMessage={error} />
{/await}
