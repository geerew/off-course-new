<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import type { Writable } from 'svelte/store';

	// -------------------
	// Exports
	// -------------------
	export let columns: Array<{ id: string; label: string }>;
	export let columnStore: Writable<Array<string>>;
	export let disabled: boolean = false;

	// -------------------
	// Variables
	// -------------------
	const dispatch = createEventDispatcher<Record<'changed', string[]>>();
</script>

<DropdownMenu.Root closeOnItemClick={false}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button variant="outline" {disabled} class="group flex h-8 px-2" builders={[builder]}>
			<div class="flex items-center gap-1.5 pr-3">
				<Icons.EyeSlash class="size-4" />
				<span>Columns</span>
			</div>

			<Icons.CaretRight class="size-3 duration-200 group-data-[state=open]:rotate-90" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="flex flex-col text-sm" align="end" fitViewport={true}>
		{#each columns as col}
			<DropdownMenu.Item
				class="hover:bg-muted relative cursor-pointer select-none rounded-md p-1 pl-8 pr-4 text-left focus:z-10"
				on:click={() => {
					if ($columnStore.includes(col.id)) {
						columnStore.update((store) => store.filter((id) => id !== col.id));
					} else {
						columnStore.update((store) => [...store, col.id]);
					}

					dispatch('changed', $columnStore);
				}}
			>
				<div
					class={cn(
						'absolute left-2 top-1/2 z-20 -translate-y-1/2',
						!$columnStore.includes(col.id) ? 'block' : 'hidden'
					)}
				>
					<Icons.Check class="size-3.5" />
				</div>

				<span class={$columnStore.includes(col.id) ? 'text-muted-foreground' : ''}>
					{col.label}
				</span>
			</DropdownMenu.Item>
		{/each}
	</DropdownMenu.Content>
</DropdownMenu.Root>
