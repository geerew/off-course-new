<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { CourseProgress } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let progress: CourseProgress | undefined;

	// ----------------------
	// Variables
	// ----------------------

	// As the progress change, dispatch an event
	const dispatchEvent = createEventDispatcher();
</script>

<DropdownMenu.Root typeahead={false}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="ghost"
			class="group h-auto w-32 items-center justify-between gap-2.5 border border-alt-1/60 px-2.5 text-xs hover:border-alt-1/100 hover:bg-inherit data-[state=open]:border-alt-1/100"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-2">
				<Icons.Hourglass class={cn('size-3', progress && 'text-primary')} />
				<span>Progress </span>
			</div>

			<Icons.CaretRight class="size-3 duration-200 group-data-[state=open]:rotate-90" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content
		class="flex w-48 flex-col border-alt-1/60 bg-muted text-foreground"
		fitViewport={true}
		align="start"
	>
		<div class="max-h-40 overflow-y-scroll">
			{#each Object.values(CourseProgress) as cp}
				<DropdownMenu.CheckboxItem
					class="cursor-pointer data-[highlighted]:bg-alt-1/40"
					checked={progress === cp}
					onCheckedChange={(checked) => {
						if (checked) {
							progress = cp;
						} else {
							progress = undefined;
						}

						dispatchEvent('change');
					}}
				>
					{cp}
				</DropdownMenu.CheckboxItem>
			{/each}
		</div>
	</DropdownMenu.Content>
</DropdownMenu.Root>
