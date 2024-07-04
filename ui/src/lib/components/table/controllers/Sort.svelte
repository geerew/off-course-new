<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Popover from '$components/ui/popover';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import type { SortKey, WritableSortKeys } from 'svelte-headless-table/plugins';

	// -------------------
	// Exports
	// -------------------
	export let columns: Array<{ id: string; label: string }>;
	export let sortedColumn: WritableSortKeys;
	export let disabled: boolean = false;

	// -------------------
	// Variables
	// -------------------
	const dispatch = createEventDispatcher<Record<'changed', SortKey>>();

	let isOpen = false;
</script>

<Popover.Root bind:open={isOpen}>
	<Popover.Trigger asChild let:builder>
		<Button variant="outline" {disabled} class="flex h-8 px-2" builders={[builder]}>
			<div class="flex items-center gap-1.5 pr-3">
				<Icons.ArrowsDownUp class="size-4" />
				<span>Sort</span>
			</div>

			<Icons.CaretRight class={cn('size-3 duration-200', isOpen && 'rotate-90')} />
		</Button>
	</Popover.Trigger>

	<Popover.Content
		class="flex w-auto min-w-[8rem] flex-col text-sm"
		align="end"
		sideOffset={4}
		fitViewport={true}
	>
		{#each columns as column}
			{@const isAscSorted =
				$sortedColumn.length >= 1 &&
				$sortedColumn[0].order === 'asc' &&
				$sortedColumn[0].id === column.id}
			{@const isDescSorted =
				$sortedColumn.length >= 1 &&
				$sortedColumn[0].order === 'desc' &&
				$sortedColumn[0].id === column.id}

			<div
				class={cn(
					'hover:bg-muted relative flex select-none items-center justify-between gap-2.5 rounded-md px-2 py-1 focus:z-10'
				)}
			>
				<span>{column.label}</span>

				<div class="flex gap-0.5">
					<button
						on:click={() => {
							if (isAscSorted) return;
							sortedColumn.set([{ id: column.id, order: 'asc' }]);
							dispatch('changed', { id: column.id, order: 'asc' });
						}}
					>
						<Icons.CaretUp
							class={cn(
								'text-muted-foreground size-4 duration-150',
								isAscSorted ? 'text-primary' : 'hover:text-foreground'
							)}
						/>
					</button>

					<button
						on:click={() => {
							if (isDescSorted) return;
							sortedColumn.set([{ id: column.id, order: 'desc' }]);
							dispatch('changed', { id: column.id, order: 'desc' });
						}}
					>
						<Icons.CaretDown
							class={cn(
								'text-muted-foreground size-4 duration-150',
								isDescSorted ? 'text-primary' : 'hover:text-foreground'
							)}
						/>
					</button>
				</div>
			</div>
		{/each}
	</Popover.Content>
</Popover.Root>
