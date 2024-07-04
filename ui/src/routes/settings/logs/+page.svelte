<script lang="ts">
	import { Err, LogLevel as LL, Loading, LogMessage, Pagination } from '$components/generic';
	import { LogsFilter } from '$components/pages/settings_logs';
	import { preferences } from '$components/pages/settings_logs/store';
	import * as Table from '$components/ui/table';
	import { GetLogs } from '$lib/api';
	import { LogLevelMapping, type Log, type LogsGetParams } from '$lib/types/models';
	import type { PaginationParams } from '$lib/types/pagination';
	import { cn } from '$lib/utils';
	import { Render, Subscribe, createRender, createTable } from 'svelte-headless-table';
	import { writable } from 'svelte/store';

	// ----------------------
	// Variables
	// ----------------------

	// Holds the current page of logs
	const fetchedLogs = writable<Log[]>([]);

	// The messages to filter on
	let filterMessages = $preferences.messages;

	// The levels to filter on
	let filterLevels = $preferences.levels;

	// The types to filter on
	let filterTypes = $preferences.types;

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
				return createRender(LL, { level: value });
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
			const params: LogsGetParams = {
				page: pagination.page,
				perPage: pagination.perPage
			};

			if (filterLevels.length > 0) {
				params.levels = filterLevels
					.map((level) => {
						return LogLevelMapping[level];
					})
					.join(',');
			}

			if (filterMessages && filterMessages.length > 0) {
				params.messages = filterMessages.join(',');
			}

			if (filterTypes && filterTypes.length > 0) {
				params.types = filterTypes.join(',');
			}

			const response = await GetLogs(params);

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
		<LogsFilter
			{filterMessages}
			{filterLevels}
			{filterTypes}
			on:filterMessages={(ev) => {
				preferences.set({ ...$preferences, messages: ev.detail });
				filterMessages = ev.detail;
				pagination.page = 1;
				load = getLogs();
			}}
			on:filterLevels={(ev) => {
				preferences.set({ ...$preferences, levels: ev.detail });
				filterLevels = ev.detail;
				pagination.page = 1;
				load = getLogs();
			}}
			on:filterTypes={(ev) => {
				preferences.set({ ...$preferences, types: ev.detail });
				filterTypes = ev.detail;
				pagination.page = 1;
				load = getLogs();
			}}
			on:clear={() => {
				filterMessages = [];
				filterLevels = [];
				filterTypes = [];
				preferences.set({ messages: [], levels: [], types: [] });
				pagination.page = 1;
				load = getLogs();
			}}
		/>

		<div class="flex h-full w-full flex-col">
			{#await load}
				<Loading class="max-h-96" />
			{:then _}
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
										<div
											class="flex w-full flex-grow flex-col place-content-center items-center p-5"
										>
											{#if filterMessages.length > 0 || filterLevels.length > 0 || filterTypes.length > 0}
												<span class="text-muted-foreground"
													>No logs found with the selected filters.</span
												>
											{:else}
												<span class="text-muted-foreground">No logs.</span>
											{/if}
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
															cell.id === 'message'
																? 'flex min-w-96 flex-wrap whitespace-normal'
																: 'min-w-[1%]'
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
</div>
