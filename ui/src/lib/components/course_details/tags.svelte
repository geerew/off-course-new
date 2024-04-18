<script lang="ts">
	import { Loading } from '$components';
	import { Badge } from '$components/ui/badge';
	import Button from '$components/ui/button/button.svelte';
	import * as Tooltip from '$components/ui/tooltip';
	import { AddCourseTag, DeleteCourseTag, ErrorMessage, GetCourseTags } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { Tag } from '$lib/types/models';
	import { cn, flyAndScale } from '$lib/utils';
	import { isBrowser } from '@melt-ui/svelte/internal/helpers';
	import { Pencil, RotateCcw, X } from 'lucide-svelte';
	import { onMount } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;

	// ----------------------
	// Variables
	// ----------------------

	// Holds the tags for this course
	let tags: Tag[];

	let toDelete: Tag[] = [];

	let toAdd: string[] = [];

	// True while the course tags are loading
	let loadingCourseTags = true;

	// True when there was an error getting the assets
	let gotCourseTagsError = false;

	// True when processing tags
	let processingTags = false;

	// True when editing tags
	let editMode = false;

	let tagInputEl: HTMLInputElement;

	// ----------------------
	// Functions
	// ----------------------

	// Sorter for tags
	const sortTags = (tags: Tag[]) => {
		tags.sort((a, b) => {
			if (a.tag.toLowerCase() < b.tag.toLowerCase()) {
				return -1;
			}
			if (a.tag.toLowerCase() > b.tag.toLowerCase()) {
				return 1;
			}
			return 0;
		});

		return tags;
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Gets the tags for this course
	const getCourseTags = async (courseId: string) => {
		if (!isBrowser) return false;

		return await GetCourseTags(courseId)
			.then(async (resp) => {
				if (!resp) return false;
				tags = sortTags(resp);
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

	// Use:action for inputting tags
	const tagInput = (node: HTMLInputElement) => {
		function handleInput(e: KeyboardEvent) {
			if (e.key === 'Enter') {
				// Add the tag
				e.preventDefault();
				if (node.value) {
					if (tags.find((t) => t.tag.toLowerCase() === node.value.toLowerCase())) {
						$addToast({
							data: {
								message: `Tag already exists`,
								status: 'error'
							}
						});
						return;
					}

					toAdd = [...toAdd, node.value];
					tags = [...sortTags([...tags, { id: '0', tag: node.value }])];

					node.value = '';
				}
			}
		}

		node.addEventListener('keydown', handleInput);

		return {
			destroy() {
				node.removeEventListener('keydown', handleInput);
			}
		};
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Add a tag for this course
	const addCourseTag = async (courseId: string, tag: string) => {
		return await AddCourseTag(courseId, tag)
			.then(async (resp) => {
				if (!resp) return false;

				// Update the id, createdAt and updatedAt
				const tagIndex = tags.findIndex((t) => t.tag === tag);
				tags[tagIndex] = {
					id: resp.id,
					tag: tag
				};

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

	// Delete a tag for this course
	const deleteCourseTag = async (courseId: string, tagId: string) => {
		return await DeleteCourseTag(courseId, tagId)
			.then(async (resp) => {
				if (!resp) return false;
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

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(async () => {
		if (!(await getCourseTags(courseId))) {
			loadingCourseTags = false;
			gotCourseTagsError = true;
			return;
		}
		loadingCourseTags = false;
	});
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex flex-row items-center gap-2">
		<span class="font-bold">Tags</span>
		{#if !loadingCourseTags && !gotCourseTagsError}
			{#if !editMode}
				<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
					<Tooltip.Trigger asChild let:builder>
						<Button
							builders={[builder]}
							variant="ghost"
							class="text-muted-foreground hover:text-foreground h-auto cursor-pointer px-2.5 py-1"
							on:click={() => {
								editMode = true;
								setTimeout(() => {
									tagInputEl.focus();
								}, 10);
							}}
						>
							<Pencil class="h-4 w-4" />
						</Button>
					</Tooltip.Trigger>

					<Tooltip.Content
						class="bg-foreground text-background select-none rounded-sm border-none px-1.5 py-1 text-xs"
						transition={flyAndScale}
						transitionConfig={{ y: 8, duration: 100 }}
						side="bottom"
					>
						Edit
						<Tooltip.Arrow class="bg-background" />
					</Tooltip.Content>
				</Tooltip.Root>
			{:else}
				<!-- Fixed bar for cancel/save -->
				<div class="bg-muted border-alt-1 fixed bottom-0 left-0 h-16 w-full border-t">
					<div class="container flex h-full flex-row gap-4">
						<div class="flex w-full items-center justify-center gap-6">
							<!-- Cancel -->
							<Button
								variant="ghost"
								class="border-muted-foreground/60 hover:bg-alt-1/60 hover:border-alt-1/60 h-8 w-20 gap-2 rounded border"
								on:click={() => {
									editMode = false;

									// Add back in the deleted tags
									const tmpTags = sortTags([...tags, ...toDelete]);

									// Remove the added tags
									tags = tmpTags.filter((t) => !toAdd.includes(t.tag));

									// Reset
									toDelete = toAdd = [];

									tagInputEl.value = '';
								}}
							>
								Cancel
							</Button>

							<!-- Save -->
							<Button
								disabled={(toAdd.length === 0 && toDelete.length === 0) || processingTags}
								variant="ghost"
								class="bg-success hover:bg-success text-success-foreground h-8 w-20 gap-2 rounded hover:brightness-110"
								on:click={async () => {
									if (toAdd.length === 0 && toDelete.length === 0) {
										editMode = false;
										return;
									}

									processingTags = true;

									// Add tags
									toAdd.forEach(async (t) => {
										await addCourseTag(courseId, t);
									});
									toAdd = [];

									// Delete tags
									toDelete.forEach(async (t) => {
										await deleteCourseTag(courseId, t.id);
									});

									// Remove the deleted tags
									tags = tags.filter((t) => !toDelete.includes(t));
									toDelete = [];

									processingTags = false;
									editMode = false;
								}}
							>
								{#if processingTags}
									<Loading class="h-6 w-6" />
								{:else}
									Save
								{/if}
							</Button>
						</div>
					</div>
				</div>
			{/if}
		{/if}
	</div>
	{#if loadingCourseTags}
		<Loading class="h-6 w-6" />
	{:else if gotCourseTagsError}
		<Badge
			class="bg-destructive text-destructive-foreground hover:bg-destructive items-center gap-1.5 rounded-sm"
		>
			Failed to load tags
		</Badge>
	{:else}
		<!-- Tags -->
		<div class="flex flex-row flex-wrap gap-2.5">
			{#each tags as tag}
				<div class="flex flex-row">
					<!-- Tag -->
					<Badge
						class={cn(
							'bg-alt-1 hover:bg-alt-1 text-foreground min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm',
							editMode && 'rounded-r-none',
							editMode &&
								toDelete.includes(tag) &&
								'bg-destructive text-destructive-foreground hover:bg-destructive',
							toAdd.includes(tag.tag) && 'bg-success text-success-foreground hover:bg-success'
						)}
					>
						{tag.tag}
					</Badge>
					{#if editMode}
						{#if !toDelete.includes(tag)}
							<!-- Delete button -->
							<Button
								class={cn(
									'bg-alt-1 hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5',
									toAdd.includes(tag.tag) && 'bg-success text-success-foreground'
								)}
								on:click={() => {
									if (toAdd.includes(tag.tag)) {
										// When its a newly added tag, just delete it completely
										toAdd = toAdd.filter((t) => t !== tag.tag);
										tags = tags.filter((t) => t.tag !== tag.tag);
									} else {
										// When it's an existing tag, add it to the delete list
										toDelete = [...toDelete, tag];
									}
								}}
							>
								<X class="size-3" />
							</Button>
						{:else}
							<!-- Undo delete button -->
							<Button
								class="bg-destructive hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5 hover:brightness-110"
								on:click={() => {
									toDelete = toDelete.filter((t) => t !== tag);
								}}
							>
								<RotateCcw class="size-3" />
							</Button>
						{/if}
					{/if}
				</div>
			{/each}
			<input
				bind:this={tagInputEl}
				use:tagInput
				contenteditable="true"
				hidden={!editMode}
				role="textbox"
				placeholder="Add tag..."
				class="bg-muted text-foreground border-alt-1 focus:border-foreground w-20 min-w-[8rem] rounded-sm border px-2 py-0 text-sm outline-none duration-200 focus:!ring-0 data-[invalid]:text-red-500"
			/>
		</div>
	{/if}
</div>
