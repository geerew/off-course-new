<script lang="ts">
	import { Err, Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import { GetAllTags } from '$lib/api';
	import type { Tag as TagModel } from '$lib/types/models';
	import { IsBrowser, cn } from '$lib/utils';
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
	let allTags: TagModel[] = [];

	// A list of tags that are currently being displayed. When the search input is empty, this will be equal to `allTags`, otherwise it will be a filtered version of `allTags`
	let workingTags: TagModel[] = [];

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

			allTags = response as TagModel[];
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
			class="group h-auto w-32 items-center justify-between gap-2.5 border border-alt-1/60 px-2.5 text-xs hover:border-alt-1/100 hover:bg-inherit data-[state=open]:border-alt-1/100"
			on:click={(e) => {
				e.stopPropagation();
			}}
		>
			<div class="flex items-center gap-2">
				<Icons.Tag class={cn('size-3', Object.keys(filterTags).length > 0 && 'text-primary')} />
				<span>Tags</span>
			</div>

			<Icons.CaretRight class="size-3 duration-200 group-data-[state=open]:rotate-90" />
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content
		class="flex w-48 flex-col border-alt-1/60 bg-muted text-foreground"
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
							<Icons.Search
								class="absolute left-2 top-1/2 size-3 -translate-y-1/2 text-muted-foreground"
							/>
						</label>

						<input
							id="tags-input"
							class="w-full rounded-md border border-none border-alt-1/60 bg-background px-7 text-sm text-foreground placeholder-muted-foreground/60 focus-visible:outline-none focus-visible:ring-0"
							placeholder="Search tags"
							bind:value={searchValue}
						/>

						{#if searchValue.length > 0}
							<Button
								class="absolute right-1 top-1/2 h-auto -translate-y-1/2 transform px-2 py-1 text-muted-foreground hover:bg-inherit hover:text-foreground"
								variant="ghost"
								on:click={() => {
									searchValue = '';
								}}
							>
								<Icons.X class="size-3" />
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
								class="cursor-pointer data-[highlighted]:bg-alt-1/40"
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
				<Err class="min-h-[6rem] p-5 text-sm text-muted" imgClass="size-6" errorMessage={error} />
			{/await}
		</div>
	</DropdownMenu.Content>
</DropdownMenu.Root>
