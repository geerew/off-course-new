<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import type { Course } from '$lib/types/models';
	import { BookText, FolderSearch, MoreHorizontal, Play, Trash2 } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let course: Course;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();
</script>

<DropdownMenu.Root disableFocusFirstItem={true}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			variant="ghost"
			class="text-muted-foreground data-[highlighted]:text-foreground data-[state=open]:text-foreground data-[state=open]:bg-muted h-auto p-1.5"
			builders={[builder]}
		>
			<MoreHorizontal class="h-4 w-4" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="flex flex-col text-sm" align="end" fitViewport={true}>
		<DropdownMenu.Item
			class="data-[highlighted]:bg-primary data-[highlighted]:text-primary-foreground cursor-pointer gap-2.5"
			href="/course?id={course.id}"
		>
			<Play class="h-4 w-4" />
			Open
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />

		<DropdownMenu.Item
			class="cursor-pointer gap-2.5"
			href="/settings/courses/details?id={course.id}"
		>
			<BookText class="h-4 w-4" />
			Details
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="cursor-pointer gap-2.5"
			disabled={course.scanStatus !== ''}
			on:click={() => {
				dispatch('scan', { id: course.id });
			}}
		>
			<FolderSearch class="h-4 w-4" />
			Scan
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />

		<DropdownMenu.Item
			class="text-destructive data-[highlighted]:bg-destructive data-[highlighted]:text-destructive-foreground cursor-pointer gap-2.5"
			on:click={() => {
				dispatch('delete', { id: course.id });
			}}
		>
			<Trash2 class="h-4 w-4" />
			Delete
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
