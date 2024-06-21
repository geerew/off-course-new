<script lang="ts">
	import { Err, Loading } from '$components/generic';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { GetLogTypes } from '$lib/api';
	import { IsBrowser, cn } from '$lib/utils';
	import { ChevronRight, Type } from 'lucide-svelte';
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
			variant="ghost"
			class="data-[state=open]:border-alt-1/100 border-alt-1/60 hover:border-alt-1/100 group h-auto w-32 items-center justify-between gap-2.5 border px-2.5 text-xs hover:bg-inherit"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-2">
				<Type class={cn('size-3', Object.keys(filterTypes).length > 0 && 'text-primary')} />
				<span>Log Type</span>
			</div>

			<ChevronRight class="size-3 duration-200 group-data-[state=open]:rotate-90" />
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
