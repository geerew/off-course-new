<script lang="ts">
	import { Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import { AddCourseTag, DeleteCourseTag, GetTag, GetTags } from '$lib/api';
	import type { CourseTag, Tag } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { createCombobox, type ComboboxOption } from '@melt-ui/svelte';
	import axios from 'axios';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { writable, type Writable } from 'svelte/store';
	import { fly } from 'svelte/transition';

	// ----------------------
	// Exports
	// ----------------------

	// The course id to add tags to
	export let courseId: string;

	// The existing tags for the course
	export let existingTags: CourseTag[] = [];

	// True when the dialog is open
	export let isOpen: boolean;

	// An array of tags that should be added to the course. This is bound to the parent component
	// so that the parent can use the tags when flipping the the type of dialog to use
	export let toAdd: string[];

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();

	// True when the combobox is open. This is used to show a spinner when backend events are happening
	let showSpinner = false;

	// The elements containing the edited/added tags
	let tagsEl: HTMLDivElement;

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
	async function appendTag(tag: string) {
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
					if (tagEl.classList.contains('animate-shake')) return;

					tagEl.classList.add('animate-shake');
					setTimeout(() => {
						tagEl.classList.remove('animate-shake');
					}, 1000);
				}
			}

			return;
		}

		// Check if the tag exists in the backend with a case-insensitive search. If it does, use
		// the tag from the backend
		try {
			const response = await GetTag(tag, { byName: true, insensitive: true });

			if (response) {
				tag = response.tag;
			}
		} catch (error) {
			if (axios.isAxiosError(error) && error.response && error.response.status !== 404) {
				toast.error(error instanceof Error ? error.message : (error as string));
				throw error;
			}
		}

		// Append and increment the number of changes
		toAdd = [...toAdd, tag];

		// Clear some things out
		selected.set({ value: '', label: '' });
		inputValue.set('');
		filteredTags = [];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// A debounce function to prevent the backend from being called too often
	function debounce(callback: () => void) {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(callback, 250);
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
	$: if (!isOpen) {
		toAdd = [];
		filteredTags = [];
		showSpinner = false;
		inputValue.set('');
		selected.set({ value: '', label: '' });

		existingTags.forEach((tag) => {
			tag.forDeletion = false;
		});
	}
</script>

{#if isOpen}
	<!-- Input -->
	<div class="group relative flex flex-row items-center border-b border-alt-1/60">
		<!-- svelte-ignore a11y-label-has-associated-control - $label contains the 'for' attribute -->
		<label {...$label} use:label>
			<Icons.Search
				class="absolute start-3 top-1/2 size-6 -translate-y-1/2 text-muted-foreground"
			/>
		</label>

		<input
			{...$input}
			use:input
			class="h-14 w-full rounded-none border-none bg-inherit px-14 text-foreground placeholder-muted-foreground/60 focus-visible:outline-none focus-visible:ring-0"
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
			<div class="ml-10 mr-2 gap-1.5 rounded-b-md bg-background py-2">
				<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
				{#each filteredTags as t (t.tag)}
					<li
						{...$option({ value: t.tag, label: t.tag, disabled: isDisabled(t.tag) })}
						use:option
						class="rounded-button flex h-10 w-full cursor-pointer select-none items-center p-3 text-sm outline-none transition-all duration-75 data-[disabled]:cursor-auto data-[highlighted]:bg-muted/60 data-[disabled]:text-muted-foreground/70"
						on:m-click={() => {
							canAppendTag = true;
						}}
					>
						{t.tag}

						{#if t.tag.toLowerCase() === $inputValue.toLowerCase()}
							<div class="ml-auto">
								<Icons.ArrowLeft class="size-3" />
							</div>
						{/if}
					</li>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Body -->
	<main
		class="flex max-h-[20rem] min-h-[7rem] flex-col gap-2 overflow-hidden overflow-y-auto px-4"
		data-vaul-no-drag=""
	>
		<div class="flex flex-row flex-wrap gap-2.5" bind:this={tagsEl}>
			{#each existingTags as tag}
				<div class="flex flex-row" data-tag={tag.tag}>
					<!-- Tag -->
					<Badge
						class={cn(
							'min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1 text-foreground hover:bg-alt-1',
							tag.forDeletion && 'text-destructive-foreground opacity-60'
						)}
					>
						{tag.tag}
					</Badge>

					<Button
						class={cn(
							'inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l bg-alt-1 px-1.5 py-0.5 duration-200 hover:bg-destructive',
							toAdd.includes(tag.tag) && 'bg-success text-success-foreground',
							tag.forDeletion &&
								'text-destructive-foreground opacity-60 transition-opacity hover:bg-alt-1 hover:opacity-100'
						)}
						on:click={() => {
							tag.forDeletion = !tag.forDeletion;
						}}
					>
						<svelte:component
							this={tag.forDeletion ? Icons.ArrowCounterClockwise : Icons.X}
							class="size-3"
						/>
					</Button>
				</div>
			{/each}

			{#each toAdd as tag}
				<div class="flex flex-row" data-tag={tag}>
					<!-- Tag -->
					<Badge
						class={cn(
							'min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none bg-success text-success-foreground hover:bg-success'
						)}
					>
						{tag}
					</Badge>

					<!-- Delete button -->
					<Button
						class={cn(
							'inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l bg-success px-1.5 py-0.5 text-success-foreground hover:bg-destructive'
						)}
						on:click={() => {
							toAdd = toAdd.filter((t) => t !== tag);
						}}
					>
						<Icons.X class="size-3" />
					</Button>
				</div>
			{/each}
		</div>
	</main>

	<footer class="h-14 gap-3 overflow-y-auto border-t border-alt-1/60 px-3">
		<div class="flex h-full flex-row items-center justify-end gap-4">
			<Button
				variant="outline"
				class="h-8 w-20 border-alt-1/60 bg-muted hover:bg-alt-1/60"
				on:click={() => (isOpen = false)}
			>
				Cancel
			</Button>

			<Button
				class="h-8 px-6"
				disabled={toAdd.length === 0 && existingTags.filter((tag) => tag.forDeletion).length === 0}
				on:click={async () => {
					await addTags();
					await deleteTags();
					dispatch('updated');
					isOpen = false;
				}}
			>
				Save
			</Button>
		</div>
	</footer>
{/if}
