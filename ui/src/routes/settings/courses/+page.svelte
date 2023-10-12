<script lang="ts">
	import { Error, Loading } from '$components';
	import { Icons } from '$components/icons';
	import { Pagination, PerPage } from '$components/pagination';
	import { AddCoursesDrawer, DeleteCourseDialog } from '$components/settings';
	import { TableCourseAction, TableCourseTitle, TableDate } from '$components/table';
	import type { Course } from '$lib/types/models';
	import type { PaginationData } from '$lib/types/pagination';
	import { AddScan, GetCourses } from '$lib/utils/api';
	import { cn, flattenOrderBy, isBrowser } from '$lib/utils/general';
	import { createDialog } from '@melt-ui/svelte';
	import { onMount } from 'svelte';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import type { SortKey } from 'svelte-headless-table/lib/plugins/addSortBy';
	import { addHiddenColumns, addSortBy } from 'svelte-headless-table/plugins';
	import { writable } from 'svelte/store';
	import theme from 'tailwindcss/defaultTheme';

	// ----------------------
	// Variables
	// ----------------------

	const courses = writable<Course[]>([]);

	// True when loading the courses. It is used to render a loading icon
	let loadingCourses = true;

	// When a course is selected for scan/delete/details, this is set to the course id
	let selectedCourseId = '';

	// True when the API call errors
	let gotError = false;

	// Pagination
	let pagination: PaginationData = {
		page: 1,
		perPage: 25,
		perPages: [10, 25, 100, 200],
		totalItems: -1,
		totalPages: -1
	};

	// Set the current sort column. This is updated
	let currentSortColumn: SortKey = { id: 'createdAt', order: 'desc' };

	// When the screen size moves to small or below, the table columns are hidden and this will be
	// set to true. When the screen size moves to medium or above, the columns are shown again and
	// this will be set to false
	let columnsHidden = false;

	// Create the table
	const table = createTable(courses, {
		sort: addSortBy({
			initialSortKeys: [currentSortColumn],
			toggleOrder: ['asc', 'desc'],
			serverSide: true
		}),
		hideCols: addHiddenColumns()
	});

	// Define the table columns
	const columns = table.createColumns([
		table.column({
			accessor: (item) => item,
			id: 'title',
			header: 'course',
			cell: ({ value }) => {
				return createRender(TableCourseTitle, { course: value })
					.on('delete', () => {
						selectedCourseId = value.id;
						openDeleteDialog.set(true);
					})
					.on('scan', async () => {
						await AddScan({ courseId: value.id })
							.then(() => {
								value.scanStatus = 'waiting';
								updateCourseById(value);
							})
							.catch((err) => {
								console.error(err);
							});
					})
					.on('change', () => {
						updateCourseById(value);
					});
			}
		}),
		table.column({
			accessor: 'createdAt',
			header: 'added',
			cell: ({ value }) => {
				return createRender(TableDate, { date: value });
			}
		}),
		table.column({
			accessor: 'updatedAt',
			header: 'updated',
			cell: ({ value }) => {
				return createRender(TableDate, { date: value });
			}
		}),
		table.column({
			accessor: (item) => item,
			header: '',
			plugins: {
				sort: {
					disable: true
				}
			},
			cell: ({ value }) => {
				return createRender(TableCourseAction, { course: value })
					.on('delete', () => {
						selectedCourseId = value.id;
						openDeleteDialog.set(true);
					})
					.on('scan', async () => {
						await AddScan({ courseId: value.id })
							.then(() => {
								value.scanStatus = 'waiting';
								updateCourseById(value);
							})
							.catch((err) => {
								console.error(err);
							});
					})
					.on('change', () => {
						updateCourseById(value);
					});
			}
		})
	]);

	// Create the table view
	const { flatColumns, headerRows, rows, tableAttrs, tableBodyAttrs, pluginStates } =
		table.createViewModel(columns);

	// Get the sortKeys and hiddenColumnIds
	const { sortKeys } = pluginStates.sort;
	const { hiddenColumnIds } = pluginStates.hideCols;

	// Create a delete dialog
	const deleteDialog = createDialog({
		role: 'alertdialog'
	});

	// Get the open state of the delete dialog
	const openDeleteDialog = deleteDialog.states.open;

	// ----------------------
	// Functions
	// ----------------------

	// GET all courses from the backend. The response is paginated
	const getCourses = async () => {
		loadingCourses = true;

		const orderBy = flattenOrderBy($sortKeys);

		await GetCourses({
			orderBy: orderBy,
			page: pagination.page,
			perPage: pagination.perPage
		})
			.then((resp) => {
				if (!resp) {
					courses.set([]);
					pagination = { ...pagination, totalItems: 0, totalPages: 0 };
				} else {
					courses.set(resp.items as Course[]);

					pagination = {
						...pagination,
						totalItems: resp.totalItems,
						totalPages: resp.totalPages
					};
				}
			})
			.catch((err) => {
				console.error(err);
				loadingCourses = false;
				gotError = true;
				return;
			});

		loadingCourses = false;
	};

	// When the screen is resized to small or below, the table can be difficult to read. This
	// function hides the columns that are not important on small screens and shows them again
	// when the screen is resized to medium or above
	//
	// On small and below the component <TableCourseTitle /> will render this additional information
	const handleResize = () => {
		if (!isBrowser) return;

		// Check if the screen is small or lower, based upon tailwinds breakpoints
		const isSmall = window.innerWidth < +theme.screens.md.replace('px', '');

		if (isSmall && !columnsHidden) {
			$hiddenColumnIds = flatColumns.filter((c) => c.id !== 'title').map((c) => c.id);
			columnsHidden = true;
		} else if (!isSmall && columnsHidden) {
			$hiddenColumnIds = [];
			columnsHidden = false;
		}
	};

	const updateCourseById = (updatedCourse: Course) => {
		courses.update((currentCourses) => {
			const index = currentCourses.findIndex((course) => course.id === updatedCourse.id);
			if (index !== -1) {
				currentCourses[index] = updatedCourse;
			}
			return [...currentCourses]; // Return a new array to ensure reactivity
		});
	};

	// ----------------------
	// Reactive
	// ----------------------

	// When the table sorting is changed update the currentSortColumn
	$: (async () => {
		if ($sortKeys.length >= 1 && $sortKeys[0] !== currentSortColumn) {
			await getCourses();
			currentSortColumn = $sortKeys[0];
		}
	})();

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(async () => {
		handleResize();
		await getCourses();
	});
</script>

<svelte:window on:resize={handleResize} />

<div class="bg-background flex w-full flex-col gap-4 pb-10">
	<!-- Heading -->
	<div class="bg-background/50 w-full border-b">
		<div class="container flex items-center py-4 md:py-6">
			<span class="grow text-lg font-semibold md:text-2xl">Courses</span>
			<AddCoursesDrawer
				on:added={() => {
					// Move back to page 1 and pull the courses
					pagination.page = 1;
					getCourses();
				}}
			/>
		</div>
	</div>

	<!-- Table -->
	<div class="container">
		{#if loadingCourses}
			<div
				class="flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center p-10"
			>
				<Loading />
			</div>
		{:else if gotError}
			<Error />
		{:else if pagination.totalItems === 0}
			<!-- class="text-foreground-muted flex w-full justify-center py-5 font-semibold"> -->
			<div
				class="flex min-h-[20rem] w-full flex-grow flex-col place-content-center items-center gap-5 p-10"
			>
				<span class="text-xl font-bold">No courses</span>
				<span class="text-foreground-muted text-sm">You could add one?</span>
			</div>
		{:else}
			<div class="">
				<div class="w-full overflow-auto">
					<table {...$tableAttrs} class="w-full border-collapse indent-0 text-sm">
						<thead>
							{#each $headerRows as headerRow (headerRow.id)}
								<Subscribe rowAttrs={headerRow.attrs()} let:rowAttrs>
									<tr {...rowAttrs} class="border-b">
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
												<th
													{...attrs}
													data-column={cell.label}
													class="text-foreground-muted h-12 px-4 text-left font-semibold tracking-wide"
												>
													{#if props.sort.disabled}
														<Render of={cell.render()} />
													{:else}
														<button class="flex items-center gap-1.5" on:click={props.sort.toggle}>
															<Render of={cell.render()} />
															<Icons.arrowUpDown
																class={cn(
																	'stroke-foreground-muted h-4 w-4',
																	ascSort ? '[&>:nth-child(n+3)]:stroke-secondary' : undefined,
																	descSort ? '[&>:nth-child(-n+2)]:stroke-secondary' : undefined
																)}
															/>
														</button>
													{/if}
												</th>
											</Subscribe>
										{/each}
									</tr>
								</Subscribe>
							{/each}
						</thead>
						<tbody {...$tableBodyAttrs} class="">
							{#each $rows as row (row.id)}
								<Subscribe rowAttrs={row.attrs()} let:rowAttrs>
									<tr {...rowAttrs} class="hover:bg-accent-1/50 h-16 border-b sm:h-12">
										{#each row.cells as cell (cell.id)}
											<Subscribe attrs={cell.attrs()} let:attrs>
												<td
													{...attrs}
													class={cn(
														'px-4 py-2',
														cell.column.header === 'course'
															? 'min-w-[15rem]'
															: 'min-w-[1%] whitespace-nowrap'
													)}
													data-row={cell.column.header}
												>
													<Render of={cell.render()} />
												</td>
											</Subscribe>
										{/each}
									</tr>
								</Subscribe>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		{#if pagination.totalItems > 0}
			<div class="grid grid-cols-2 gap-4 pt-5 md:grid-cols-5">
				<PerPage
					class="order-2 md:order-1"
					perPages={pagination.perPages}
					perPage={pagination.perPage}
					on:change={(ev) => {
						pagination = { ...pagination, page: 1, perPage: ev.detail };
						getCourses();
					}}
				/>

				<Pagination
					class="col-span-2 flex flex-col items-center md:order-2 md:col-span-3"
					bind:pagination
					on:change={() => {
						getCourses();
					}}
				/>

				<div
					class="text-foreground-muted order-3 flex items-center justify-end text-sm {pagination.totalPages ===
					1
						? 'md:col-span-4'
						: undefined}"
				>
					{pagination.totalItems} course{pagination.totalItems > 1 ? 's' : ''}
				</div>
			</div>
		{/if}
	</div>

	<!-- Delete Dialog -->
	<DeleteCourseDialog
		dialog={deleteDialog}
		bind:id={selectedCourseId}
		on:confirmed={() => {
			// It is possible that the user deleted the last course for this page. When this is the
			// case, set the page to the previous one
			if (pagination.page > 1 && (pagination.totalItems - 1) % pagination.perPage === 0)
				pagination.page = pagination.page - 1;

			getCourses();
		}}
	/>
</div>
