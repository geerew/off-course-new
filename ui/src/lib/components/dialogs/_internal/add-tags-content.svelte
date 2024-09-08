<!-- TODO: Handle adding a tag that goes off screen (scroll to it?) -->
<script lang="ts">
	import Loading from '$components/generic/loading.svelte';
	import { Icons } from '$components/icons';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import { AddTag, GetTag } from '$lib/api';
	import { cn } from '$lib/utils';
	import axios from 'axios';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------

	// True when the dialog is open
	export let isOpen: boolean;

	// An array of tags that should be added to the course. This is bound to the parent component
	// so that the parent can use the tags when flipping the the type of dialog to use
	export let toAdd: string[];

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();

	let showSpinner = false;

	let tagsEl: HTMLDivElement;

	// ----------------------
	// Functions
	// ----------------------

	// Use:action for inputting tags
	const tagInput = (node: HTMLInputElement) => {
		async function handleInput(e: KeyboardEvent) {
			if (e.key === 'Enter') {
				e.preventDefault();

				if (!node.value) return;

				showSpinner = true;

				const tagToAdd = node.value.trim();

				// Check if the tag already exists in the list
				const foundTag = toAdd.find((tag) => tag.toLowerCase() === tagToAdd.toLowerCase());

				if (foundTag) {
					toast.error(`Tag '${tagToAdd}' is already added`);
					showSpinner = false;

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

				// Check if tag already exists in the backend
				try {
					await GetTag(tagToAdd, { byName: true, insensitive: true });

					toast.error(`Tag '${tagToAdd}' is an existing tag`);
					showSpinner = false;
					return;
				} catch (error) {
					if (!axios.isAxiosError(error) || (error.response && error.response.status !== 404)) {
						toast.error(error instanceof Error ? error.message : String(error));
					}
				}

				toAdd = [...toAdd, tagToAdd];
				node.value = '';

				showSpinner = false;
			}
		}

		node.addEventListener('keydown', handleInput);

		return {
			destroy() {
				node.removeEventListener('keydown', handleInput);
			}
		};
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function addTags() {
		try {
			await Promise.all(
				toAdd.map(async (tag) => {
					try {
						await AddTag(tag);
					} catch (error) {
						toast.error('Failed to add tag: ' + tag);
					}
				})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}

	$: (() => {
		if (!isOpen) {
			toAdd = [];
		}
	})();
</script>

{#if isOpen}
	<div class="group relative flex flex-row items-center border-b border-alt-1/60">
		<label class="px-5" for="add-tag-input">
			<Icons.Search class="size-6 text-muted-foreground" />
		</label>

		<input
			type="text"
			id="add-tag-input"
			use:tagInput
			placeholder="Add tag..."
			class="h-14 w-full rounded-none border-none bg-inherit px-0 text-foreground placeholder-muted-foreground/60 focus-visible:outline-none focus-visible:ring-0"
		/>

		<Loading
			class={cn('absolute right-3 h-auto min-h-0 w-auto p-0', !showSpinner && 'hidden')}
			loaderClass="size-5"
		/>
	</div>

	<main
		class="flex max-h-[20rem] min-h-[7rem] flex-col gap-2 overflow-hidden overflow-y-auto px-4"
		data-vaul-no-drag=""
	>
		<div class="flex flex-row flex-wrap gap-2.5" bind:this={tagsEl}>
			{#each toAdd as tag}
				<div class="flex flex-row" data-tag={tag}>
					<!-- Tag -->
					<Badge
						class={cn(
							'min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none border-none bg-success text-sm text-success-foreground hover:bg-success'
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
							// When its a newly added tag, just delete it completely
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
				disabled={toAdd.length === 0}
				on:click={async () => {
					await addTags();
					dispatch('added');
					toAdd = [];
					isOpen = false;
				}}
			>
				Add {toAdd.length > 0 ? `(${toAdd.length})` : ''}
			</Button>
		</div>
	</footer>
{/if}
