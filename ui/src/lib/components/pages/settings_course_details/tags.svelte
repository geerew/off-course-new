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
	export let refresh: boolean;

	// ----------------------
	// Variables
	// ----------------------

	let fetchedTags: CourseTag[] = [];
	let tags = getTags(courseId);

	// ----------------------
	// Functions
	// ----------------------

	// Sorter for tags
	function sortTags(tags: CourseTag[]) {
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
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Gets the tags for this course. During a refresh there is a small delay to prevent flickering
	async function getTags(courseId: string): Promise<boolean> {
		const refreshPromise = new Promise((resolve) => setTimeout(resolve, refresh ? 500 : 0));

		refresh = false;
		try {
			let response: CourseTag[];

			await Promise.all([(response = await GetCourseTags(courseId)), refreshPromise]);

			fetchedTags = sortTags(response);
			return true;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	}

	// ----------------------
	// Exports
	// ----------------------

	$: if (refresh) {
		tags = getTags(courseId);
	}
</script>

<div class="flex flex-col gap-3.5">
	<div class="flex flex-row items-center gap-2">
		<span class="font-bold">Tags</span>
		{#await tags then _}
			<EditCourseTags
				{courseId}
				existingTags={fetchedTags}
				on:updated={() => {
					tags = getTags(courseId);
				}}
			/>
		{/await}
	</div>

	{#await tags}
		<Loading class="min-h-1.5 w-full p-0" loaderClass="size-5" />
	{:then _}
		<!-- Tags -->
		<div class="flex flex-row flex-wrap gap-2.5">
			{#each fetchedTags as tag}
				<div class="flex flex-row">
					<!-- Tag -->
					<Badge
						class="min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm bg-alt-1 text-foreground hover:bg-alt-1"
					>
						{tag.tag}
					</Badge>
				</div>
			{/each}
		</div>
	{:catch _}
		<Badge
			class="items-center gap-1.5 rounded-sm bg-destructive text-destructive-foreground hover:bg-destructive"
		>
			Failed to load tags
		</Badge>
	{/await}
</div>
