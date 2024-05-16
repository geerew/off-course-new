<script lang="ts">
	import { Err, Loading } from '$components/generic';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { GetAllTags } from '$lib/api';
	import type { Tag } from '$lib/types/models';
	import { IsBrowser } from '$lib/utils';
	import { ChevronRight, Search, Tag as TagIcon, X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let filterTags: Record<string, string>;

	// ----------------------
	// Variables
	// ----------------------

	// A boolean promise that initially fetches the tags. It is used in an `await` block
	let tags = getTags();

	// A list of all tags
	let allTags: Tag[] = [];

	// A list of tags that are currently being displayed. When the search input is empty, this will be equal to `allTags`, otherwise it will be a filtered version of `allTags`
	let workingTags: Tag[] = [];

	// As the selected tags change, dispatch an event
	const dispatchEvent = createEventDispatcher();

	// Bound to the search input
	let searchValue = '';

	// ----------------------
	// Functions
	// ----------------------

	// Get all tags
	async function getTags(): Promise<boolean> {
		if (!IsBrowser) return false;
		try {
			const response = await GetAllTags();

			allTags = response as Tag[];
			workingTags = allTags;

			return true;
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Filter tags
	function doFilter(value: string) {
		if (value === '') {
			workingTags = allTags;
		} else {
			workingTags = allTags.filter((tag) => tag.tag.toLowerCase().includes(value.toLowerCase()));
		}
	}

	// ----------------------
	// Reactive
	// ----------------------

	// As the search value changes, filter the tags
	$: doFilter(searchValue);
</script>

<DropdownMenu.Root
	closeOnItemClick={false}
	typeahead={false}
	onOpenChange={(open) => {
		if (!open) searchValue = '';
	}}
>
	<DropdownMenu.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="ghost"
			class="data-[state=open]:border-alt-1/100 border-alt-1/60 hover:border-alt-1/100 group h-auto w-32 items-center justify-between gap-2.5 border px-2.5 text-xs hover:bg-inherit"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-1.5">
				<TagIcon class="size-3" />
				<span>Tags</span>
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
			{#await tags}
				<Loading class="max-h-20" />
			{:then _}
				{#if allTags.length === 0}
					<div
						class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10"
					>
						<span class="text-muted-foreground">No tags.</span>
					</div>
				{:else}
					<div class="relative mb-1.5">
						<label for="tags-input">
							<Search
								class="text-muted-foreground absolute left-2 top-1/2 size-3 -translate-y-1/2"
							/>
						</label>

						<input
							id="tags-input"
							class="placeholder-muted-foreground/60 text-foreground bg-background border-alt-1/60 w-full rounded-md border border-none px-7 text-sm focus-visible:outline-none focus-visible:ring-0"
							placeholder="Search tags"
							bind:value={searchValue}
						/>

						{#if searchValue.length > 0}
							<Button
								class="text-muted-foreground hover:text-foreground absolute right-1 top-1/2 h-auto -translate-y-1/2 transform px-2 py-1 hover:bg-inherit"
								variant="ghost"
								on:click={() => {
									searchValue = '';
								}}
							>
								<X class="size-3" />
							</Button>
						{/if}
					</div>

					{#if workingTags.length === 0}
						<div
							class="flex min-h-[6rem] w-full flex-grow flex-col place-content-center items-center p-10"
						>
							<span class="text-muted-foreground">No tags.</span>
						</div>
					{:else}
						{#each workingTags as tag}
							<DropdownMenu.CheckboxItem
								class="data-[highlighted]:bg-alt-1/40 cursor-pointer"
								checked={filterTags[tag.id] ? true : false}
								onCheckedChange={(checked) => {
									if (checked) {
										filterTags[tag.id] = tag.tag;
									} else {
										delete filterTags[tag.id];
									}

									filterTags = { ...filterTags };
									dispatchEvent('change');
								}}
							>
								{tag.tag}
							</DropdownMenu.CheckboxItem>
						{/each}
					{/if}
				{/if}
			{:catch error}
				<Err class="text-muted min-h-[6rem] p-5 text-sm" imgClass="size-6" errorMessage={error} />
			{/await}
		</div>
	</DropdownMenu.Content>
</DropdownMenu.Root>

<!--  -->
