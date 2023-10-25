<script lang="ts">
	import { Icons } from '$components/icons';
	import type { ClassName } from '$lib/types/general';
	import type { PaginationParams } from '$lib/types/pagination';
	import { cn } from '$lib/utils/general';
	import { createPagination } from '@melt-ui/svelte';
	import { createEventDispatcher } from 'svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------

	// The total number of items to paginate
	export let pagination: PaginationParams;
	export { className as class };

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();

	const {
		elements: { root, pageTrigger, prevButton, nextButton },
		states: { pages, page },
		options: { count: countOption, perPage: perPageOption }
	} = createPagination({
		count: pagination.totalItems,
		perPage: pagination.perPage,
		defaultPage: pagination.page
	});

	// When the pagination.page changes, update the page store
	$: page.set(pagination.page);

	// Holds the current page. This stops reactivity from triggering when the component is first
	// loaded
	let currentPage = $page;

	// ----------------------
	// Reactive
	// ----------------------

	// Update the per page option when the user selects a new option from the <PerPage /> component
	$: perPageOption.set(pagination.perPage);

	// Update the total number of items. This may be due to courses being added or removed
	$: countOption.set(pagination.totalItems);

	// Update the current page when the user selects a new page
	$: if (currentPage !== $page) {
		currentPage = $page;
		pagination = { ...pagination, page: $page };
		dispatch('change');
	}
</script>

{#if pagination.totalItems > pagination.perPage}
	<nav class={className} aria-label="pagination" {...$root} use:root>
		<div class="flex items-center">
			<!-- Previous -->
			<button class="previous" {...$prevButton} use:prevButton>
				<Icons.chevronLeft class="h-4 w-4" />
				<span>Previous</span>
			</button>

			<!-- Pages -->
			{#each $pages as page, i (page.key)}
				{#if page.type === 'ellipsis'}
					<span class="border-y px-2 py-2.5">
						<Icons.moreHorizontal class="text-foreground-muted h-4 w-4" />
					</span>
				{:else}
					<button
						class={cn('page', i === 0 && 'first', i === $pages.length - 1 && 'last')}
						{...$pageTrigger(page)}
						use:pageTrigger
					>
						{page.value}
					</button>
				{/if}
			{/each}

			<!-- Next -->
			<button class="next" {...$nextButton} use:nextButton>
				<span>Next</span>
				<Icons.chevronRight class="h-4 w-4" />
			</button>
		</div>
	</nav>
{/if}

<style lang="postcss">
	button {
		@apply relative inline-flex items-center justify-center whitespace-nowrap rounded border-y px-3 py-2 text-center text-sm;
		@apply enabled:hover:bg-accent-1/50 disabled:text-foreground-muted;
	}

	.previous {
		@apply rounded-r-none border-l;
		@apply enabled:hover:after:bg-border enabled:hover:after:absolute enabled:hover:after:right-0 enabled:hover:after:z-10 enabled:hover:after:h-full enabled:hover:after:w-px;
	}

	.next {
		@apply rounded-l-none border-r;
		@apply enabled:hover:before:bg-border enabled:hover:before:absolute enabled:hover:before:left-0 enabled:hover:before:z-10 enabled:hover:before:h-full enabled:hover:before:w-px;
	}

	.page {
		@apply rounded-none !px-3;
		@apply data-[selected]:border-foreground data-[selected]:border-x;
	}

	.page.first:not([data-selected]) {
		@apply before:bg-border before:absolute before:-left-px before:h-6 before:w-px;
		@apply hover:before:bg-border hover:before:h-full;
	}

	.page.last:not([data-selected]) {
		@apply after:bg-border after:absolute after:-right-px after:h-6 after:w-px;
		@apply hover:after:bg-border hover:after:h-full;
	}
</style>
