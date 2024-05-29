<script lang="ts">
	import { Err, Loading, LogLevel, LogMessage, Pagination } from '$components/generic';
	import { LogsMinLogLevel } from '$components/pages/settings_logs';
	import * as Table from '$components/ui/table';
	import { GetLogs } from '$lib/api';
	import type { Log } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { cn } from '$lib/utils';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { writable } from 'svelte/store';

	// ----------------------
	// Variables
	// ----------------------

	// Holds the current page of logs
	const fetchedLogs = writable<Log[]>([]);

	// Holds the minimum log level. -4 is DEBUG, 0 is INFO, 4 is WARNING, 8 is ERROR
	let minLogLevel = -4;

	// Pagination
	let pagination: PaginationParams = {
		page: 1,
		perPage: 25,
		perPages: [10, 25, 100, 200],
		totalItems: -1,
		totalPages: -1
	};

	// Create the table
	const table = createTable(fetchedLogs);

	// Define the table columns
	const columns = table.createColumns([
		table.column({
			header: 'Level',
			accessor: 'level',
			cell: ({ value }) => {
				return createRender(LogLevel, { level: value });
			}
		}),
		table.column({
			header: 'Message',
			accessor: 'message',
			cell: ({ value, row }) => {
				if (!row.isData()) return value;

				return createRender(LogMessage, { message: value, data: row.original.data });
			}
		}),
		table.column({
			header: 'Created',
			accessor: 'createdAt'
		})
	]);

	// Create the view, which is used when building the table
	const { headerRows, pageRows, tableAttrs, tableBodyAttrs, flatColumns } =
		table.createViewModel(columns);

	// Start loading page 1 of the logs
	let load = getLogs();

	// ----------------------
	// Functions
	// ----------------------

	// GET a paginated set of logs from the backend
	async function getLogs(): Promise<boolean> {
		try {
			const response = await GetLogs({
				level: minLogLevel,
				page: pagination.page,
				perPage: pagination.perPage
			});

			if (!response) {
				fetchedLogs.set([]);
				pagination = { ...pagination, totalItems: 0, totalPages: 0 };
				return true;
			}

			fetchedLogs.set(response.items as Log[]);

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
</script>

<div class="bg-background flex w-full flex-col gap-4 pb-10 pt-6">
	<div class="container flex flex-col gap-10">
		{#await load}
			<Loading />
		{:then _}
			<div class="flex w-full flex-row">
				<div class="flex w-full justify-between">
					<LogsMinLogLevel
						level={minLogLevel}
						on:change={(ev) => {
							minLogLevel = ev.detail;
							pagination.page = 1;
							load = getLogs();
						}}
					/>
				</div>
			</div>

			<div class="flex flex-col gap-5">
				<Table.Root {...$tableAttrs} class="min-w-[15rem] border-collapse">
					<Table.Header>
						{#each $headerRows as headerRow}
							<Subscribe rowAttrs={headerRow.attrs()}>
								<Table.Row class="hover:bg-transparent">
									{#each headerRow.cells as cell (cell.id)}
										<Subscribe attrs={cell.attrs()} let:attrs props={cell.props()}>
											<Table.Head
												{...attrs}
												class={cn(
													'relative whitespace-nowrap px-6 tracking-wide [&:has([role=checkbox])]:pl-3',
													cell.id === 'message' ? 'min-w-96' : 'min-w-[1%]'
												)}
											>
												<div
													class={cn(
														'flex select-none items-center gap-2.5',
														cell.id !== 'message' && 'justify-center'
													)}
												>
													<Render of={cell.render()} />
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
										<p class="text-muted-foreground text-center text-sm">No logs found.</p>
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							{#each $pageRows as row (row.id)}
								<Subscribe rowAttrs={row.attrs()} let:rowAttrs>
									<Table.Row {...rowAttrs} data-row={row.id}>
										{#each row.cells as cell (cell.id)}
											<Subscribe attrs={cell.attrs()} let:attrs>
												<Table.Cell
													class={cn(
														'whitespace-nowrap px-6 text-sm [&:has([role=checkbox])]:pl-3',
														cell.id === 'message' ? 'min-w-96' : 'min-w-[1%]'
													)}
													{...attrs}
												>
													<div class={cn(cell.id !== 'message' && 'text-center')}>
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
				type={'log'}
				{pagination}
				on:pageChange={(ev) => {
					pagination.page = ev.detail;
					load = getLogs();
				}}
				on:perPageChange={(ev) => {
					pagination.perPage = ev.detail;
					pagination.page = 1;
					load = getLogs();
				}}
			/>
		{:catch error}
			<Err errorMessage={error} />
		{/await}
	</div>
</div>
