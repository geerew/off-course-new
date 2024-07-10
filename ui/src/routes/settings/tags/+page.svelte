<script lang="ts">
	import { AddTagsDialog, DeleteTagsDialog, RenameTagDialog } from '$components/dialogs';
	import { Checkbox, Err, Loading, Pagination, SelectAllCheckbox } from '$components/generic';
	import { Icons } from '$components/icons';
	import { preferences } from '$components/pages/settings_tags/store';
	import { TableSortController } from '$components/table/controllers';
	import { GetTags } from '$lib/api';
	import { TagsRowAction, TagsTableActions } from '$lib/components/pages/settings_tags';
	import * as Table from '$lib/components/ui/table';
	import type { Tag } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { FlattenOrderBy, cn } from '$lib/utils';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { addSortBy } from 'svelte-headless-table/plugins';
	import { toast } from 'svelte-sonner';
	import { get, writable, type Writable } from 'svelte/store';

	// ----------------------
	// Types
	// ----------------------

	type rowProps = {
		title: string;

		// Used to determine if the checkbox is checked or not
		checked: Writable<boolean>;

		// True when an action is selected. Only ONE row can be selected at a time
		rowAction: boolean;
	};

	// ----------------------
	// Variables
	// ----------------------

	// Holds the current page of tags
	const fetchedTags = writable<Tag[]>([]);

	let selectedTag: Tag | undefined = undefined;

	// Populated during getTags(). It holds the state of each row for the current page of
	// the table + any rows that were checked on a previous page
	let workingRows: Record<string, rowProps> = {};

	// The number of rows that are checked
	const checkedRowsCount = writable<number>(0);

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
			initialSortKeys: [$preferences.sortBy],
			toggleOrder: ['desc', 'asc'],
			serverSide: true
		})
	});

	const columns = table.createColumns([
		table.column({
			header: () => {
				return createRender(SelectAllCheckbox, {
					count: checkedRowsCount,
					totalItems: pagination.totalItems
				}).on('click', () => {
					const foundUnchecked = $fetchedTags.find((tag) => !get(workingRows[tag.id].checked));

					if (foundUnchecked) {
						$fetchedTags.forEach((tag) => {
							workingRows[tag.id].checked.set(true);
						});
					} else {
						$fetchedTags.forEach((tag) => {
							workingRows[tag.id].checked.set(false);
						});
					}

					rowsChange();
				});
			},
			accessor: 'id',
			cell: ({ value }) => {
				return createRender(Checkbox, {
					value: value,
					checked: workingRows[value].checked
				}).on('click', (ev) => {
					// Flip the checked state for this row
					workingRows[ev.detail].checked.update((checked) => {
						return !checked;
					});

					rowsChange();
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
				return createRender(TagsRowAction, { tagId: value.id })
					.on('delete', (ev) => {
						// Set to false for all rows. Only 1 rowAction can be active at a time
						Object.keys(workingRows).forEach((value) => {
							workingRows[value].rowAction = false;
						});

						// Set to true for this row and open the delete dialog
						workingRows[ev.detail].rowAction = true;
						openDeleteDialog = true;
					})
					.on('rename', (ev) => {
						// Set to false for all rows. Only 1 rowAction can be active at a time
						Object.keys(workingRows).forEach((value) => {
							workingRows[value].rowAction = false;
						});

						// Set to true for this row and open the rename dialog
						workingRows[ev.detail].rowAction = true;
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
	const ignoredSortIds = ['id', 'actions'];
	const availableSortColumns: Array<{ id: string; label: string }> = flatColumns
		.filter((col) => !ignoredSortIds.includes(col.id.toString()))
		.map((col) => {
			return { id: col.id.toString(), label: col.header.toString() };
		});

	// Start loading page 1 of the tags
	let load = getTags();

	// ----------------------
	// Functions
	// ----------------------

	// GET a paginated list of tags
	async function getTags(): Promise<boolean> {
		try {
			// sleep for 2 seconds
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

			const tags = response.items as Tag[];

			// Find the rows that were checked and keep them
			const keptRows = Object.keys(workingRows)
				.filter((id) => get(workingRows[id].checked))
				.reduce((acc, id) => {
					return { ...acc, [id]: workingRows[id] };
				}, {});

			// Create a new working row for each row for the current page + any rows that were
			// checked on previous pages
			workingRows = {
				...tags.reduce(
					(acc, tag) => ({
						...acc,
						[tag.id]: {
							title: tag.tag,
							checked: writable(false),
							rowAction: false
						}
					}),
					{}
				),
				...keptRows
			};

			rowsChange(false);

			fetchedTags.set(response.items as Tag[]);

			pagination = {
				...pagination,
				totalItems: response.totalItems,
				totalPages: response.totalPages
			};

			return true;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Returns the row action item or all rows that are checked
	function getRowActionOrChecked(): Record<string, rowProps> {
		const rowAction = Object.keys(workingRows).find((id) => workingRows[id].rowAction);

		if (rowAction) {
			return { [rowAction]: workingRows[rowAction] };
		} else {
			return Object.keys(workingRows).reduce((acc, id) => {
				if (get(workingRows[id].checked)) {
					return { ...acc, [id]: workingRows[id] };
				}
				return acc;
			}, {});
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update `checkedRowsCount` and optionally display a toast as rows are check/
	// unchecked
	function rowsChange(showToast = true) {
		const count = Object.values(workingRows).filter((row) => get(row.checked)).length;

		checkedRowsCount.set(count);

		if (showToast) {
			let message = 'Selected ' + count + ' tag' + (count > 1 ? 's' : '');
			if (count === 0) message = 'Deselected all tags';

			toast.success(message, {
				duration: 2000
			});
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Build a map of tag IDs to their titles
	function buildIdTitleMap(): Record<string, string> {
		const rows = getRowActionOrChecked();

		return Object.keys(rows).reduce<Record<string, string>>((acc, id) => {
			acc[id] = rows[id].title;
			return acc;
		}, {});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	function getRowActionAsTag() {
		const rowAction = Object.keys(workingRows).find((id) => workingRows[id].rowAction);

		if (rowAction) return $fetchedTags.find((tag) => tag.id === rowAction);
	}
</script>

<div class="flex w-full flex-col gap-4 bg-background pb-10 pt-6">
	<div class="container flex flex-col gap-5 md:gap-10">
		{#await load}
			<Loading class="max-h-96" />
		{:then _}
			<div class="flex w-full flex-row">
				<div class="flex w-full flex-col gap-y-5 md:flex-row md:justify-between">
					<div class="flex">
						<AddTagsDialog
							on:added={() => {
								load = getTags();
							}}
						/>
					</div>

					<div class="flex w-full justify-between gap-2.5 md:justify-end">
						<TagsTableActions
							count={checkedRowsCount}
							on:deselect={() => {
								// Set all rows to unchecked
								Object.keys(workingRows).forEach((id) => {
									workingRows[id].checked.set(false);
								});

								rowsChange();
							}}
							on:delete={() => {
								openDeleteDialog = true;
							}}
						/>

						<TableSortController
							columns={availableSortColumns}
							sortedColumn={sortKeys}
							on:changed={(ev) => {
								preferences.set({ ...$preferences, sortBy: ev.detail });
								getTags();
							}}
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
														<Icons.CaretUp
															class="absolute right-0 top-1/2 size-4 -translate-y-1/2 stroke-[2] text-secondary/80"
														/>
													{:else if descSort}
														<Icons.CaretDown
															class="absolute right-0 top-1/2 size-4 -translate-y-1/2 stroke-[2] text-secondary/80"
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
										<p class="text-center text-sm text-muted-foreground">No tags found.</p>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							{#each $pageRows as row (row.id)}
								<Subscribe rowAttrs={row.attrs()} let:rowAttrs>
									<Table.Row
										{...rowAttrs}
										data-row={row.id}
										data-state={workingRows[row.id] && 'selected'}
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
{#if openDeleteDialog}
	<DeleteTagsDialog
		tags={buildIdTitleMap()}
		bind:open={openDeleteDialog}
		on:cancelled={() => {
			// Set to false for all rows
			Object.keys(workingRows).forEach((id) => {
				workingRows[id].rowAction = false;
			});
		}}
		on:deleted={() => {
			// It is possible we need to go back a page
			const count = Object.keys(getRowActionOrChecked()).length;
			if (
				pagination.page > 1 &&
				Math.ceil((pagination.totalItems - count) / pagination.perPage) !== pagination.page
			) {
				pagination.page = pagination.page - 1;
			}

			// Clear all row state
			const rowAction = Object.keys(workingRows).find((id) => workingRows[id].rowAction);

			if (rowAction) {
				workingRows[rowAction].checked.set(false);
				workingRows[rowAction].rowAction = false;
			} else {
				Object.keys(workingRows).forEach((id) => {
					workingRows[id].checked.set(false);
				});
			}

			load = getTags();
		}}
	/>
{/if}

{#if openRenameDialog}
	{#if (selectedTag = getRowActionAsTag())}
		<RenameTagDialog
			tag={selectedTag}
			bind:open={openRenameDialog}
			on:cancelled={() => {
				Object.keys(workingRows).forEach((id) => {
					workingRows[id].rowAction = false;
				});
			}}
			on:renamed={() => {
				Object.keys(workingRows).forEach((id) => {
					workingRows[id].rowAction = false;
				});

				load = getTags();
			}}
		/>
	{/if}
{/if}
