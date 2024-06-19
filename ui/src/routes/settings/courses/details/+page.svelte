<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { DeleteCourseDialog } from '$components/dialogs';
	import { CourseCard, Err, Loading, NiceDate, ScanStatus } from '$components/generic';
	import {
		CourseDetailsChapters,
		CourseDetailsTags
	} from '$components/pages/settings_course_details';
	import Badge from '$components/ui/badge/badge.svelte';
	import Button from '$components/ui/button/button.svelte';
	import { AddScan, GetCourseFromParams } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import { IsBrowser } from '$lib/utils';
	import {
		CalendarPlus,
		CalendarSearch,
		CheckCircle2,
		Circle,
		CircleSlash,
		Folder,
		FolderSearch,
		Play,
		PlayCircle,
		Trash2
	} from 'lucide-svelte';
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

	// True when the assets need to be refreshed
	let assetRefresh = false;

	// True when the card needs to be refreshed
	let cardRefresh = false;

	// True when the tags need to be refreshed
	let tagsRefresh = false;

	// ----------------------
	// Functions
	// ----------------------

	// Lookup the course based upon the search params
	async function getCourse(): Promise<boolean> {
		if (!IsBrowser) false;

		try {
			const response = await GetCourseFromParams(pageParams);
			if (!response) throw new Error('Course not found');

			fetchedCourse = response;
			return true;
		} catch (error) {
			throw error;
			return true;
		}
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

<div class="bg-background flex w-full flex-col gap-4 pb-10">
	{#await coursePromise}
		<Loading />
	{:then _}
		<!-- Course Details -->
		<div class="bg-muted">
			<div class="container flex flex-col gap-4 py-6 md:py-10">
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
							<div class="flex max-w-[30rem] flex-row items-center gap-2.5 md:max-w-full">
								<Folder class="size-4 shrink-0" />
								<span class="text-xs">{fetchedCourse.path}</span>
							</div>

							<!-- Created at -->
							<div class="flex flex-row items-center gap-2.5">
								<CalendarPlus class="size-4 shrink-0" />
								<NiceDate date={fetchedCourse.createdAt} class="text-foreground text-xs" />
							</div>

							<!-- Update at || scan status -->
							<div class="flex flex-row items-center gap-2.5">
								<CalendarSearch class="size-4 shrink-0" />
								{#if !fetchedCourse.scanStatus}
									<NiceDate date={fetchedCourse.updatedAt} class="text-foreground text-xs" />
								{:else}
									<ScanStatus
										courseId={fetchedCourse.id}
										initialStatus={fetchedCourse.scanStatus}
										poll={writable(true)}
										waitingText="Queued for scan"
										processingText="Scanning"
										class="text-foreground justify-start text-xs"
										on:empty={(e) => {
											fetchedCourse = e.detail;

											assetRefresh = true;
											cardRefresh = true;
											tagsRefresh = true;
										}}
									/>
								{/if}
							</div>

							<div class="flex flex-row place-content-center gap-3 pt-3.5 md:place-content-start">
								<!-- Availability -->
								{#if fetchedCourse.available}
									<Badge
										class="bg-success text-success-foreground hover:bg-success items-center gap-1.5 rounded-sm"
									>
										<CheckCircle2 class="size-4" />
										Available
									</Badge>
								{:else}
									<Badge
										variant="destructive"
										class="hover:bg-destructive items-center gap-1.5 rounded-sm"
									>
										<CircleSlash class="size-4" />
										Unavailable
									</Badge>
								{/if}

								<!-- % completed -->
								{#if !fetchedCourse.started}
									<Badge
										class="bg-alt-1 hover:bg-alt-1 text-foreground items-center gap-1.5 rounded-sm"
									>
										<Circle class="size-4" />
										Not Started
									</Badge>
								{:else if fetchedCourse.percent === 100}
									<Badge class="hover:bg-primary items-center gap-1.5 rounded-sm">
										<CheckCircle2 class="size-4" />
										Completed
									</Badge>
								{:else}
									<Badge
										variant="secondary"
										class="hover:bg-secondary items-center gap-1.5 rounded-sm"
									>
										<PlayCircle class="size-4" />
										{fetchedCourse.percent}% completed
									</Badge>
								{/if}
							</div>
						</div>

						<!-- Course actions -->
						<div class="flex flex-row items-center gap-2.5">
							<Button
								variant="outline"
								class="hover:bg-primary bg-muted border-muted-foreground hover:text-primary-foreground hover:border-primary h-8 cursor-pointer gap-2 px-2.5"
								href="/course?id={fetchedCourse.id}"
							>
								<Play class="size-4" />
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
								class="hover:bg-primary bg-muted border-muted-foreground hover:text-primary-foreground hover:border-primary h-8 cursor-pointer gap-2 px-2.5"
								disabled={fetchedCourse.scanStatus !== ''}
								on:click={async () => {
									const s = await startScan(fetchedCourse.id);
									if (s) fetchedCourse.scanStatus = s;
								}}
							>
								<FolderSearch class="size-4" />
								Scan
							</Button>

							<Button
								variant="outline"
								class="hover:bg-destructive bg-muted border-muted-foreground hover:border-destructive hover:text-destructive-foreground h-8 cursor-pointer gap-2 px-2.5"
								on:click={() => {
									openDeleteDialog = true;
								}}
							>
								<Trash2 class="size-4" />
								Delete
							</Button>
						</div>

						<CourseDetailsTags courseId={fetchedCourse.id} bind:refresh={tagsRefresh} />
					</div>

					<div class="order-1 md:order-2">
						<CourseCard
							courseId={fetchedCourse.id}
							hasCard={fetchedCourse.hasCard}
							bind:refresh={cardRefresh}
							class="flex h-48 max-h-48 w-auto flex-col rounded-none"
							imgClass="border-alt-1/60 mx-auto min-h-0 max-w-full rounded-lg border"
							fallbackClass="bg-alt-1 inline-flex grow max-w-72 place-content-center items-center rounded-lg"
						/>
					</div>
				</div>
			</div>
		</div>

		<!-- Course content -->
		<CourseDetailsChapters courseId={fetchedCourse.id} bind:refresh={assetRefresh} />

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
