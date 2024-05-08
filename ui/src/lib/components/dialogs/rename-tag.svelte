<script lang="ts">
	import { Loading } from '$components/generic';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import { GetTag, UpdateTag } from '$lib/api';
	import type { Tag } from '$lib/types/models';
	import axios from 'axios';
	import { SquarePen } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------
	export let tag: Tag;
	export let open = false;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();

	// Bound to the input element. The new tag name
	let newTagName: string = '';

	// Used to focus the input element
	let inputEl: HTMLInputElement;

	// True when the tag is being saved
	let isSaving = false;

	// ----------------------
	// Functions
	// ----------------------

	// Returns false if the tag cannot be renamed, else true
	async function canRename() {
		try {
			await GetTag(newTagName, { byName: true });

			// Tag already exists
			toast.error(`Tag '${newTagName}' already exists`);
			return false;
		} catch (error) {
			if (!axios.isAxiosError(error) || (error.response && error.response.status !== 404)) {
				toast.error(error instanceof Error ? error.message : (error as string));
				return false;
			}

			return true;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Renames a tag
	async function renameTag() {
		if (newTagName.trim() === '') return;

		isSaving = true;

		if (!canRename()) return false;

		try {
			tag.tag = newTagName;
			await UpdateTag(tag);
			return true;
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
			return false;
		} finally {
			isSaving = false;
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		if (inputEl) inputEl.focus();
	});
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="bg-muted top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md px-0 py-0 duration-200 md:max-w-md [&>button[data-dialog-close]]:hidden"
	>
		<div class="flex flex-col gap-5 overflow-y-scroll px-8 pt-4">
			<div class="flex flex-row items-center gap-2.5">
				<SquarePen class="size-4" />
				<span>Rename Tag</span>
			</div>

			<div class="flex flex-col items-center gap-3">
				<span
					class="border-alt-1/60 text-muted-foreground flex h-10 items-center rounded-md border px-3 text-sm"
				>
					{tag.tag}
				</span>

				<span>to</span>

				<input
					id="rename-tag"
					bind:this={inputEl}
					bind:value={newTagName}
					class="placeholder-muted-foreground/60 text-foreground border-alt-1/60 focus-visible:border-alt-1/60 h-10 w-full rounded-md border bg-inherit focus-visible:outline-none focus-visible:ring-0"
					placeholder="..."
					on:keydown={async (e) => {
						if (e.key === 'Enter') {
							e.preventDefault();
							if (newTagName.trim() === '' || isSaving) return;

							if (!(await renameTag())) return;
							dispatch('renamed');
							open = false;
						}
					}}
				/>
			</div>
		</div>

		<Dialog.Footer
			class="border-alt-1/60 h-14 flex-row items-center justify-end gap-2 border-t px-4"
		>
			<Button
				variant="outline"
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 h-8 w-20"
				on:click={() => {
					dispatch('cancelled');
					open = false;
				}}
			>
				Cancel
			</Button>
			<Button
				class="h-8 w-20"
				disabled={newTagName.trim() === '' || isSaving}
				on:click={async () => {
					if (!(await renameTag())) return;
					dispatch('renamed');
					open = false;
				}}
			>
				{#if isSaving}
					<Loading loaderClass="text-muted/70 size-5" />
				{:else}
					Save
				{/if}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
