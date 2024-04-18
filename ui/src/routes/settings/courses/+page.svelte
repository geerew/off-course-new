<script lang="ts">
	import { Err, Loading } from '$components';
	import { DeleteCourse } from '$components/dialogs';
	import { AddCourses } from '$components/sheets';
	import { Columns, Sort } from '$components/table/controllers';
	import { Pagination } from '$components/table/pagination';
	import {
		CourseAction,
		CourseAvailability,
		CourseProgress,
		NiceDate,
		ScanStatus
	} from '$components/table/renderers';
	import { AddScan, GetCourses } from '$lib/api';
	import type { Course } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { cn, flattenOrderBy } from '$lib/utils';
	import { ChevronDown, ChevronUp } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { addHiddenColumns, addSortBy } from 'svelte-headless-table/plugins';
	import { toast } from 'svelte-sonner';
	import { writable } from 'svelte/store';

	// ----------------------
	// Variables
	// ----------------------

	const courses = writable<Course[]>([]);

	// True when loading the courses
	let loadingCourses = true;

	// True when an error occurs
	let gotError = false;

	// Set when a course is selected for delete
	let deleteCourseId = '';

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
	const table = createTable(courses, {
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
			header: 'Course',
			accessor: 'title'
		}),
		table.column({
			header: 'Availability',
			accessor: 'available',
			cell: ({ value }) => {
				return createRender(CourseAvailability, { available: value });
			}
		}),
		table.column({
			header: 'Progress',
			accessor: 'percent',
			cell: ({ value }) => {
				return createRender(CourseProgress, { percent: value });
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
			cell: ({ value }) => {
				return createRender(CourseAction, { course: value })
					.on('delete', () => {
						deleteCourseId = value.id;
						openDeleteDialog = true;
					})
					.on('scan', () => startScan(value));
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
	const availableSortIds = ['actions'];
	const availableSortColumns: Array<{ id: string; label: string }> = flatColumns
		.filter((col) => !availableSortIds.includes(col.id.toString()))
		.map((col) => {
			return { id: col.id.toString(), label: col.header.toString() };
		});

	// The columns that can be hidden
	const availableExcludeIds = ['title', 'actions'];
	const availableHiddenColumns: Array<{ id: string; label: string }> = flatColumns
		.filter((col) => !availableExcludeIds.includes(col.id.toString()))
		.map((col) => {
			return { id: col.id.toString(), label: col.header.toString() };
		});

	// ----------------------
	// Functions
	// ----------------------

	// GET all courses from the backend. The response is paginated
	const getCourses = async () => {
		loadingCourses = true;

		const orderBy = flattenOrderBy($sortKeys);

		try {
			const response = await GetCourses({
				orderBy: orderBy,
				page: pagination.page,
				perPage: pagination.perPage
			});

			if (!response) {
				courses.set([]);
				pagination = { ...pagination, totalItems: 0, totalPages: 0 };
				return;
			}

			courses.set(response.items as Course[]);

			pagination = {
				...pagination,
				totalItems: response.totalItems,
				totalPages: response.totalPages
			};

			loadingCourses = false;
			return true;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));

			loadingCourses = false;
			gotError = true;

			return false;
		}
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update a course in the courses array
	const updateCourseInCourses = (updatedCourse: Course) => {
		courses.update((currentCourses) => {
			const index = currentCourses.findIndex((course) => course.id === updatedCourse.id);
			if (index !== -1) {
				currentCourses[index] = updatedCourse;
			}
			return [...currentCourses]; // Return a new array to ensure reactivity
		});
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Start a scan for a course
	const startScan = async (course: Course) => {
		try {
			const response = await AddScan(course.id);
			if (!response) throw new Error('Failed to start scan');

			course.scanStatus = 'waiting';
			updateCourseInCourses(course);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	};

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		if (!(await getCourses())) return;
	});
</script>

<div class="bg-background flex w-full flex-col gap-4 pb-10">
	<!-- Heading -->
	<div class="w-full border-b">
		<div class="container flex items-center gap-2.5 py-2 md:py-4">
			<span class="grow text-lg font-semibold md:text-2xl">Courses</span>
			<AddCourses
				on:added={() => {
					pagination.page = 1;
					getCourses();
				}}
			/>
		</div>
	</div>

	<div class="container flex flex-col gap-4">
		<div class="flex w-full flex-row">
			<div class="flex w-full justify-end gap-2.5">
				<Sort
					columns={availableSortColumns}
					sortedColumn={sortKeys}
					on:changed={getCourses}
					disabled={gotError || $courses.length === 0}
				/>
				<Columns
					columns={availableHiddenColumns}
					columnStore={hiddenColumnIds}
					disabled={gotError || $courses.length === 0}
				/>
			</div>
		</div>

		{#if loadingCourses}
			<div
				class="flex min-h-[10rem] w-full flex-grow flex-col place-content-center items-center p-10"
			>
				<Loading />
			</div>
		{:else if gotError}
			<Err />
		{:else}
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
										<Subscribe attrs={cell.attrs()} let:attrs props={cell.props()} let:props>
											<th {...attrs}>
												<div
													class={cn(
														'flex items-center gap-2.5',
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
									<tr {...rowAttrs}>
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
				{pagination}
				on:pageChange={(ev) => {
					pagination.page = ev.detail;
					getCourses();
				}}
				on:perPageChange={(ev) => {
					pagination.perPage = ev.detail;
					pagination.page = 1;
					getCourses();
				}}
			/>
		{/if}
	</div>
</div>

<!-- Delete dialog -->
<DeleteCourse
	courseId={deleteCourseId}
	bind:open={openDeleteDialog}
	on:courseDeleted={() => {
		// It is possible that the user deleted the last course on this page,
		// therefore we need to set the page to the previous one
		if (pagination.page > 1 && (pagination.totalItems - 1) % pagination.perPage === 0)
			pagination.page = pagination.page - 1;

		getCourses();
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
			@apply border-y px-4 py-2.5 text-left text-sm;
		}
	}
</style>
