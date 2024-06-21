<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { LogLevel } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { ChevronRight, ScrollText } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let filterLevels: LogLevel[] = [];

	// ----------------------
	// Variables
	// ----------------------

	// As the progress change, dispatch an event
	const dispatchEvent = createEventDispatcher();
</script>

<DropdownMenu.Root closeOnItemClick={false} typeahead={false}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="ghost"
			class="data-[state=open]:border-alt-1/100 border-alt-1/60 hover:border-alt-1/100 group h-auto w-36 items-center justify-between gap-2.5 border px-2.5 text-xs hover:bg-inherit"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-2">
				<ScrollText class={cn('size-3', Object.keys(filterLevels).length > 0 && 'text-primary')} />
				<span>Log Level</span>
			</div>

			<ChevronRight class="size-3 duration-200 group-data-[state=open]:rotate-90" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content
		class="bg-muted text-foreground border-alt-1/60 flex w-48 flex-col"
		fitViewport={true}
		align="start"
	>
		<div class="max-h-40 overflow-y-scroll">
			{#each Object.values(LogLevel) as l}
				<DropdownMenu.CheckboxItem
					class="data-[highlighted]:bg-alt-1/40 cursor-pointer"
					checked={filterLevels.find((level) => level === l) ? true : false}
					onCheckedChange={(checked) => {
						if (checked) {
							filterLevels = [...filterLevels, l];
						} else {
							filterLevels = filterLevels.filter((level) => level !== l);
						}

						dispatchEvent('change', l);
					}}
				>
					{l}
				</DropdownMenu.CheckboxItem>
			{/each}
		</div>
	</DropdownMenu.Content>
</DropdownMenu.Root>
