<script lang="ts">
	import { Err, Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { GetLogTypes } from '$lib/api';
	import { IsBrowser, cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let filterTypes: string[];

	// ----------------------
	// Variables
	// ----------------------

	// A boolean promise that initially fetches the tags. It is used in an `await` block
	let types = getLogTypes();

	let workingTypes: string[] = [];

	// As the selected tags change, dispatch an event
	const dispatchEvent = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	// Get all log types
	async function getLogTypes(): Promise<boolean> {
		if (!IsBrowser) return false;
		try {
			const response = await GetLogTypes();

			workingTypes = response as string[];

			return true;
		} catch (error) {
			throw error;
		}
	}
</script>

<DropdownMenu.Root closeOnItemClick={false} typeahead={false}>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="outline"
			class="border-alt-1/60 group h-auto w-36 justify-between gap-2.5 border px-2"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-2">
				<Icons.Text class={cn('size-4', Object.keys(filterTypes).length > 0 && 'text-primary')} />
				<span>Log Type</span>
			</div>

			<Icons.CaretRight class="size-3.5 duration-200 group-data-[state=open]:rotate-90" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content
		class="bg-muted text-foreground border-alt-1/60 flex w-48 flex-col"
		fitViewport={true}
		align="start"
	>
		<div class="max-h-40 overflow-y-scroll">
			{#await types}
				<Loading class="max-h-20" />
			{:then _}
				<div class="max-h-40 overflow-y-scroll">
					{#each workingTypes as t}
						<DropdownMenu.CheckboxItem
							class="data-[highlighted]:bg-alt-1/40 cursor-pointer"
							checked={filterTypes.find((type) => type === t) ? true : false}
							onCheckedChange={(checked) => {
								if (checked) {
									filterTypes = [...filterTypes, t];
								} else {
									filterTypes = filterTypes.filter((type) => type !== t);
								}

								dispatchEvent('change', t);
							}}
						>
							{t}
						</DropdownMenu.CheckboxItem>
					{/each}
				</div>
			{:catch error}
				<Err class="text-muted min-h-[6rem] p-5 text-sm" imgClass="size-6" errorMessage={error} />
			{/await}
		</div>
	</DropdownMenu.Content>
</DropdownMenu.Root>
