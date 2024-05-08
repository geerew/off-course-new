<script lang="ts">
	import { Loading } from '$components/generic';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Tooltip from '$components/ui/tooltip';
	import { AddCourseTag, DeleteCourseTag, GetTags } from '$lib/api';
	import type { CourseTag, Tag } from '$lib/types/models';
	import { cn, flyAndScale } from '$lib/utils';
	import { createCombobox, type ComboboxOption } from '@melt-ui/svelte';
	import { ArrowLeft, Pencil, RotateCcw, Search, X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { writable, type Writable } from 'svelte/store';
	import { fly } from 'svelte/transition';

	// ----------------------
	// Exports
	// ----------------------

	export let courseId: string;
	export let existingTags: CourseTag[];

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();

	// True when the dialog is open. This is bound to the dialog component
	let isDialogOpen = false;

	// An array of tags that should be added to the course
	let toAdd: string[] = [];

	// True when the combobox is open. This is used to show a spinner when backend events are happening
	let showSpinner = false;

	// The elements containing the edited/added tags
	let tagsEl: HTMLDivElement;

	// Every time the tags are added or existing tag is deleted, this counted will increment. When the
	// inverse happens, it the counter will decrement. This can be used to enable/disable parts of the UI
	// while the counter is 0
	let changes = 0;

	// This will be populated from a filtered list of tags will be fetched from the backend
	let filteredTags: Tag[] = [];

	// The selected tag from the combobox. This is bound to the combobox component
	let selected: Writable<ComboboxOption<string>> = writable({ value: '', label: '' });

	// A debounce timer to prevent the backend from being called too often
	let debounceTimer: ReturnType<typeof setTimeout>;

	// True when the tag can be appended to `toAdd`
	let canAppendTag = false;

	const {
		elements: { menu, input, option, label },
		states: { open: isComboOpen, inputValue }
	} = createCombobox<string>({
		selected,
		loop: false
	});

	// ----------------------
	// Functions
	// ----------------------

	// Append the tags to the list of tags to be added to the course
	function appendTag(tag: string) {
		if (!tag || !tag.trim()) return;

		const foundTag =
			toAdd.find((t) => t.toLowerCase() === tag.toLowerCase()) ||
			existingTags.find((t) => t.tag.toLowerCase() === tag.toLowerCase());

		if (foundTag) {
			toast.error(`Tag '${tag}' already added`);
			inputValue.set(tag);
			selected.set({ value: '', label: '' });
			isComboOpen.set(true);

			if (tagsEl) {
				const tagEl = tagsEl.querySelector(`[data-tag="${foundTag}"]`);
				if (tagEl) {
					// shake the tag if not already shaking
					if (tagEl.classList.contains('animate-shake')) return;

					tagEl.classList.add('animate-shake');
					setTimeout(() => {
						tagEl.classList.remove('animate-shake');
					}, 1000);
				}
			}

			return;
		}

		// Append and increment the number of changes
		toAdd = [...toAdd, tag];
		changes++;

		// Clear some things out
		selected.set({ value: '', label: '' });
		inputValue.set('');
		filteredTags = [];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// A debounce function to prevent the backend from being called too often
	function debounce(callback: () => void) {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(callback, 350);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Get a filtered array of tags from the backend
	async function getFilteredTags(input: string) {
		input = input.trim();

		// Do nothing when the input is empty
		if (input === '') {
			selected.set({ value: '', label: '' });
			filteredTags = [];
			isComboOpen.set(false);
			return;
		}

		debounce(async () => {
			showSpinner = true;

			try {
				const response = await GetTags({ filter: input, perPage: 10 });

				const respTags = response.items as Tag[];

				if (respTags.length === 0) {
					filteredTags = [];
					return;
				}

				// If the input has changed since the backend call was made and is now empty, do
				// nothing
				if (!$inputValue) return;

				// Set the filtered tags
				filteredTags = respTags;

				isComboOpen.set(true);
			} catch (error) {
				toast.error(error instanceof Error ? error.message : (error as string));
				throw error;
			} finally {
				showSpinner = false;
			}
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Backend -> Add the new tags to the course
	async function addTags() {
		try {
			await Promise.all(
				toAdd.map(async (tag) => {
					try {
						await AddCourseTag(courseId, tag);
					} catch (error) {
						toast.error('Failed to add tag: ' + tag);
					}
				})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Backend -> Delete the tags that are marked for deletion from the course
	async function deleteTags() {
		try {
			await Promise.all(
				existingTags
					.filter((tag) => tag.forDeletion)
					.map(async (tag) => {
						try {
							await DeleteCourseTag(courseId, tag.id);
						} catch (error) {
							toast.error('Failed to delete tag: ' + tag.tag);
						}
					})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function isDisabled(tag: string): boolean {
		return (
			toAdd.find((t) => t.toLowerCase() === tag.toLowerCase()) !== undefined ||
			existingTags.find((t) => t.tag.toLowerCase() === tag.toLowerCase()) !== undefined
		);
	}

	// ----------------------
	// Reactive
	// ----------------------

	// As the `inputValue` changes, fetch the filtered tags from the backend
	$: (async () => {
		await getFilteredTags($inputValue);
	})();

	// When the selected tag changes, append the tag to the list of tags to be added. `canAppendTag` is set to true
	// when the user presses enter or selects a tag from the combobox
	$: if (canAppendTag && $selected && $selected.value !== '') {
		canAppendTag = false;
		appendTag($selected.value);
	}

	// Reset the state when the dialog is closed
	$: if (!isDialogOpen) {
		toAdd = [];
		changes = 0;
		filteredTags = [];
		showSpinner = false;
		inputValue.set('');
		selected.set({ value: '', label: '' });

		existingTags.forEach((tag) => {
			tag.forDeletion = false;
		});
	}
</script>

<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
	<Tooltip.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="ghost"
			class="text-muted-foreground hover:text-foreground h-auto cursor-pointer px-2.5 py-1"
			on:click={() => {
				isDialogOpen = true;
			}}
		>
			<Pencil class="h-4 w-4" />
		</Button>
	</Tooltip.Trigger>

	<Tooltip.Content
		class="bg-foreground text-background select-none rounded-sm border-none px-1.5 py-1 text-xs"
		transition={flyAndScale}
		transitionConfig={{ y: 8, duration: 100 }}
		side="bottom"
	>
		Edit
		<Tooltip.Arrow class="bg-background" />
	</Tooltip.Content>
</Tooltip.Root>

<Dialog.Root bind:open={isDialogOpen} closeOnEscape={false} closeOnOutsideClick={false}>
	<Dialog.Content
		class="bg-muted top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md px-0 py-0 duration-200 md:max-w-xl [&>button[data-dialog-close]]:hidden"
	>
		<!-- Input -->
		<div class="border-alt-1/60 group relative flex flex-row items-center border-b">
			<!-- svelte-ignore a11y-label-has-associated-control - $label contains the 'for' attribute -->
			<label {...$label} use:label>
				<Search class="text-muted-foreground absolute start-3 top-1/2 size-6 -translate-y-1/2" />
			</label>

			<input
				{...$input}
				use:input
				class="placeholder-muted-foreground/60 text-foreground h-14 w-full rounded-none border-none bg-inherit px-14 focus-visible:outline-none focus-visible:ring-0"
				placeholder="Enter a tag..."
				on:m-keydown={(e) => {
					if (e.detail.originalEvent.key === 'Enter') {
						canAppendTag = true;
						selected.set({ value: $inputValue, label: $inputValue });
					}
				}}
			/>

			<Loading
				class={cn('absolute right-3 h-auto min-h-0 w-auto p-0', !showSpinner && 'hidden')}
				loaderClass="size-5"
			/>
		</div>

		<!-- Popup for input -->
		{#if $isComboOpen && filteredTags.length > 0}
			<div class=" z-50" {...$menu} use:menu transition:fly={{ duration: 150, y: -5 }}>
				<div class="bg-background ml-10 mr-2 gap-1.5 rounded-b-md py-2">
					<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
					{#each filteredTags as t (t.tag)}
						<li
							{...$option({ value: t.tag, label: t.tag, disabled: isDisabled(t.tag) })}
							use:option
							class="rounded-button data-[highlighted]:bg-muted/60 data-[disabled]:text-muted-foreground/70 flex h-10 w-full cursor-pointer select-none items-center p-3 text-sm outline-none transition-all duration-75 data-[disabled]:cursor-auto"
							on:m-click={() => {
								canAppendTag = true;
							}}
						>
							{t.tag}

							{#if t.tag.toLowerCase() === $inputValue.toLowerCase()}
								<div class="ml-auto">
									<ArrowLeft class="size-3" />
								</div>
							{/if}
						</li>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Body -->
		<div
			class="flex max-h-[20rem] min-h-[7rem] flex-col gap-2 overflow-hidden overflow-y-auto px-4"
		>
			<div class="flex flex-row flex-wrap gap-2.5" bind:this={tagsEl}>
				{#each existingTags as tag}
					<div class="flex flex-row" data-tag={tag.tag}>
						<!-- Tag -->
						<Badge
							class={cn(
								'bg-alt-1 hover:bg-alt-1 text-foreground min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none',
								tag.forDeletion && 'text-destructive-foreground opacity-60'
							)}
						>
							{tag.tag}
						</Badge>

						<Button
							class={cn(
								'bg-alt-1 hover:bg-destructive inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5 duration-200',
								toAdd.includes(tag.tag) && 'bg-success text-success-foreground',
								tag.forDeletion &&
									'text-destructive-foreground hover:bg-alt-1 opacity-60 transition-opacity hover:opacity-100'
							)}
							on:click={() => {
								changes += tag.forDeletion ? 1 : -1;
								tag.forDeletion = !tag.forDeletion;
							}}
						>
							<svelte:component this={tag.forDeletion ? RotateCcw : X} class="size-3" />
						</Button>
					</div>
				{/each}

				{#each toAdd as tag}
					<div class="flex flex-row" data-tag={tag}>
						<!-- Tag -->
						<Badge
							class={cn(
								'bg-success text-success-foreground hover:bg-success min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none'
							)}
						>
							{tag}
						</Badge>

						<!-- Delete button -->
						<Button
							class={cn(
								'hover:bg-destructive bg-success text-success-foreground inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5'
							)}
							on:click={() => {
								toAdd = toAdd.filter((t) => t !== tag);
								changes--;
							}}
						>
							<X class="size-3" />
						</Button>
					</div>
				{/each}
			</div>
		</div>

		<Dialog.Footer
			class="border-alt-1/60 h-14 flex-row items-center justify-end gap-2 border-t px-4"
		>
			<Button
				variant="outline"
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 h-8 w-20"
				on:click={() => {
					isDialogOpen = false;
				}}
			>
				Cancel
			</Button>

			<Button
				class="h-8 px-6"
				disabled={changes === 0}
				on:click={async () => {
					await addTags();
					await deleteTags();
					dispatch('updated');
					isDialogOpen = false;
				}}
			>
				Save
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
