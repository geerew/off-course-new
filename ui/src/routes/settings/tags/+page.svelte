<script lang="ts">
	import { AddTagsDialog, DeleteTagsDialog, RenameTagDialog } from '$components/dialogs';
	import { Checkbox, Err, Loading, Pagination, SelectAllCheckbox } from '$components/generic';
	import { TableSortController } from '$components/table/controllers';
	import { GetTags } from '$lib/api';
	import { TagsRowAction, TagsTableActions } from '$lib/components/pages/settings_tags';
	import * as Table from '$lib/components/ui/table';
	import type { Tag } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { FlattenOrderBy, cn } from '$lib/utils';
	import { ChevronDown, ChevronUp } from 'lucide-svelte';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { addSortBy } from 'svelte-headless-table/plugins';
	import { toast } from 'svelte-sonner';
	import { writable } from 'svelte/store';

	// ----------------------
	// Variables
	// ----------------------
	const fetchedTags = writable<Tag[]>([]);

	// Set when a single tag is selected via the row action
	let selectedTag: Tag | undefined = undefined;

	// Set when tags are selected via the checkbox
	const selectedTags = writable<Record<string, string>>({});
	const selectedTagsCount = writable<number>(0);

	// Dialogs
	let openDeleteDialog = false;
	let openRenameDialog = false;

	let pagination: PaginationParams = {
		page: 1,
		perPage: 10,
		perPages: [10, 25, 100, 200],
		totalItems: -1,
		totalPages: -1
	};

	const table = createTable(fetchedTags, {
		sort: addSortBy({
			initialSortKeys: [{ id: 'tag', order: 'asc' }],
			toggleOrder: ['desc', 'asc'],
			serverSide: true
		})
	});

	const columns = table.createColumns([
		table.column({
			header: () => {
				return createRender(SelectAllCheckbox, {
					selectedCount: selectedTagsCount,
					totalItems: pagination.totalItems
				}).on('click', () => {
					if (Object.keys($selectedTags).length === 0) {
						// Add all current fetched tags to selected tags
						selectedTags.set(
							$fetchedTags.reduce((acc, tag) => ({ ...acc, [tag.id]: tag.tag }), {})
						);
					} else {
						// Search for an unchecked tag on this page
						const foundUncheckedTag = $fetchedTags.find((tag) => !$selectedTags[tag.id]);

						if (foundUncheckedTag) {
							// Add all fetched tags on this page to selected tags
							selectedTags.update((tags) => {
								const newTags = $fetchedTags.reduce((acc, tag) => {
									if (!tags[tag.id]) {
										acc[tag.id] = tag.tag;
									}
									return acc;
								}, tags);

								return newTags;
							});
						} else {
							// All tags on this page are checked. Remove them from selected tags but keep the rest
							selectedTags.update((tags) => {
								const newTags = { ...tags };
								$fetchedTags.forEach((tag) => {
									delete newTags[tag.id];
								});
								return newTags;
							});
						}
					}

					selectedTagsToast();
				});
			},
			accessor: 'id',
			cell: ({ value, row }) => {
				return createRender(Checkbox, {
					selected: selectedTags,
					id: value
				}).on('click', () => {
					selectedTags.update((tags) => {
						if (tags[value]) {
							delete tags[value];
						} else {
							if (row.isData()) {
								tags[value] = row.original.tag;
							}
						}
						return { ...tags };
					});

					selectedTagsToast();
				});
			}
		}),
		table.column({
			header: 'Tag',
			accessor: 'tag'
		}),
		table.column({
			header: 'Course Count',
			accessor: 'courseCount'
		}),
		table.column({
			accessor: (item) => item,
			header: '',
			id: 'actions',
			plugins: {
				sort: {
					disable: true
				}
			},
			cell: ({ value }) => {
				return createRender(TagsRowAction, { tag: value })
					.on('delete', () => {
						selectedTag = value;
						openDeleteDialog = true;
					})
					.on('rename', () => {
						selectedTag = value;
						openRenameDialog = true;
					});
			}
		})
	]);

	const { headerRows, pageRows, tableAttrs, tableBodyAttrs, flatColumns, pluginStates, rows } =
		table.createViewModel(columns, { rowDataId: (row) => row.id });

	// Writable plugin stores
	const { sortKeys } = pluginStates.sort;

	// The columns that can be sorted
	const ignoredSortIds = ['id'];
	const availableSortColumns: Array<{ id: string; label: string }> = flatColumns
		.filter((col) => !ignoredSortIds.includes(col.id.toString()))
		.map((col) => {
			return { id: col.id.toString(), label: col.header.toString() };
		});

	// ----------------------
	// Functions
	// ----------------------

	// GET a paginated list of tags
	async function getTags() {
		try {
			const response = await GetTags({
				orderBy: FlattenOrderBy($sortKeys),
				page: pagination.page,
				perPage: pagination.perPage,
				expand: true
			});

			if (!response) {
				fetchedTags.set([]);
				pagination = { ...pagination, totalItems: 0, totalPages: 0 };
				return true;
			}

			fetchedTags.set(response.items as Tag[]);

			pagination = {
				...pagination,
				totalItems: response.totalItems,
				totalPages: response.totalPages
			};

			return true;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Display a toast when a tag is selected/deselected
	function selectedTagsToast() {
		const count = Object.keys($selectedTags).length;
		let message = 'Selected ' + count + ' tag' + (count > 1 ? 's' : '');

		if (count === 0) {
			message = 'Deselected all tags';
		}

		toast.success(message, {
			duration: 2000
		});
	}

	// ----------------------
	// Reactive
	// ----------------------

	$: selectedTagsCount.set(Object.keys($selectedTags).length);

	// ----------------------
	// Variables
	// ----------------------

	let load = getTags();
</script>

<div class="bg-background flex w-full flex-col gap-4 pb-10 pt-6">
	<div class="container flex flex-col gap-7">
		{#await load}
			<Loading />
		{:then _}
			<div class="flex w-full flex-row">
				<div class="flex w-full justify-between">
					<div class="flex">
						<AddTagsDialog
							on:added={() => {
								// It is possible that the user deleted the last course on this page,
								// therefore we need to set the page to the previous one
								if (pagination.page > 1 && (pagination.totalItems - 1) % pagination.perPage === 0)
									pagination.page = pagination.page - 1;

								selectedTags.set({});

								load = getTags();
							}}
						/>
					</div>

					<div class="flex w-full justify-end gap-2.5">
						<TagsTableActions
							{selectedTagsCount}
							on:deselect={() => {
								selectedTags.set({});
								selectedTagsToast();
							}}
							on:delete={() => {
								openDeleteDialog = true;
							}}
						/>

						<TableSortController
							columns={availableSortColumns}
							sortedColumn={sortKeys}
							on:changed={getTags}
							disabled={$fetchedTags.length === 0}
						/>
					</div>
				</div>
			</div>

			<div class="flex flex-col gap-5">
				<Table.Root {...$tableAttrs} class="min-w-[15rem] border-collapse">
					<Table.Header>
						{#each $headerRows as headerRow}
							<Subscribe rowAttrs={headerRow.attrs()}>
								<Table.Row class="hover:bg-transparent">
									{#each headerRow.cells as cell (cell.id)}
										{@const ascSort =
											$sortKeys.length >= 1 &&
											$sortKeys[0].order === 'asc' &&
											$sortKeys[0].id === cell.id}
										{@const descSort =
											$sortKeys.length >= 1 &&
											$sortKeys[0].order === 'desc' &&
											$sortKeys[0].id === cell.id}

										<Subscribe attrs={cell.attrs()} let:attrs props={cell.props()}>
											<Table.Head
												{...attrs}
												class={cn(
													'relative whitespace-nowrap px-6 tracking-wide [&:has([role=checkbox])]:pl-3',
													cell.id === 'tag' ? 'min-w-96' : 'min-w-[1%]'
												)}
											>
												<div
													class={cn(
														'flex items-center gap-2.5',
														cell.id !== 'tag' && 'justify-center'
													)}
												>
													<Render of={cell.render()} />

													{#if ascSort}
														<ChevronUp
															class="text-secondary/80 absolute right-0 top-1/2 size-4 -translate-y-1/2 stroke-[2]"
														/>
													{:else if descSort}
														<ChevronDown
															class="text-secondary/80 absolute right-0 top-1/2 size-4 -translate-y-1/2 stroke-[2]"
														/>
													{/if}
												</div>
											</Table.Head>
										</Subscribe>
									{/each}
								</Table.Row>
							</Subscribe>
						{/each}
					</Table.Header>

					<Table.Body {...$tableBodyAttrs}>
						{#if $rows.length === 0}
							<Table.Row class="hover:bg-transparent">
								<Table.Cell colspan={flatColumns.length}>
									<div class="flex w-full flex-grow flex-col place-content-center items-center p-5">
										<p class="text-muted-foreground text-center text-sm">No tags found.</p>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							{#each $pageRows as row (row.id)}
								<Subscribe rowAttrs={row.attrs()} let:rowAttrs>
									<Table.Row
										{...rowAttrs}
										data-row={row.id}
										data-state={$selectedTags[row.id] && 'selected'}
									>
										{#each row.cells as cell (cell.id)}
											<Subscribe attrs={cell.attrs()} let:attrs>
												<Table.Cell
													class={cn(
														'whitespace-nowrap px-6 text-sm [&:has([role=checkbox])]:pl-3',
														cell.id === 'tag' ? 'min-w-96' : 'min-w-[1%]'
													)}
													{...attrs}
												>
													<div class={cn(cell.id !== 'tag' && 'text-center')}>
														<Render of={cell.render()} />
													</div>
												</Table.Cell>
											</Subscribe>
										{/each}
									</Table.Row>
								</Subscribe>
							{/each}
						{/if}
					</Table.Body>
				</Table.Root>

				<Pagination
					type="tag"
					{pagination}
					on:pageChange={(ev) => {
						pagination.page = ev.detail;
						load = getTags();
					}}
					on:perPageChange={(ev) => {
						pagination.perPage = ev.detail;
						pagination.page = 1;
						load = getTags();
					}}
				/>
			</div>
		{:catch error}
			<Err errorMessage={error} />
		{/await}
	</div>
</div>

<!-- Delete dialog -->
<DeleteTagsDialog
	tags={selectedTag ? { [selectedTag.id]: selectedTag.tag } : $selectedTags}
	bind:open={openDeleteDialog}
	on:cancelled={() => {
		selectedTag = undefined;
	}}
	on:deleted={() => {
		// It is possible we need to go back a page
		const count = selectedTag ? 1 : Object.keys($selectedTags).length;
		if (
			pagination.page > 1 &&
			Math.ceil((pagination.totalItems - count) / pagination.perPage) !== pagination.page
		) {
			pagination.page = pagination.page - 1;
		}

		if (selectedTag) {
			// If a single tag was deleted, remove it from the selected tags
			selectedTags.update((tags) => {
				if (selectedTag) delete tags[selectedTag.id];
				return { ...tags };
			});

			selectedTag = undefined;
		} else {
			selectedTags.set({});
		}

		load = getTags();
	}}
/>

{#if selectedTag}
	<RenameTagDialog
		tag={selectedTag}
		bind:open={openRenameDialog}
		on:cancelled={() => {
			selectedTag = undefined;
		}}
		on:renamed={() => {
			selectedTag = undefined;
			load = getTags();
		}}
	/>
{/if}
