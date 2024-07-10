<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let tagId: string;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();
</script>

<DropdownMenu.Root disableFocusFirstItem={true}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			variant="ghost"
			class="h-auto p-1.5 text-muted-foreground hover:bg-alt-1/60 data-[state=open]:bg-alt-1/60 data-[state=open]:text-foreground"
			builders={[builder]}
		>
			<Icons.MoreHorizontal weight="fill" class="size-4" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="flex flex-col text-sm" align="end" fitViewport={true}>
		<DropdownMenu.Item
			class="cursor-pointer gap-2.5"
			on:click={() => {
				dispatch('rename', tagId);
			}}
		>
			<Icons.Edit class="size-4" />
			Rename
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="my-1 -ml-1 -mr-1 block h-px bg-muted" />

		<DropdownMenu.Item
			class="cursor-pointer gap-2.5 text-destructive data-[highlighted]:bg-destructive data-[highlighted]:text-destructive-foreground"
			on:click={() => {
				dispatch('delete', tagId);
			}}
		>
			<Icons.Trash class="size-4" />
			Delete
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
