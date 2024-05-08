<script lang="ts">
	import { AddCourseTagsDialog, AddCoursesDialog, DeleteCourseDialog } from '$components/dialogs';
	import {
		Checkbox,
		Err,
		Loading,
		NiceDate,
		ScanStatus,
		SelectAllCheckbox
	} from '$components/generic';
	import {
		CoursesRowAction,
		CoursesRowAvailability,
		CoursesRowProgress,
		CoursesTableActions
	} from '$components/pages/settings_courses';
	import { TableColumnsController, TableSortController } from '$components/table/controllers';
	import { Pagination } from '$components/table/pagination';
	import * as Table from '$components/ui/table';
	import { AddScan, GetCourses } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { cn, flattenOrderBy } from '$lib/utils';
	import { ChevronDown, ChevronUp } from 'lucide-svelte';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { addHiddenColumns, addSortBy } from 'svelte-headless-table/plugins';
	import { toast } from 'svelte-sonner';
	import { writable } from 'svelte/store';

	// ----------------------
	// Variables
	// ----------------------
	const fetchedCourses = writable<Course[]>([]);

	// Set when a single course is selected via the row action
	let selectedCourse = <Record<string, string>>{};

	// Set when courses are selected via the checkbox
	const selectedCourses = writable<Record<string, string>>({});
	const selectedCoursesCount = writable<number>(0);

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
			initialSortKeys: [{ id: 'createdAt', order: 'desc' }],
			toggleOrder: ['desc', 'asc'],
			serverSide: true
		}),
		hide: addHiddenColumns({
			initialHiddenColumnIds: ['updatedAt']
		})
	});

	// Define the table columns
	const columns = table.createColumns([
		table.column({
			header: () => {
				return createRender(SelectAllCheckbox, {
					selectedCount: selectedCoursesCount,
					totalItems: pagination.totalItems
				}).on('click', () => {
					if (Object.keys($selectedCourses).length === 0) {
						// Add all current fetched courses to selected courses
						selectedCourses.set(
							$fetchedCourses.reduce((acc, course) => ({ ...acc, [course.id]: course.title }), {})
						);
					} else {
						// Search for an unchecked course on this page
						const foundUncheckedTag = $fetchedCourses.find(
							(course) => !$selectedCourses[course.id]
						);

						if (foundUncheckedTag) {
							// Add all fetched courses on this page to selected courses
							selectedCourses.update((courses) => {
								const newTags = $fetchedCourses.reduce((acc, course) => {
									if (!courses[course.id]) {
										acc[course.id] = course.title;
									}
									return acc;
								}, courses);

								return newTags;
							});
						} else {
							// All curses on this page are checked. Remove them from selected courses but keep the rest
							selectedCourses.update((courses) => {
								const newTags = { ...courses };
								$fetchedCourses.forEach((course) => {
									delete newTags[course.id];
								});
								return newTags;
							});
						}
					}

					selectedCoursesToast();
				});
			},
			accessor: 'id',
			cell: ({ value, row }) => {
				return createRender(Checkbox, {
					selected: selectedCourses,
					id: value
				}).on('click', () => {
					selectedCourses.update((courses) => {
						if (courses[value]) {
							delete courses[value];
						} else {
							if (row.isData()) {
								courses[value] = row.original.title;
							}
						}
						return { ...courses };
					});

					selectedCoursesToast();
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
				return createRender(ScanStatus, { courseId: row.original.id, scanStatus: value }).on(
					'change',
					(ev) => {
						// row.original.scanStatus = ev.detail;
						updateCourseInCourses(ev.detail);
					}
				);
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
			cell: ({ value, row }) => {
				return createRender(CoursesRowAction, { course: value })
					.on('delete', () => {
						if (row.isData()) {
							selectedCourse[value.id] = value.title;
							openDeleteDialog = true;
						}
					})
					.on('scan', () => {
						selectedCourse[value.id] = value.title;
						startScans(selectedCourse);
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

	// ----------------------
	// Functions
	// ----------------------

	// GET all courses from the backend. The response is paginated
	const getCourses = async () => {
		const orderBy = flattenOrderBy($sortKeys);

		try {
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

			fetchedCourses.set(response.items as Course[]);

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
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update a course in the courses array
	const updateCourseInCourses = (updatedCourse: Course) => {
		fetchedCourses.update((currentCourses) => {
			const index = currentCourses.findIndex((course) => course.id === updatedCourse.id);
			if (index !== -1) {
				currentCourses[index] = updatedCourse;
			}
			return [...currentCourses]; // Return a new array to ensure reactivity
		});
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Start a scan for a course
	const startScans = async (courses: Record<string, string>) => {
		try {
			const ids = Object.keys(courses);

			await Promise.all(
				ids.map(async (id) => {
					try {
						await AddScan(id);
						const course = $fetchedCourses.find((course) => course.id === id);

						// Update the course so it reflects in the table
						if (course) {
							course.scanStatus = 'waiting';
							updateCourseInCourses(course);
						}

						toast.success('Started scan for ' + courses[id]);
					} catch (error) {
						toast.error('Failed to start a scan for: ' + courses[id]);
					}
				})
			);

			selectedCourses.set({});
			selectedCourse = {};
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Display a toast when a course is selected/deselected
	const selectedCoursesToast = () => {
		const count = Object.keys($selectedCourses).length;
		let message = 'Selected ' + count + ' course' + (count > 1 ? 's' : '');

		if (count === 0) message = 'Deselected all courses';

		toast.success(message, {
			duration: 2000
		});
	};

	// ----------------------
	// Variables
	// ----------------------

	$: selectedCoursesCount.set(Object.keys($selectedCourses).length);

	// ----------------------
	// Variables
	// ----------------------

	let load = getCourses();
</script>

<div class="bg-background flex w-full flex-col gap-4 pb-10 pt-6">
	<div class="container flex flex-col gap-10">
		{#await load}
			<Loading />
		{:then _}
			<div class="flex w-full flex-row">
				<div class="flex w-full justify-between">
					<AddCoursesDialog
						on:added={() => {
							pagination.page = 1;
							load = getCourses();
						}}
					/>

					<div class="flex w-full justify-end gap-2.5">
						<CoursesTableActions
							{selectedCoursesCount}
							on:deselect={() => {
								selectedCourses.set({});
								selectedCoursesToast();
							}}
							on:scan={() => {
								startScans($selectedCourses);
							}}
							on:tags={() => {
								openAddTagsDialog = true;
							}}
							on:delete={() => {
								openDeleteDialog = true;
							}}
						/>

						<TableSortController
							columns={availableSortColumns}
							sortedColumn={sortKeys}
							on:changed={getCourses}
							disabled={$fetchedCourses.length === 0}
						/>
						<TableColumnsController
							columns={availableHiddenColumns}
							columnStore={hiddenColumnIds}
							disabled={$fetchedCourses.length === 0}
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
						{#if $pageRows.length === 0}
							<Table.Row class="hover:bg-transparent">
								<Table.Cell colspan={flatColumns.length}>
									<div class="flex w-full flex-grow flex-col place-content-center items-center p-5">
										<p class="text-muted-foreground text-center text-sm">No courses found.</p>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							{#each $pageRows as row (row.id)}
								<Subscribe rowAttrs={row.attrs()} let:rowAttrs>
									<Table.Row
										{...rowAttrs}
										data-row={row.id}
										data-state={$selectedCourses[row.id] && 'selected'}
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
				type={'course'}
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
<DeleteCourseDialog
	courses={Object.keys(selectedCourse).length > 0 ? selectedCourse : $selectedCourses}
	bind:open={openDeleteDialog}
	on:cancelled={() => {
		selectedCourse = {};
	}}
	on:deleted={() => {
		// It is possible that the user deleted the last course on this page,
		// therefore we need to set the page to the previous one
		if (pagination.page > 1 && (pagination.totalItems - 1) % pagination.perPage === 0)
			pagination.page = pagination.page - 1;

		if (Object.keys(selectedCourse).length > 0) {
			// If a single course was deleted, remove it from the selected courses
			selectedCourses.update((courses) => {
				delete courses[Object.keys(selectedCourse)[0]];
				return { ...courses };
			});

			selectedCourse = {};
		} else {
			selectedCourses.set({});
		}

		load = getCourses();
	}}
/>

<!-- Add tags dialog -->
<AddCourseTagsDialog
	courseIds={Object.keys($selectedCourses)}
	bind:open={openAddTagsDialog}
	on:deleted={() => {
		// It is possible that the user deleted the last course on this page,
		// therefore we need to set the page to the previous one
		if (pagination.page > 1 && (pagination.totalItems - 1) % pagination.perPage === 0)
			pagination.page = pagination.page - 1;

		selectedCourses.set({});
		load = getCourses();
	}}
/>

<style lang="postcss">
	table {
		@apply w-full min-w-[50rem] border-collapse;

		& > thead > tr > th {
			@apply relative whitespace-nowrap border-y px-6 py-4 text-left text-sm font-semibold tracking-wide;
			@apply text-muted-foreground;
		}

		& > tbody > tr > td {
			@apply border-y px-6 py-2.5 text-left text-sm;
		}
	}
</style>
