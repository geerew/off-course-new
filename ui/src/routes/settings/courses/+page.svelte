<script lang="ts">
	import { DeleteCourseDialog } from '$components/dialogs';
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
	import { AddCoursesSheet } from '$components/sheets';
	import { TableColumnsController, TableSortController } from '$components/table/controllers';
	import { Pagination } from '$components/table/pagination';
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
	const { headerRows, rows, tableAttrs, tableBodyAttrs, pluginStates, flatColumns } =
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
					<AddCoursesSheet
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

			<div class="flex w-full overflow-x-auto">
				<table {...$tableAttrs}>
					<thead>
						{#each $headerRows as headerRow (headerRow.id)}
							<Subscribe rowAttrs={headerRow.attrs()} let:rowAttrs>
								<tr {...rowAttrs}>
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
											<th {...attrs}>
												<div
													class={cn(
														'flex select-none items-center gap-2.5',
														cell.id !== 'title' && 'justify-center'
													)}
												>
													<Render of={cell.render()} />

													{#if ascSort}
														<ChevronUp
															class="text-secondary/80 absolute right-0 top-1/2 h-4 w-4 -translate-y-1/2 stroke-[2]"
														/>
													{:else if descSort}
														<ChevronDown
															class="text-secondary/80 absolute right-0 top-1/2 h-4 w-4 -translate-y-1/2 stroke-[2]"
														/>
													{/if}
												</div>
											</th>
										</Subscribe>
									{/each}
								</tr>
							</Subscribe>
						{/each}
					</thead>
					<tbody {...$tableBodyAttrs}>
						{#if $rows.length === 0}
							<tr>
								<td colspan={flatColumns.length}>
									<div class="flex w-full flex-grow flex-col place-content-center items-center p-5">
										<p class="text-muted-foreground text-center text-sm">No courses found.</p>
									</div>
								</td>
							</tr>
						{:else}
							{#each $rows as row (row.id)}
								<Subscribe rowAttrs={row.attrs()} let:rowAttrs>
									<tr {...rowAttrs} class="hover:bg-muted">
										{#each row.cells as cell (cell.id)}
											<Subscribe attrs={cell.attrs()} let:attrs>
												<td
													{...attrs}
													class={cell.id === 'title' ? 'min-w-96' : 'min-w-[1%] whitespace-nowrap'}
												>
													<Render of={cell.render()} />
												</td>
											</Subscribe>
										{/each}
									</tr>
								</Subscribe>
							{/each}
						{/if}
					</tbody>
				</table>
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
	on:deleted={() => {
		selectedCourse = {};
	}}
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
