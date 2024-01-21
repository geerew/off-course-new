<script lang="ts">
	import * as Pagination from '$components/ui/pagination';
	import * as Select from '$components/ui/select';
	import type { PaginationParams } from '$lib/types/pagination';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { mediaQuery } from 'svelte-legos';

	// ----------------------
	// Exports
	// ----------------------
	export let pagination: PaginationParams;

	// ----------------------
	// Variables
	// ----------------------
	const isDesktop = mediaQuery('(min-width: 768px)');

	const dispatch = createEventDispatcher<Record<'pageChange' | 'perPageChange', number>>();

	// ----------------------
	// Reactive
	// ----------------------
	$: siblingCount = $isDesktop ? 1 : 0;

	$: currentPerPages = pagination.perPage;
</script>

{#if pagination.totalItems > 0}
	<div class="grid grid-cols-2 gap-4 pt-5 md:grid-cols-5">
		<!-- Per pages -->
		<div class="order-2 md:order-1">
			<Select.Root
				portal={null}
				selected={{ value: pagination.perPage }}
				onSelectedChange={(v) => {
					if (!v || v.value === currentPerPages) return;
					dispatch('perPageChange', v.value);
				}}
			>
				<Select.Trigger class="w-[140px]">
					<Select.Value placeholder={String(pagination.perPage)} />
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
		</div>

		<!-- Pagination -->
		{#if pagination.totalPages > 1}
			<div class="col-span-2 flex flex-col items-center md:order-2 md:col-span-3">
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
								<ChevronLeft class="h-4 w-4" />
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
								<ChevronRight class="h-4 w-4" />
							</Pagination.NextButton>
						</Pagination.Item>
					</Pagination.Content>
				</Pagination.Root>
			</div>
		{/if}

		<!-- Count -->
		<div
			class="text-muted-foreground order-3 flex items-center justify-end text-sm {pagination.totalPages ===
			1
				? 'md:col-span-4'
				: undefined}"
		>
			{pagination.totalItems} course{pagination.totalItems > 1 ? 's' : ''}
		</div>
	</div>
{/if}
