<script lang="ts">
	import { Err, Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import * as Accordion from '$components/ui/accordion';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { ATTACHMENT_API, GetAllCourseAssets, GetBackendUrl } from '$lib/api';
	import { type Asset, type CourseChapters } from '$lib/types/models';
	import { BuildChapterStructure, cn } from '$lib/utils';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------

	export let courseId: string;
	export let refresh: boolean;

	// ----------------------
	// Variables
	// ----------------------

	let fectchedCourseChapters: CourseChapters = {};

	let courseChapters = getCourseChapters(courseId);

	// ----------------------
	// Functions
	// ----------------------

	// Gets the assets + attachments for the given course. It will then build a chapter structure
	// for the assets and selected the first asset that is not completed. If the course itself is
	// completed, the first asset will be selected
	//
	// During a refresh there is a small delay to prevent flickering
	async function getCourseChapters(courseId: string): Promise<boolean> {
		const refreshPromise = new Promise((resolve) => setTimeout(resolve, refresh ? 500 : 0));

		refresh = false;

		try {
			let response: Asset[];

			await Promise.all([
				(response = await GetAllCourseAssets(courseId, {
					orderBy: 'chapter asc, prefix asc',
					expand: true
				})),
				refreshPromise
			]);

			fectchedCourseChapters = BuildChapterStructure(response);

			return true;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function totalAssetCount(courseChapters: CourseChapters) {
		return Object.values(courseChapters).reduce((sum, currentAssets) => {
			return sum + currentAssets.length;
		}, 0);
	}

	// ----------------------
	// Reactive
	// ----------------------

	// Update course chapters when `refresh` is set to true
	$: if (refresh) {
		courseChapters = getCourseChapters(courseId);
	}
</script>

<div class="container flex flex-col gap-2 py-4">
	<div class="flex flex-col pl-2">
		<span class="text-xl font-bold">Course Content</span>
	</div>

	{#await courseChapters}
		<Loading class="min-h-24" />
	{:then _}
		<!-- n chapters / n assets -->
		<div class="flex flex-row items-center pb-4 pl-2.5">
			<span class="text-muted-foreground text-sm">
				{Object.keys(fectchedCourseChapters).length}
				{Object.keys(fectchedCourseChapters).length ? 'chapters' : 'chapter'}
			</span>
			<Icons.Dot weight="fill" class="text-muted-foreground size-5" />
			<span class="text-muted-foreground text-sm">
				{totalAssetCount(fectchedCourseChapters)}
				{totalAssetCount(fectchedCourseChapters) ? 'assets' : 'asset'}
			</span>
		</div>

		<Accordion.Root class="border-muted/70 w-full rounded-lg border">
			{#each Object.keys(fectchedCourseChapters) as chapter, i}
				{@const numAssets = fectchedCourseChapters[chapter].length}
				{@const lastChapter = Object.keys(fectchedCourseChapters).length - 1 == i}

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
						{#each fectchedCourseChapters[chapter] as asset, i}
							{@const lastAsset = fectchedCourseChapters[chapter].length - 1 == i}

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
												<Icons.Dot weight="fill" class="size-5" />
												<span class="text-success font-bold"> completed </span>
											{:else if asset.assetType === 'video' && asset.videoPos > 0}
												<Icons.Dot weight="fill" class="size-5" />
												<span class="text-secondary"> in-progress </span>
											{/if}

											<!-- Attachments -->
											{#if asset.attachments && asset.attachments.length > 0}
												<Icons.Dot weight="fill" class="size-5" />

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

															<Icons.CaretRight
																class="size-3 duration-200 group-data-[state=open]:rotate-90"
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
																href={GetBackendUrl(ATTACHMENT_API) +
																	'/' +
																	attachment.id +
																	'/serve'}
																download
															>
																<div class="flex flex-row gap-1.5">
																	<span class="shrink-0">{i + 1}.</span>
																	<span class="grow">{attachment.title}</span>
																</div>

																<Icons.Download class="flex size-3 shrink-0" />
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
											href="/course?id={courseId}&a={asset.id}"
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
	{:catch error}
		<Err class="min-h-[5rem]" errorMessage={error} />
	{/await}
</div>
