<script lang="ts">
	import { Loading } from '$components';
	import { Badge } from '$components/ui/badge';
	import Button from '$components/ui/button/button.svelte';
	import * as Tooltip from '$components/ui/tooltip';
	import { AddCourseTag, DeleteCourseTag, GetCourseTags } from '$lib/api';
	import type { Tag } from '$lib/types/models';
	import { cn, flyAndScale } from '$lib/utils';
	import { Pencil, RotateCcw, X } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;
	export let tagsRefresh: boolean;

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
	const getTags = async (courseId: string): Promise<Tag[]> => {
		tagsRefresh = false;

		try {
			const response = await GetCourseTags(courseId);
			if (!response) return [];
			return sortTags(response);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Use:action for inputting tags
	const tagInput = (node: HTMLInputElement, tags: Tag[]) => {
		function handleInput(e: KeyboardEvent) {
			if (e.key === 'Enter') {
				e.preventDefault();

				if (!node.value) return;

				if (
					tags.find((t) => t.tag.toLowerCase() === node.value.toLowerCase()) ||
					toAdd.find((t) => t.toLowerCase() === node.value.toLowerCase())
				) {
					toast.error('Tag already exists');
					return;
				}

				toAdd = [...toAdd, node.value];
				node.value = '';
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
		try {
			const response = await AddCourseTag(courseId, tag);
			if (!response) throw new Error('Failed to add tag');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Delete a tag for this course
	const deleteCourseTag = async (courseId: string, tagId: string) => {
		try {
			const response = await DeleteCourseTag(courseId, tagId);
			if (!response) throw new Error('Failed to delete tag');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	};

	// ----------------------
	// Variables
	// ----------------------

	// Holds the tags for this course
	let tags = getTags(courseId);

	let toDelete: string[] = [];
	let toAdd: string[] = [];

	// True when processing tags
	let processingTags = false;

	// True when editing tags
	let editMode = false;

	// Used to get focus when editing starts
	let tagInputEl: HTMLInputElement;

	// ----------------------
	// Reactive
	// ----------------------

	// Update tags when `getTags` is set to true
	$: if (tagsRefresh) {
		tags = getTags(courseId);
	}
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex flex-row items-center gap-2">
		<span class="font-bold">Tags</span>
		{#await tags then _}
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
			{/if}
		{/await}
	</div>

	{#await tags}
		<Loading class="h-6 w-6" />
	{:then data}
		<!-- Tags -->
		<div class="flex flex-row flex-wrap gap-2.5">
			{#each data as tag}
				<div class="flex flex-row">
					<!-- Tag -->
					<Badge
						class={cn(
							'bg-alt-1 hover:bg-alt-1 text-foreground min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm',
							editMode && 'rounded-r-none',
							editMode &&
								toDelete.includes(tag.tag) &&
								'bg-destructive text-destructive-foreground hover:bg-destructive'
						)}
					>
						{tag.tag}
					</Badge>
					{#if editMode}
						{#if !toDelete.includes(tag.tag)}
							<!-- Delete button -->
							<Button
								class={cn(
									'bg-alt-1 hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5',
									toAdd.includes(tag.tag) && 'bg-success text-success-foreground'
								)}
								on:click={() => {
									toDelete = [...toDelete, tag.tag];
								}}
							>
								<X class="size-3" />
							</Button>
						{:else}
							<!-- Undo delete button -->
							<Button
								class="bg-destructive hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5 hover:brightness-110"
								on:click={() => {
									toDelete = toDelete.filter((t) => t !== tag.tag);
								}}
							>
								<RotateCcw class="size-3" />
							</Button>
						{/if}
					{/if}
				</div>
			{/each}

			{#each toAdd as tag}
				<div class="flex flex-row">
					<!-- Tag -->
					<Badge
						class={cn(
							'bg-success text-success-foreground hover:bg-success min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none'
						)}
					>
						{tag}
					</Badge>

					<!-- Delete button -->
					<Button
						class={cn(
							'hover:bg-destructive bg-success text-success-foreground inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5'
						)}
						on:click={() => {
							// When its a newly added tag, just delete it completely
							toAdd = toAdd.filter((t) => t !== tag);
						}}
					>
						<X class="size-3" />
					</Button>
				</div>
			{/each}
			<input
				bind:this={tagInputEl}
				use:tagInput={data}
				contenteditable="true"
				hidden={!editMode}
				role="textbox"
				placeholder="Add tag..."
				class="bg-muted text-foreground border-alt-1 focus:border-foreground w-20 min-w-[8rem] rounded-sm border px-2 py-0 text-sm outline-none duration-200 focus:!ring-0 data-[invalid]:text-red-500"
			/>
		</div>

		<!-- Fixed bar for cancel/save -->
		{#if editMode}
			<div class="bg-muted border-alt-1 fixed bottom-0 left-0 h-16 w-full border-t">
				<div class="container flex h-full flex-row gap-4">
					<div class="flex w-full items-center justify-center gap-6">
						<!-- Cancel -->
						<Button
							variant="ghost"
							class="border-muted-foreground/60 hover:bg-alt-1/60 hover:border-alt-1/60 h-8 w-20 gap-2 rounded border"
							on:click={() => {
								editMode = false;
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

								// Add and delete tags
								await Promise.all(toAdd.map((tag) => addCourseTag(courseId, tag)));
								await Promise.all(toDelete.map((tagId) => deleteCourseTag(courseId, tagId)));

								toAdd = [];
								toDelete = [];

								processingTags = false;
								editMode = false;

								tagsRefresh = true;
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
	{:catch _}
		<Badge
			class="bg-destructive text-destructive-foreground hover:bg-destructive items-center gap-1.5 rounded-sm"
		>
			Failed to load tags
		</Badge>
	{/await}
</div>
