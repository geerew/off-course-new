<script lang="ts">
	import { Loading } from '$components/generic';
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import { GetTag, UpdateTag } from '$lib/api';
	import type { Tag } from '$lib/types/models';
	import axios from 'axios';
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
		class="top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md bg-muted px-0 py-0 duration-200 md:max-w-md [&>button[data-dialog-close]]:hidden"
	>
		<div class="flex flex-col gap-5 overflow-y-scroll px-8 pt-4">
			<div class="flex flex-row items-center gap-2.5">
				<Icons.Edit class="size-4" />
				<span>Rename Tag</span>
			</div>

			<div class="flex flex-col items-center gap-3">
				<span
					class="flex h-10 items-center rounded-md border border-alt-1/60 px-3 text-sm text-muted-foreground"
				>
					{tag.tag}
				</span>

				<span>to</span>

				<input
					id="rename-tag"
					bind:this={inputEl}
					bind:value={newTagName}
					class="h-10 w-full rounded-md border border-alt-1/60 bg-inherit text-foreground placeholder-muted-foreground/60 focus-visible:border-alt-1/60 focus-visible:outline-none focus-visible:ring-0"
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
			class="h-14 flex-row items-center justify-end gap-2 border-t border-alt-1/60 px-4"
		>
			<Button
				variant="outline"
				class="h-8 w-20 border-alt-1/60 bg-muted hover:bg-alt-1/60"
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
