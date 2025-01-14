<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { DeleteCourseDialog } from '$components/dialogs';
	import { CourseCard, Err, Loading, NiceDate, ScanStatus } from '$components/generic';
	import { Icons } from '$components/icons';
	import {
		CourseDetailsChapters,
		CourseDetailsTags
	} from '$components/pages/settings_course_details';
	import Button from '$components/ui/button/button.svelte';
	import { Progress } from '$components/ui/progress';
	import { AddScan, GetCourseFromParams } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import { cn, IsBrowser } from '$lib/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { writable } from 'svelte/store';

	// ----------------------
	// variables
	// ----------------------

	// Used during the #await. It is initially set to a promise that never resolves to prevent
	// the page from rendering before the course is fetched, which occurs during onMount. This
	// is because the site is pre-rendered and as such we can only get the search params after
	// the page is mounted
	let coursePromise: Promise<boolean> = new Promise(() => {});

	// Holds the fetched course
	let fetchedCourse: Course;

	let pageParams: URLSearchParams;

	// True when the course is being deleted
	let openDeleteDialog = false;

	// True when a scan is completed and parts of the page need to be refreshed
	let refresh = false;

	// ----------------------
	// Functions
	// ----------------------

	// Lookup the course based upon the search params
	async function getCourse(): Promise<boolean> {
		if (!IsBrowser) false;

		const response = await GetCourseFromParams(pageParams);
		if (!response) throw new Error('Course not found');

		fetchedCourse = response;
		return true;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Start a scan for a course
	async function startScan(courseId: string) {
		try {
			const response = await AddScan(courseId);
			return response.status;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	// Due to the site being pre-rendered, we need to wait for the page to be mounted before we
	// can get the search params
	onMount(async () => {
		pageParams = $page.url.searchParams;
		coursePromise = getCourse();
	});
</script>

<div class="main pt-0">
	{#await coursePromise}
		<Loading class="max-h-96" />
	{:then _}
		<!-- Course Details -->
		<div class="bg-muted pb-6 pt-6">
			<div class="container flex flex-col gap-4">
				<div class="grid grid-cols-1 gap-5 md:grid-cols-3">
					<div
						class="order-2 flex flex-col items-center justify-between gap-5 md:order-1 md:col-span-2 md:items-start md:gap-8"
					>
						<!-- Course title -->
						<div class="text-center text-2xl font-bold md:text-start md:text-3xl">
							{fetchedCourse.title}
						</div>

						<div class="flex flex-col gap-2.5">
							<!-- Path -->
							<div
								class={cn(
									'flex max-w-[30rem] flex-row items-start gap-2.5 text-xs md:max-w-full',
									!fetchedCourse.available && 'text-destructive'
								)}
							>
								<Icons.Folder class="size-4 shrink-0" />
								{fetchedCourse.path}
							</div>

							<!-- Created at -->
							<div class="flex flex-row items-center gap-2.5">
								<Icons.CalendarPlus class="size-4 shrink-0" />
								<NiceDate
									date={fetchedCourse.createdAt}
									prefix="Added: "
									class="text-xs text-foreground"
								/>
							</div>

							<!-- Update at || scan status -->
							<div class="flex flex-row items-center gap-2.5">
								<Icons.CalendarScan class="size-4 shrink-0" />
								{#if !fetchedCourse.scanStatus}
									<NiceDate
										date={fetchedCourse.updatedAt}
										prefix="Last Scanned: "
										class="text-xs text-foreground"
									/>
								{:else}
									<ScanStatus
										courseId={fetchedCourse.id}
										initialStatus={fetchedCourse.scanStatus}
										poll={writable(true)}
										waitingText="Queued for scan"
										processingText="Scanning"
										class="justify-start text-xs text-foreground"
										on:empty={(e) => {
											fetchedCourse = e.detail;
											refresh = true;
										}}
									/>
								{/if}
							</div>

							<!-- Progress -->
							<div class="flex flex-row items-center gap-2.5 text-xs">
								<Icons.Hourglass class="size-4 shrink-0" />

								<div class="flex min-w-24 flex-row items-center gap-2.5">
									<Progress value={fetchedCourse.percent} class="h-1.5 bg-foreground/40" />
									<span class="min-w-5 text-xs">{fetchedCourse.percent}%</span>
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex flex-row items-center gap-2.5">
							<Button
								variant="outline"
								class="h-8 cursor-pointer gap-2 border-muted-foreground bg-muted px-2.5 hover:border-primary hover:bg-primary hover:text-primary-foreground"
								href="/course?id={fetchedCourse.id}"
							>
								<Icons.Play class="size-4" />
								{#if fetchedCourse.percent === 0}
									Start
								{:else if fetchedCourse.percent === 100}
									Replay
								{:else}
									Resume
								{/if}
							</Button>

							<Button
								variant="outline"
								class="h-8 cursor-pointer gap-2 border-muted-foreground bg-muted px-2.5 hover:border-primary hover:bg-primary hover:text-primary-foreground"
								disabled={fetchedCourse.scanStatus !== ''}
								on:click={async () => {
									const s = await startScan(fetchedCourse.id);
									if (s) fetchedCourse.scanStatus = s;
								}}
							>
								<Icons.Scan class="size-4" />
								Scan
							</Button>

							<Button
								variant="outline"
								class="h-8 cursor-pointer gap-2 border-muted-foreground bg-muted px-2.5 hover:border-destructive hover:bg-destructive hover:text-destructive-foreground"
								on:click={() => {
									openDeleteDialog = true;
								}}
							>
								<Icons.Trash class="size-4" />
								Delete
							</Button>
						</div>

						<!-- Tags -->
						<CourseDetailsTags courseId={fetchedCourse.id} bind:refresh />
					</div>

					<!-- Card -->
					<div class="order-1 md:order-2">
						<CourseCard
							courseId={fetchedCourse.id}
							hasCard={fetchedCourse.hasCard}
							bind:refresh
							class="flex h-48 max-h-48 w-auto flex-col items-center rounded-none md:items-end"
							imgClass="border-alt-1/60 min-h-0 max-w-full rounded-lg border"
							fallbackClass="bg-alt-1 inline-flex  grow max-w-72 place-content-center items-center w-full rounded-lg"
						/>
					</div>
				</div>
			</div>
		</div>

		<!-- Course content -->
		<CourseDetailsChapters courseId={fetchedCourse.id} bind:refresh />

		<!-- Delete dialog -->
		<DeleteCourseDialog
			courses={{ [fetchedCourse.id]: fetchedCourse.title }}
			bind:open={openDeleteDialog}
			on:deleted={() => {
				goto('/settings/courses');
			}}
		/>
	{:catch error}
		<Err errorMessage={error} />
	{/await}
</div>
