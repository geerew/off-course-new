<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { DeleteCourseDialog } from '$components/dialogs';
	import { Err, Loading, NiceDate, ScanStatus } from '$components/generic';
	import {
		CourseDetailsCard,
		CourseDetailsChapters,
		CourseDetailsTags
	} from '$components/pages/settings_course_details';
	import Badge from '$components/ui/badge/badge.svelte';
	import Button from '$components/ui/button/button.svelte';
	import { AddScan, GetCourseFromParams } from '$lib/api';
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
	import { toast } from 'svelte-sonner';

	// ----------------------
	// variables
	// ----------------------

	let course: Record<string, string> = {};

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
	async function getCourse() {
		try {
			const response = await GetCourseFromParams($page.url.searchParams);
			if (!response) throw new Error('Course not found');

			return response;
		} catch (error) {
			throw error;
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
</script>

<div class="bg-background flex w-full flex-col gap-4 pb-10">
	{#await getCourse()}
		<Loading />
	{:then data}
		<!-- Course Details -->
		<div class="bg-muted">
			<div class="container flex flex-col gap-4 py-6 md:py-10">
				<div class="grid grid-cols-1 gap-5 md:grid-cols-3">
					<div
						class="order-2 flex flex-col items-center justify-between gap-5 md:order-1 md:col-span-2 md:items-start md:gap-8"
					>
						<!-- Course title -->
						<div class="text-center text-2xl font-bold md:text-start md:text-3xl">
							{data.title}
						</div>

						<div class="flex flex-col gap-2.5">
							<!-- Path -->
							<div class="flex max-w-[30rem] flex-row items-center gap-2.5 md:max-w-full">
								<Folder class="size-4 shrink-0" />
								<span class="text-xs">{data.path}</span>
							</div>

							<!-- Created at -->
							<div class="flex flex-row items-center gap-2.5">
								<CalendarPlus class="size-4 shrink-0" />
								<NiceDate date={data.createdAt} class="text-foreground text-xs" />
							</div>

							<!-- Update at || scan status -->
							<div class="flex flex-row items-center gap-2.5">
								<CalendarSearch class="size-4 shrink-0" />
								{#if !data.scanStatus}
									<NiceDate date={data.updatedAt} class="text-foreground text-xs" />
								{:else}
									<ScanStatus
										courseId={data.id}
										scanStatus={data.scanStatus}
										waitingText="Queued for scan"
										processingText="Scanning"
										class="text-foreground justify-start text-xs"
										on:change={async (e) => {
											Object.assign(data, e.detail);

											// Explicitly set the scan status
											data.scanStatus = e.detail.scanStatus;

											// Pull any changes
											if (!data.scanStatus) {
												assetRefresh = true;
												cardRefresh = true;
												tagsRefresh = true;
											}
										}}
									/>
								{/if}
							</div>

							<div class="flex flex-row place-content-center gap-3 pt-3.5 md:place-content-start">
								<!-- Availability -->
								{#if data.available}
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
								{#if !data.started}
									<Badge
										class="bg-alt-1 hover:bg-alt-1 text-foreground items-center gap-1.5 rounded-sm"
									>
										<Circle class="size-4" />
										Not Started
									</Badge>
								{:else if data.percent === 100}
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
										{data.percent}% completed
									</Badge>
								{/if}
							</div>
						</div>

						<!-- Course actions -->
						<div class="flex flex-row items-center gap-2.5">
							<Button
								variant="outline"
								class="hover:bg-primary bg-muted border-muted-foreground hover:text-primary-foreground hover:border-primary h-8 cursor-pointer gap-2 px-2.5"
								href="/course?id={data.id}"
							>
								<Play class="size-4" />
								{#if data.percent === 0}
									Start
								{:else if data.percent === 100}
									Replay
								{:else}
									Resume
								{/if}
							</Button>

							<Button
								variant="outline"
								class="hover:bg-primary bg-muted border-muted-foreground hover:text-primary-foreground hover:border-primary h-8 cursor-pointer gap-2 px-2.5"
								disabled={data.scanStatus !== ''}
								on:click={async () => {
									const newStatus = await startScan(data.id);
									if (newStatus) data.scanStatus = newStatus;
								}}
							>
								<FolderSearch class="size-4" />
								Scan
							</Button>

							<Button
								variant="outline"
								class="hover:bg-destructive bg-muted border-muted-foreground hover:border-destructive hover:text-destructive-foreground h-8 cursor-pointer gap-2 px-2.5"
								on:click={() => {
									course = { [data.id]: data.title };
									openDeleteDialog = true;
								}}
							>
								<Trash2 class="size-4" />
								Delete
							</Button>
						</div>

						<CourseDetailsTags courseId={data.id} bind:tagsRefresh />
					</div>

					<CourseDetailsCard courseId={data.id} hasCard={data.hasCard} bind:cardRefresh />
				</div>
			</div>
		</div>

		<!-- Course content -->
		<CourseDetailsChapters courseId={data.id} bind:assetRefresh />

		<!-- Delete dialog -->
		<DeleteCourseDialog
			courses={course}
			bind:open={openDeleteDialog}
			on:courseDeleted={() => {
				goto('/settings/courses');
			}}
		/>
	{:catch error}
		<Err errorMessage={error} />
	{/await}
</div>
