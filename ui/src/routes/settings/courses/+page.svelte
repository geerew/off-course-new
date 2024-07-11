<script lang="ts">
	import { AddCourseTagsDialog, AddCoursesDialog, DeleteCourseDialog } from '$components/dialogs';
	import {
		Checkbox,
		Err,
		Loading,
		NiceDate,
		Pagination,
		ScanStatus,
		SelectAllCheckbox
	} from '$components/generic';
	import { Icons } from '$components/icons';
	import {
		CoursesRowAction,
		CoursesRowAvailability,
		CoursesRowProgress,
		CoursesTableActions
	} from '$components/pages/settings_courses';
	import { preferences } from '$components/pages/settings_courses/store';
	import { TableColumnsController, TableSortController } from '$components/table/controllers';
	import * as Table from '$components/ui/table';
	import { AddScan, GetCourses } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { FlattenOrderBy, cn } from '$lib/utils';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { addHiddenColumns, addSortBy } from 'svelte-headless-table/plugins';
	import { toast } from 'svelte-sonner';
	import { get, writable, type Writable } from 'svelte/store';

	// ----------------------
	// Types
	// ----------------------

	type rowProps = {
		title: string;

		// Used to determine if the checkbox is checked or not
		checked: Writable<boolean>;

		// Used to determine if a scan is in progress for this row
		scanPoll: Writable<boolean>;

		// True when an action is selected. Only ONE row can be selected at a time
		rowAction: boolean;
	};

	// ----------------------
	// Variables
	// ----------------------

	// Holds the current page of courses
	const fetchedCourses = writable<Course[]>([]);

	// Populated during getCourses(). It holds the state of each row for the current page of
	// the table + any rows that were checked on a previous page
	let workingRows: Record<string, rowProps> = {};

	// The number of rows that are checked
	const checkedRowsCount = writable<number>(0);

	let openDeleteDialog = false;
	let openAddTagsDialog = false;

	// Pagination
	let pagination: PaginationParams = {
		page: 1,
		perPage: 10,
		perPages: [10, 25, 100, 200],
		totalItems: -1,
		totalPages: -1
	};

	// Create the table
	const table = createTable(fetchedCourses, {
		sort: addSortBy({
			initialSortKeys: [$preferences.sortBy],
			toggleOrder: ['desc', 'asc'],
			serverSide: true
		}),
		hide: addHiddenColumns({
			initialHiddenColumnIds: $preferences.hiddenColumns
		})
	});

	// Define the table columns
	const columns = table.createColumns([
		table.column({
			header: () => {
				return createRender(SelectAllCheckbox, {
					count: checkedRowsCount,
					totalItems: pagination.totalItems
				}).on('click', () => {
					const foundUnchecked = $fetchedCourses.find(
						(course) => !get(workingRows[course.id].checked)
					);

					if (foundUnchecked) {
						$fetchedCourses.forEach((course) => {
							workingRows[course.id].checked.set(true);
						});
					} else {
						$fetchedCourses.forEach((course) => {
							workingRows[course.id].checked.set(false);
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
					workingRows[ev.detail].checked.update((checked) => {
						return !checked;
					});

					rowsChange();
				});
			}
		}),
		table.column({
			header: 'Course',
			accessor: 'title'
		}),
		table.column({
			header: 'Availability',
			accessor: 'available',
			cell: ({ value }) => {
				return createRender(CoursesRowAvailability, { available: value });
			}
		}),
		table.column({
			header: 'Progress',
			accessor: 'percent',
			cell: ({ value }) => {
				return createRender(CoursesRowProgress, { percent: value });
			}
		}),
		table.column({
			header: 'Added',
			accessor: 'createdAt',
			cell: ({ value }) => {
				return createRender(NiceDate, { date: value });
			}
		}),
		table.column({
			header: 'Updated',
			accessor: 'updatedAt',
			cell: ({ value }) => {
				return createRender(NiceDate, { date: value });
			}
		}),
		table.column({
			header: 'Scan Status',
			accessor: 'scanStatus',
			cell: ({ row, value }) => {
				if (!row.isData()) return value;

				return createRender(ScanStatus, {
					courseId: row.original.id,
					initialStatus: row.original.scanStatus,
					poll: workingRows[row.original.id].scanPoll
				}).on('empty', (ev) => {
					updateCourseInCourses(ev.detail);
				});
			}
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
				return createRender(CoursesRowAction, {
					courseId: value.id,
					scanning: workingRows[value.id].scanPoll
				})
					.on('delete', (ev) => {
						// Set to false for all rows. Only 1 rowAction can be active at a time
						Object.keys(workingRows).forEach((value) => {
							workingRows[value].rowAction = false;
						});

						// Set to true for this row and open the delete dialog
						workingRows[ev.detail].rowAction = true;
						openDeleteDialog = true;
					})
					.on('scan', (ev) => {
						// Set to false for all rows. Only 1 rowAction can be active at a time
						Object.keys(workingRows).forEach((value) => {
							workingRows[value].rowAction = false;
						});

						// Set to true for this row and start the scan
						workingRows[ev.detail].rowAction = true;
						startScans();
					});
			}
		})
	]);

	// Create the view, which is used when building the table
	const { headerRows, pageRows, tableAttrs, tableBodyAttrs, pluginStates, flatColumns } =
		table.createViewModel(columns);

	// Writable plugin stores
	const { sortKeys } = pluginStates.sort;
	const { hiddenColumnIds } = pluginStates.hide;

	// The columns that can be sorted
	const ignoredSortIds = ['id', 'actions'];
	const availableSortColumns: Array<{ id: string; label: string }> = flatColumns
		.filter((col) => !ignoredSortIds.includes(col.id.toString()))
		.map((col) => {
			return { id: col.id.toString(), label: col.header.toString() };
		});

	// The columns that can be hidden
	const ignoredExcludeIds = ['id', 'title', 'actions'];
	const availableHiddenColumns: Array<{ id: string; label: string }> = flatColumns
		.filter((col) => !ignoredExcludeIds.includes(col.id.toString()))
		.map((col) => {
			return { id: col.id.toString(), label: col.header.toString() };
		});

	// Start loading page 1 of the courses
	let load = getCourses();

	// ----------------------
	// Functions
	// ----------------------

	// GET all courses from the backend. The response is paginated
	async function getCourses(): Promise<boolean> {
		const orderBy = FlattenOrderBy($sortKeys);

		const response = await GetCourses({
			orderBy: orderBy,
			page: pagination.page,
			perPage: pagination.perPage
		});

		if (!response) {
			fetchedCourses.set([]);
			pagination = { ...pagination, totalItems: 0, totalPages: 0 };
			return true;
		}

		const courses = response.items as Course[];

		// Find the rows that were checked and keep them
		const keptRows = Object.keys(workingRows)
			.filter((id) => get(workingRows[id].checked))
			.reduce((acc, id) => {
				return { ...acc, [id]: workingRows[id] };
			}, {});

		// Create a new working row for each row for the current page + any rows that were
		// checked on previous pages
		workingRows = {
			...courses.reduce(
				(acc, course) => ({
					...acc,
					[course.id]: {
						title: course.title,
						checked: writable(false),
						scanPoll: course.scanStatus ? writable(true) : writable(false),
						rowAction: false
					}
				}),
				{}
			),
			...keptRows
		};

		rowsChange(false);

		fetchedCourses.set(courses);

		pagination = {
			...pagination,
			totalItems: response.totalItems,
			totalPages: response.totalPages
		};

		return true;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update a course in the courses array
	function updateCourseInCourses(updatedCourse: Course) {
		fetchedCourses.update((currentCourses) => {
			const index = currentCourses.findIndex((course) => course.id === updatedCourse.id);
			if (index !== -1) {
				currentCourses[index] = updatedCourse;
			}
			return [...currentCourses]; // Return a new array to ensure reactivity
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Start scanning courses
	async function startScans() {
		const rows = getRowActionOrChecked();

		try {
			await Promise.all(
				Object.keys(rows).map(async (id) => {
					try {
						await AddScan(id);
						workingRows[id].scanPoll.set(true);
					} catch (error) {
						toast.error('Failed to start a scan for: ' + workingRows[id]);
					}
				})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		} finally {
			const rowAction = Object.keys(workingRows).find((id) => workingRows[id].rowAction);

			if (rowAction) {
				workingRows[rowAction].rowAction = false;
			} else {
				Object.keys(workingRows).forEach((id) => {
					workingRows[id].checked.set(false);
				});
			}

			rowsChange(false);
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
			let message = 'Selected ' + count + ' course' + (count > 1 ? 's' : '');
			if (count === 0) message = 'Deselected all courses';

			toast.success(message, {
				duration: 2000
			});
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Build a map of course IDs to their titles
	function buildIdTitleMap(): Record<string, string> {
		const rows = getRowActionOrChecked();

		return Object.keys(rows).reduce<Record<string, string>>((acc, id) => {
			acc[id] = rows[id].title;
			return acc;
		}, {});
	}
</script>

<div class="flex w-full flex-col gap-4 bg-background pb-10 pt-6">
	<div class="container flex flex-col gap-5 md:gap-10">
		{#await load}
			<Loading class="max-h-96" />
		{:then _}
			<div class="flex w-full flex-row">
				<div class="flex w-full flex-col gap-y-5 md:flex-row md:justify-between">
					<AddCoursesDialog
						on:added={() => {
							pagination.page = 1;
							load = getCourses();
						}}
					/>

					<div class="flex w-full justify-between gap-2.5 md:justify-end">
						<CoursesTableActions
							count={checkedRowsCount}
							on:deselect={() => {
								// Set all rows to unchecked
								Object.keys(workingRows).forEach((id) => {
									workingRows[id].checked.set(false);
								});

								rowsChange();
							}}
							on:scan={() => {
								startScans();
							}}
							on:tags={() => {
								openAddTagsDialog = true;
							}}
							on:delete={() => {
								openDeleteDialog = true;
							}}
						/>

						<div class="flex gap-2.5">
							<TableSortController
								columns={availableSortColumns}
								sortedColumn={sortKeys}
								disabled={$fetchedCourses.length === 0}
								on:changed={(ev) => {
									preferences.set({ ...$preferences, sortBy: ev.detail });
									getCourses();
								}}
							/>

							<TableColumnsController
								columns={availableHiddenColumns}
								columnStore={hiddenColumnIds}
								disabled={$fetchedCourses.length === 0}
								on:changed={(ev) => {
									preferences.set({ ...$preferences, hiddenColumns: ev.detail });
								}}
							/>
						</div>
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
													cell.id === 'title' ? 'min-w-96' : 'min-w-[1%]'
												)}
											>
												<div
													class={cn(
														'flex select-none items-center gap-2.5',
														cell.id !== 'title' && 'justify-center'
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
						{#if $pageRows.length === 0}
							<Table.Row class="hover:bg-transparent">
								<Table.Cell colspan={flatColumns.length}>
									<div class="flex w-full flex-grow flex-col place-content-center items-center p-5">
										<p class="text-center text-sm text-muted-foreground">No courses found.</p>
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
														cell.id === 'title' ? 'min-w-96' : 'min-w-[1%]'
													)}
													{...attrs}
												>
													<div class={cn(cell.id !== 'title' && 'text-center')}>
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
			</div>

			<Pagination
				type="course"
				{pagination}
				on:pageChange={(ev) => {
					pagination.page = ev.detail;
					load = getCourses();
				}}
				on:perPageChange={(ev) => {
					pagination.perPage = ev.detail;
					pagination.page = 1;
					load = getCourses();
				}}
			/>
		{:catch error}
			<Err errorMessage={error} />
		{/await}
	</div>
</div>

<!-- Delete dialog -->
{#if openDeleteDialog}
	<DeleteCourseDialog
		courses={buildIdTitleMap()}
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

			load = getCourses();
		}}
	/>
{/if}

<!-- Add tags dialog -->
{#if openAddTagsDialog}
	<AddCourseTagsDialog
		courseIds={Object.keys(getRowActionOrChecked())}
		bind:open={openAddTagsDialog}
		on:updated={() => {
			// Set checked to false for all rows
			Object.keys(workingRows).forEach((id) => {
				workingRows[id].checked.set(false);
			});

			rowsChange(false);
		}}
	/>
{/if}
