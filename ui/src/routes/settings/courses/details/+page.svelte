<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Error, Loading } from '$components';
	import { CourseDetailsTags } from '$components/course_details';
	import { DeleteCourse } from '$components/dialogs';
	import { NiceDate, ScanStatus } from '$components/table/renderers';
	import * as Accordion from '$components/ui/accordion';
	import * as Avatar from '$components/ui/avatar';
	import Badge from '$components/ui/badge/badge.svelte';
	import Button from '$components/ui/button/button.svelte';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { ATTACHMENT_API, AddScan, COURSE_API, ErrorMessage, GetAllCourseAssets } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { Asset, Course, CourseChapters } from '$lib/types/models';
	import { GetCourseFromParams, buildChapterStructure, cn, isBrowser } from '$lib/utils';
	import {
		CalendarPlus,
		CalendarSearch,
		CheckCircle2,
		ChevronRight,
		Circle,
		CircleSlash,
		Dot,
		Download,
		Folder,
		FolderSearch,
		Play,
		PlayCircle,
		Trash2
	} from 'lucide-svelte';
	import { onMount } from 'svelte';

	// ----------------------
	// Variables
	// ----------------------

	// True while the page is loading
	let loadingCourse = true;

	// True when there was an error getting the course
	let gotCourseError = false;

	// True while the page assets are loading
	let loadingAssets = true;

	// True when there was an error getting the assets
	let gotAssetsError = false;

	// Holds the information about the course being viewed
	let course: Course;

	// Holds the assets + attachments for this course
	let assets: Asset[];

	// Holds the course assets in a chapter structure. Populated when getCourse is called in
	// onMount
	let chapters: CourseChapters = {};

	// This will be set to the src of the course card if the course has a card
	let cardSrc = '';

	// True when the course is being deleted
	let openDeleteDialog = false;

	// ----------------------
	// Functions
	// ----------------------

	// Lookup the course based upon the search params
	const getCourse = async () => {
		if (!isBrowser) return false;

		return await GetCourseFromParams($page.url.searchParams)
			.then((resp) => {
				course = resp;
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Gets the assets + attachments for the given course. It will then build a chapter structure
	// for the assets and selected the first asset that is not completed. If the course itself is
	// completed, the first asset will be selected
	const getAssets = async (courseId: string) => {
		if (!isBrowser) return false;

		return await GetAllCourseAssets(courseId, { orderBy: 'chapter asc, prefix asc' })
			.then(async (resp) => {
				if (!resp) return false;
				chapters = buildChapterStructure(resp);
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Sets the src if the course has a card
	function setCardSrc() {
		if (course && course.hasCard) {
			cardSrc = `${COURSE_API}/${course.id}/card?b=${new Date().getTime()}`;
		} else {
			cardSrc = '';
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Start a scan for a course
	const startScan = async () => {
		await AddScan(course.id)
			.then(() => {
				// Manually update the course scan status to 'waiting' and then update the courses
				// array
				course.scanStatus = 'waiting';
			})
			.catch((err) => {
				console.error(err);
			});
	};

	// ----------------------
	// Reactive
	// ----------------------

	$: totalChapterCount = Object.keys(chapters).length;

	$: totalAssetCount = Object.values(chapters).reduce((sum, currentAssets) => {
		return sum + currentAssets.length;
	}, 0);

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		if (!(await getCourse())) {
			loadingCourse = false;
			gotCourseError = true;
			return;
		}

		loadingCourse = false;

		setCardSrc();

		if (!(await getAssets(course.id))) {
			loadingAssets = false;
			gotAssetsError = true;
			return;
		}

		loadingAssets = false;
	});
</script>

<div class="bg-background flex w-full flex-col gap-4 pb-10">
	{#if loadingCourse}
		<div
			class="flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center p-10"
		>
			<Loading class="border-primary" />
		</div>
	{:else if gotCourseError}
		<Error />
	{:else}
		<!-- Course Details -->
		<div class="bg-muted">
			<div class="container flex flex-col gap-4 py-6 md:py-10">
				<div class="grid grid-cols-1 gap-5 md:grid-cols-3">
					<div
						class="order-2 flex flex-col items-center justify-between gap-5 md:order-1 md:col-span-2 md:items-start md:gap-8"
					>
						<!-- Course title -->
						<div class="text-center text-2xl font-bold md:text-start md:text-3xl">
							{course.title}
						</div>

						<div class="flex flex-col gap-2.5">
							<!-- Path -->
							<div class="flex max-w-[30rem] flex-row items-center gap-2.5 md:max-w-full">
								<Folder class="h-4 w-4 shrink-0" />
								<span class="text-xs">{course.path}</span>
							</div>

							<!-- Created at -->
							<div class="flex flex-row items-center gap-2.5">
								<CalendarPlus class="h-4 w-4 shrink-0" />
								<NiceDate date={course.createdAt} class="text-foreground text-xs" />
							</div>

							<!-- Update at || scan status -->
							<div class="flex flex-row items-center gap-2.5">
								<CalendarSearch class="h-4 w-4 shrink-0" />
								{#if !course.scanStatus}
									<NiceDate date={course.updatedAt} class="text-foreground text-xs" />
								{:else}
									<ScanStatus
										courseId={course.id}
										scanStatus={course.scanStatus}
										waitingText="Queued for scan"
										processingText="Scanning"
										class="text-foreground justify-start text-xs"
										on:change={async (e) => {
											course = e.detail;

											// Pull any changes to assets and pull the card (again)
											if (!course.scanStatus) {
												await getAssets(course.id);
												setCardSrc();
											}
										}}
									/>
								{/if}
							</div>

							<div class="flex flex-row place-content-center gap-3 pt-3.5 md:place-content-start">
								<!-- Availability -->
								{#if course.available}
									<Badge
										class="bg-success text-success-foreground hover:bg-success items-center gap-1.5 rounded-sm"
									>
										<CheckCircle2 class="h-4 w-4" />
										Available
									</Badge>
								{:else}
									<Badge
										variant="destructive"
										class="hover:bg-destructive items-center gap-1.5 rounded-sm"
									>
										<CircleSlash class="h-4 w-4" />
										Unavailable
									</Badge>
								{/if}

								<!-- % completed -->
								{#if !course.started}
									<Badge
										class="bg-alt-1 hover:bg-alt-1 text-foreground items-center gap-1.5 rounded-sm"
									>
										<Circle class="h-4 w-4" />
										Not Started
									</Badge>
								{:else if course.percent === 100}
									<Badge class="hover:bg-primary items-center gap-1.5 rounded-sm">
										<CheckCircle2 class="h-4 w-4" />
										Completed
									</Badge>
								{:else}
									<Badge
										variant="secondary"
										class="hover:bg-secondary items-center gap-1.5 rounded-sm"
									>
										<PlayCircle class="h-4 w-4" />
										{course.percent}% completed
									</Badge>
								{/if}
							</div>
						</div>

						<!-- Course actions -->
						<div class="flex flex-row items-center gap-2.5">
							<Button
								variant="outline"
								class="hover:bg-primary bg-muted border-muted-foreground hover:text-primary-foreground hover:border-primary h-8 cursor-pointer gap-2 px-2.5"
								href="/course?id={course.id}"
							>
								<Play class="h-4 w-4" />
								{#if course.percent === 0}
									Start
								{:else if course.percent === 100}
									Replay
								{:else}
									Resume
								{/if}
							</Button>

							<Button
								variant="outline"
								class="hover:bg-primary bg-muted border-muted-foreground hover:text-primary-foreground hover:border-primary h-8 cursor-pointer gap-2 px-2.5"
								disabled={course.scanStatus !== ''}
								on:click={startScan}
							>
								<FolderSearch class="h-4 w-4" />
								Scan
							</Button>

							<Button
								variant="outline"
								class="hover:bg-destructive bg-muted border-muted-foreground hover:border-destructive hover:text-destructive-foreground h-8 cursor-pointer gap-2 px-2.5"
								on:click={() => {
									openDeleteDialog = true;
								}}
							>
								<Trash2 class="h-4 w-4" />
								Delete
							</Button>
						</div>

						<CourseDetailsTags courseId={course.id} />
					</div>

					<!-- Card -->
					<div class="order-1 text-xl font-bold md:order-2 md:text-2xl">
						<Avatar.Root class="flex h-48 max-h-48 w-auto flex-col rounded-none">
							<Avatar.Image src={cardSrc} class="mx-auto min-h-0 max-w-full rounded-lg" />
							<Avatar.Fallback
								class="bg-background mx-auto flex h-48 max-w-72 place-content-center rounded-lg lg:w-full"
							>
								<Play class="fill-primary text-primary h-12 w-12 opacity-60" />
							</Avatar.Fallback>
						</Avatar.Root>
					</div>
				</div>
			</div>
		</div>

		<!-- Course content -->
		<div class="container flex flex-col gap-4 py-4">
			<div class="flex flex-col gap-2.5 pl-2">
				<span class="text-xl font-bold">Course Content</span>
				<div class="flex flex-row items-center">
					<span class="text-muted-foreground text-sm">
						{totalChapterCount}
						{totalChapterCount ? 'chapters' : 'chapter'}
					</span>
					<Dot class="text-muted-foreground h-5 w-5" />
					<span class="text-muted-foreground text-sm">
						{totalAssetCount}
						{totalAssetCount ? 'assets' : 'asset'}
					</span>
				</div>
			</div>

			<Accordion.Root class="border-muted/70 w-full rounded-lg border">
				{#each Object.keys(chapters) as chapter, i}
					{@const numAssets = chapters[chapter].length}
					{@const lastChapter = Object.keys(chapters).length - 1 == i}

					<Accordion.Item
						value={chapter}
						class={cn('border-background ', lastChapter && 'border-b-none')}
					>
						<!-- Chapter -->
						<Accordion.Trigger
							class={cn(
								'bg-muted/70 hover:bg-muted px-5 py-4 hover:no-underline',
								i === 0 && 'rounded-t-lg',
								lastChapter && 'rounded-b-lg'
							)}
						>
							<span class="grow text-start text-base font-semibold">{chapter}</span>
							<span class="text-muted-foreground shrink-0 px-2.5 text-sm">
								{numAssets}
								{numAssets > 1 ? 'assets' : 'asset'}
							</span>
						</Accordion.Trigger>

						<!-- Assets -->
						<Accordion.Content class="flex flex-col">
							{#each chapters[chapter] as asset, i}
								{@const lastAsset = chapters[chapter].length - 1 == i}

								<!-- Asset -->
								<div class={cn(!lastAsset && 'border-muted/70 border-b')}>
									<div class="flex flex-row gap-5 px-5 py-4">
										<!-- Asset information (left)-->
										<div class="flex grow flex-col gap-2">
											<!-- Title -->
											<div class="flex flex-row items-center gap-4">
												<span>{asset.prefix}. {asset.title}</span>
											</div>

											<div
												class="text-muted-foreground flex select-none flex-row flex-wrap items-center gap-y-2 text-xs"
											>
												<!-- Type -->
												<span>{asset.assetType}</span>

												<!-- Progress -->
												{#if asset.completed}
													<Dot class="h-5 w-5" />
													<span class="text-success font-bold"> completed </span>
												{:else if asset.assetType === 'video' && asset.videoPos > 0}
													<Dot class="h-5 w-5" />
													<span class="text-secondary"> in-progress </span>
												{/if}

												<!-- Attachments -->
												{#if asset.attachments && asset.attachments.length > 0}
													<Dot class="h-5 w-5" />

													<DropdownMenu.Root closeOnItemClick={false}>
														<DropdownMenu.Trigger asChild let:builder>
															<Button
																builders={[builder]}
																variant="ghost"
																class="group flex h-auto items-center gap-1 px-0 py-0 text-xs hover:bg-transparent"
															>
																{asset.attachments.length} attachment{asset.attachments.length > 1
																	? 's'
																	: ''}

																<ChevronRight
																	class="h-3 w-3 duration-200 group-data-[state=open]:rotate-90"
																/>
															</Button>
														</DropdownMenu.Trigger>

														<DropdownMenu.Content
															class="flex max-h-[10rem] w-auto max-w-xs flex-col overflow-y-scroll md:max-w-sm"
															fitViewport={true}
														>
															{#each asset.attachments as attachment, i}
																{@const lastAttachment = asset.attachments.length - 1 == i}
																<DropdownMenu.Item
																	class="cursor-pointer justify-between gap-3 text-xs"
																	href={ATTACHMENT_API + '/' + attachment.id + '/serve'}
																	download
																>
																	<div class="flex flex-row gap-1.5">
																		<span class="shrink-0">{i + 1}.</span>
																		<span class="grow">{attachment.title}</span>
																	</div>

																	<Download class="flex h-3 w-3 shrink-0" />
																</DropdownMenu.Item>

																{#if !lastAttachment}
																	<DropdownMenu.Separator
																		class="bg-muted my-1 -ml-1 -mr-1 block h-px"
																	/>
																{/if}
															{/each}
														</DropdownMenu.Content>
													</DropdownMenu.Root>
												{/if}
											</div>
										</div>

										<!-- Open button (right)-->
										<div class="flex items-center">
											<Button
												class="h-6 shrink-0 px-2 py-1"
												href="/course?id={course.id}&a={asset.id}"
											>
												{#if asset.assetType !== 'video' || asset.completed || asset.videoPos === 0}
													<span>Open</span>
												{:else}
													<span>Resume</span>
												{/if}
											</Button>
										</div>
									</div>
								</div>
							{/each}
						</Accordion.Content>
					</Accordion.Item>
				{/each}
			</Accordion.Root>
		</div>

		<!-- Delete dialog -->
		<DeleteCourse
			courseId={course.id}
			bind:open={openDeleteDialog}
			on:courseDeleted={() => {
				goto('/settings/courses');
			}}
		/>
	{/if}
</div>
