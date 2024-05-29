<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { ChevronRight, Tag } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let level: number | undefined = -4;

	// ----------------------
	// Variables
	// ----------------------

	// As the progress change, dispatch an event
	const dispatchEvent = createEventDispatcher<Record<'change', number>>();

	const levels: Record<string, number> = {
		DEBUG: -4,
		INFO: 0,
		WARNING: 4,
		ERROR: 8
	};
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
			<div class="flex items-center gap-1.5">
				<Tag class="size-3" />
				<span>Min Log Level</span>
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
			{#each Object.keys(levels) as l}
				<DropdownMenu.CheckboxItem
					class="data-[highlighted]:bg-alt-1/40 cursor-pointer"
					checked={level === levels[l]}
					on:click={(e) => {
						if (levels[l] === level) {
							e.preventDefault();
						}
					}}
					onCheckedChange={(checked) => {
						if (checked) {
							level = levels[l];
						} else {
							level = undefined;
						}

						dispatchEvent('change', levels[l]);
					}}
				>
					{l}
				</DropdownMenu.CheckboxItem>
			{/each}
		</div>
	</DropdownMenu.Content>
</DropdownMenu.Root>
