<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import type { Writable } from 'svelte/store';

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;
	export let scanning: Writable<boolean>;

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
		<DropdownMenu.Item class="cursor-pointer gap-2.5" href="/course?id={courseId}">
			<Icons.Play class="size-4" />
			Open
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />

		<DropdownMenu.Item
			class="cursor-pointer gap-2.5"
			href="/settings/courses/details?id={courseId}"
		>
			<Icons.Info class="size-4" />
			Details
		</DropdownMenu.Item>

		<!-- disabled={$scanning} -->
		<DropdownMenu.Item
			class={cn('cursor-pointer gap-2.5', $scanning && 'pointer-events-none opacity-50')}
			on:click={() => {
				if ($scanning) return;
				dispatch('scan', { id: courseId });
			}}
		>
			<Icons.Scan class="size-4 stroke-[1.5]" />
			Scan
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />

		<DropdownMenu.Item
			class="text-destructive data-[highlighted]:bg-destructive data-[highlighted]:text-destructive-foreground cursor-pointer gap-2.5"
			on:click={() => {
				dispatch('delete', { id: courseId });
			}}
		>
			<Icons.Trash class="size-4" />
			Delete
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
