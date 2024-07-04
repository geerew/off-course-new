<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { LogLevel } from '$lib/types/models';
	import { cn } from '$lib/utils';
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
			variant="outline"
			class="border-alt-1/60 group h-auto w-36 justify-between gap-2.5 border px-2"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-2">
				<Icons.Scroll
					class={cn('size-4', Object.keys(filterLevels).length > 0 && 'text-primary')}
				/>
				<span>Log Level</span>
			</div>

			<Icons.CaretRight class="size-3.5 duration-200 group-data-[state=open]:rotate-90" />
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
