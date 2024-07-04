<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import type { Tag } from '$lib/types/models';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let tag: Tag;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();
</script>

<DropdownMenu.Root disableFocusFirstItem={true}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			variant="ghost"
			class="text-muted-foreground hover:bg-alt-1/60 data-[state=open]:text-foreground data-[state=open]:bg-alt-1/60 h-auto p-1.5"
			builders={[builder]}
		>
			<Icons.MoreHorizontal weight="fill" class="size-4" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="flex flex-col text-sm" align="end" fitViewport={true}>
		<DropdownMenu.Item
			class="cursor-pointer gap-2.5"
			on:click={() => {
				dispatch('rename', { id: tag.id });
			}}
		>
			<Icons.Edit class="size-4" />
			Rename
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />

		<DropdownMenu.Item
			class="text-destructive data-[highlighted]:bg-destructive data-[highlighted]:text-destructive-foreground cursor-pointer gap-2.5"
			on:click={() => {
				dispatch('delete', { id: tag.id });
			}}
		>
			<Icons.Trash class="size-4" />
			Delete
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
