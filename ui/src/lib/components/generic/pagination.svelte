<script lang="ts">
	import { Icons } from '$components/icons';
	import * as Pagination from '$components/ui/pagination';
	import * as Select from '$components/ui/select';
	import type { PaginationParams } from '$lib/types/pagination';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import { mediaQuery } from 'svelte-legos';

	// ----------------------
	// Exports
	// ----------------------

	// The pagination parameters
	export let pagination: PaginationParams;

	// The type of items being paginated(eg. 'course', 'tag'). An 's' will be appended if
	// the totalItems is greater than 1
	export let type: string;

	// Whether to show the per page select
	export let showPerPage = true;

	// ----------------------
	// Variables
	// ----------------------
	const isDesktop = mediaQuery('(min-width: 768px)');

	const dispatch = createEventDispatcher();

	let isOpen = false;

	// ----------------------
	// Reactive
	// ----------------------
	$: siblingCount = $isDesktop ? 1 : 0;

	$: currentPerPages = pagination.perPage;
</script>

{#if pagination.totalItems > 0}
	<div class="grid grid-cols-2 gap-4 pt-5 lg:grid-cols-5">
		<!-- Per pages -->
		<div class="order-2 lg:order-1">
			{#if showPerPage}
				<Select.Root
					bind:open={isOpen}
					preventScroll={false}
					portal={null}
					selected={{ value: pagination.perPage }}
					onSelectedChange={(v) => {
						if (!v || v.value === currentPerPages) return;
						dispatch('perPageChange', v.value);
					}}
				>
					<Select.Trigger class="w-[140px]">
						<Select.Value placeholder={String(pagination.perPage)} />
						<Icons.CaretRight class={cn('size-4 duration-200', isOpen && 'rotate-90')} />
					</Select.Trigger>
					<Select.Content>
						<Select.Group>
							{#each pagination.perPages as pp}
								<Select.Item value={pp} label={String(pp)} class="cursor-pointer">{pp}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
					<Select.Input name="favoriteFruit" />
				</Select.Root>
			{/if}
		</div>

		<!-- Pagination -->
		{#if pagination.totalPages > 1}
			<div class="col-span-2 flex flex-col items-center lg:order-2 lg:col-span-3">
				<Pagination.Root
					count={pagination.totalItems}
					page={pagination.page}
					perPage={pagination.perPage}
					{siblingCount}
					let:pages
					let:currentPage
					onPageChange={(ev) => {
						dispatch('pageChange', ev);
					}}
				>
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.PrevButton>
								<Icons.CaretLeft class="size-4" />
								<span class="hidden sm:block">Previous</span>
							</Pagination.PrevButton>
						</Pagination.Item>
						{#each pages as page (page.key)}
							{#if page.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link {page} isActive={currentPage == page.value}>
										{page.value}
									</Pagination.Link>
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.NextButton>
								<span class="hidden sm:block">Next</span>
								<Icons.CaretRight class="size-4" />
							</Pagination.NextButton>
						</Pagination.Item>
					</Pagination.Content>
				</Pagination.Root>
			</div>
		{/if}

		<!-- Count -->
		<div
			class={cn(
				'order-3 flex items-center justify-end text-sm text-muted-foreground',
				pagination.totalPages === 1 ? 'lg:col-span-4' : undefined,
				!showPerPage && 'hidden lg:flex'
			)}
		>
			{pagination.totalItems}
			{type}{pagination.totalItems > 1 ? 's' : ''}
		</div>
	</div>
{/if}
