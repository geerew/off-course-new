<script lang="ts">
	import { Icons } from '$components/icons';
	import type { ClassName } from '$lib/types/general';
	import type { PaginationData } from '$lib/types/pagination';
	import { createSelect, type SelectOption } from '@melt-ui/svelte';
	import { createEventDispatcher } from 'svelte';
	import { writable } from 'svelte/store';
	import { fly } from 'svelte/transition';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------

	// The number of items per page
	export let perPage: PaginationData['perPage'];

	// The number of items per page
	export let perPages: PaginationData['perPages'];

	export { className as class };

	// ----------------------
	// Variables
	// ----------------------

	const selectedPerPage = writable<SelectOption<number>>({
		value: perPage,
		label: String(perPage)
	});

	const {
		elements: { trigger, menu, option },
		states: { selectedLabel, open },
		helpers: { isSelected }
	} = createSelect({
		forceVisible: true,
		selected: selectedPerPage
	});

	const dispatch = createEventDispatcher();

	// ----------------------
	// Reactive
	// ----------------------

	$: if ($selectedPerPage.value !== perPage) {
		perPage = $selectedPerPage.value;
		dispatch('change', $selectedPerPage.value);
	}
</script>

<div class={className}>
	<button
		class="bg-background hover:border-foreground relative inline-flex h-full min-w-[8rem] items-center justify-between whitespace-nowrap rounded border
		px-3 py-1.5 text-center text-sm duration-200"
		{...$trigger}
		use:trigger
		aria-label="Food"
	>
		{$selectedLabel || ''}
		<Icons.chevronRight class="h-4 w-4 duration-200 {$open ? 'rotate-90' : ''}" />
	</button>

	{#if $open}
		<div
			class="bg-background z-10 flex max-h-[300px] flex-col gap-2 overflow-y-auto rounded-md border p-1 focus:!ring-0"
			{...$menu}
			use:menu
			transition:fly={{ duration: 100, y: -5 }}
		>
			{#each perPages as item}
				<div
					class="text-foreground/60 focus:bg-accent-1/50 data-[selected]:text-primary focus:text-foreground relative cursor-pointer rounded-md py-1.5 pl-8 pr-4 outline-none data-[selected]:font-semibold"
					{...$option({ value: item, label: String(item) })}
					use:option
				>
					{#if $isSelected(item)}
						<Icons.check
							class="text-primary absolute left-2 top-1/2 h-4 w-4 -translate-y-[calc(50%-1px)]"
						/>
					{/if}
					{item}
				</div>
			{/each}
		</div>
	{/if}
</div>
