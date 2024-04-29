<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as Popover from '$components/ui/popover';
	import { cn } from '$lib/utils';
	import { ArrowDownUp, ChevronDown, ChevronRight, ChevronUp } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import type { WritableSortKeys } from 'svelte-headless-table/plugins';

	// -------------------
	// Exports
	// -------------------
	export let columns: Array<{ id: string; label: string }>;
	export let sortedColumn: WritableSortKeys;
	export let disabled: boolean = false;

	// -------------------
	// Variables
	// -------------------
	const dispatch = createEventDispatcher();

	let isOpen = false;
</script>

<Popover.Root bind:open={isOpen}>
	<Popover.Trigger asChild let:builder>
		<Button variant="outline" {disabled} class="flex h-8 gap-1.5" builders={[builder]}>
			<ArrowDownUp class="h-4 w-4" />
			<span>Sort</span>
			<ChevronRight class={cn('h-4 w-4 duration-200', isOpen && 'rotate-90')} />
		</Button>
	</Popover.Trigger>

	<!-- z-50 min-w-[8rem] rounded-md border bg-popover p-1 text-popover-foreground shadow-md
	focus:outline-none flex flex-col text-sm z-50 w-72 rounded-md border bg-popover p-4
	text-popover-foreground shadow-md outline-none flex flex-col text-sm -->

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
							dispatch('changed');
						}}
					>
						<ChevronUp
							class={cn(
								'text-muted-foreground h-5 w-5 duration-150',
								isAscSorted ? 'text-primary' : 'hover:text-foreground'
							)}
						/>
					</button>

					<button
						on:click={() => {
							if (isDescSorted) return;
							sortedColumn.set([{ id: column.id, order: 'desc' }]);
							dispatch('changed');
						}}
					>
						<ChevronDown
							class={cn(
								'text-muted-foreground h-5 w-5 duration-150',
								isDescSorted ? 'text-primary' : 'hover:text-foreground'
							)}
						/>
					</button>
				</div>
			</div>
		{/each}
	</Popover.Content>
</Popover.Root>
