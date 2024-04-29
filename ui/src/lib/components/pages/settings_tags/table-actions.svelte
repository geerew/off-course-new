<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { BoxSelect, ChevronRight, Trash2, X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import type { Writable } from 'svelte/store';

	// ----------------------
	// Exports
	// ----------------------
	export let selectedTagsCount: Writable<number>;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();
</script>

<DropdownMenu.Root disableFocusFirstItem={true}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button variant="outline" class="group flex h-8 gap-1.5" builders={[builder]}>
			<span>Actions</span>
			<ChevronRight class="size-4 duration-200 group-data-[state=open]:rotate-90" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="flex flex-col text-sm" align="start" fitViewport={true}>
		<DropdownMenu.Item
			disabled={$selectedTagsCount === 0}
			class="h-6 cursor-pointer gap-2.5 py-0"
			on:click={() => {
				dispatch('deselect');
			}}
		>
			<div class="relative size-4">
				<BoxSelect class="absolute size-4" />
				<X class="absolute left-1/2 top-1/2 size-3 -translate-x-1/2 -translate-y-1/2" />
			</div>
			<span>Deselect All</span>
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />

		<DropdownMenu.Item
			disabled={$selectedTagsCount === 0}
			class="text-destructive data-[highlighted]:bg-destructive data-[highlighted]:text-destructive-foreground h-6 cursor-pointer gap-2.5 py-0"
			on:click={() => {
				dispatch('delete');
			}}
		>
			<Trash2 class="size-4" />
			Delete
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
