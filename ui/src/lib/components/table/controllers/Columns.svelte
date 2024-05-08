<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { cn } from '$lib/utils';
	import { Check, ChevronRight, EyeOff } from 'lucide-svelte';
	import type { Writable } from 'svelte/store';

	// -------------------
	// Exports
	// -------------------
	export let columns: Array<{ id: string; label: string }>;
	export let columnStore: Writable<Array<string>>;
	export let disabled: boolean = false;
</script>

<DropdownMenu.Root closeOnItemClick={false}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button variant="outline" {disabled} class="group flex h-8 gap-2.5" builders={[builder]}>
			<EyeOff class="size-4" />
			<span>Columns</span>
			<ChevronRight class="size-4 duration-200 group-data-[state=open]:rotate-90" />
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
				}}
			>
				<div
					class={cn(
						'absolute left-2 top-1/2 z-20 -translate-y-1/2',
						!$columnStore.includes(col.id) ? 'block' : 'hidden'
					)}
				>
					<Check class="size-4" />
				</div>

				<span class={$columnStore.includes(col.id) ? 'text-muted-foreground' : ''}>
					{col.label}
				</span>
			</DropdownMenu.Item>
		{/each}
	</DropdownMenu.Content>
</DropdownMenu.Root>
