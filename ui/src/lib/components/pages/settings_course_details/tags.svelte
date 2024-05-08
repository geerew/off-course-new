<script lang="ts">
	import EditCourseTags from '$components/dialogs/edit-course-tags.svelte';
	import { Loading } from '$components/generic';
	import { Badge } from '$components/ui/badge';
	import { GetCourseTags } from '$lib/api';
	import type { CourseTag } from '$lib/types/models';
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
	const sortTags = (tags: CourseTag[]) => {
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
	async function getTags(courseId: string): Promise<CourseTag[]> {
		tagsRefresh = false;
		try {
			const response = await GetCourseTags(courseId);
			if (!response) return [];
			return sortTags(response);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	}

	// ----------------------
	// Exports
	// ----------------------

	$: if (tagsRefresh) {
		tags = getTags(courseId);
	}

	// ----------------------
	// Variables
	// ----------------------

	// Holds the tags for this course
	let tags = getTags(courseId);
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex flex-row items-center gap-2">
		<span class="font-bold">Tags</span>
		{#await tags then data}
			<EditCourseTags
				{courseId}
				existingTags={data}
				on:updated={() => {
					tags = getTags(courseId);
				}}
			/>
		{/await}
	</div>

	{#await tags}
		<Loading class="min-h-5 w-full p-1 py-2" loaderClass="size-6" />
	{:then data}
		<!-- Tags -->
		<div class="flex flex-row flex-wrap gap-2.5">
			{#each data as tag}
				<div class="flex flex-row">
					<!-- Tag -->
					<Badge
						class="bg-alt-1 hover:bg-alt-1 text-foreground min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm"
					>
						{tag.tag}
					</Badge>
				</div>
			{/each}
		</div>
	{:catch _}
		<Badge
			class="bg-destructive text-destructive-foreground hover:bg-destructive items-center gap-1.5 rounded-sm"
		>
			Failed to load tags
		</Badge>
	{/await}
</div>
