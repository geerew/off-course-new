<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { CourseProgress } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { ChevronRight, Loader2 } from 'lucide-svelte';
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
			class="data-[state=open]:border-alt-1/100 border-alt-1/60 hover:border-alt-1/100 group h-auto w-32 items-center justify-between gap-2.5 border px-2.5 text-xs hover:bg-inherit"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-1.5">
				<Loader2 class={cn('size-3', progress && 'text-primary')} />
				<span>Progress </span>
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
			{#each Object.values(CourseProgress) as cp}
				<DropdownMenu.CheckboxItem
					class="data-[highlighted]:bg-alt-1/40 cursor-pointer"
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
