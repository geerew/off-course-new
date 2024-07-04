<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Table from '$components/ui/table';
	import { DeleteTag } from '$lib/api';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------
	export let tags: Record<string, string>;
	export let open = false;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	async function deleteTags() {
		try {
			const ids = Object.keys(tags);

			await Promise.all(
				ids.map(async (id) => {
					try {
						await DeleteTag(id);
					} catch (error) {
						toast.error('Failed to delete tag: ' + tags[id]);
					}
				})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}

	// ----------------------
	// Reactive
	// ----------------------

	$: tagsCount = Object.keys(tags).length;
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="bg-muted top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md px-0 py-0 duration-200 md:max-w-md [&>button[data-dialog-close]]:hidden"
	>
		<div class="flex flex-col items-center gap-5 overflow-y-scroll px-8 pt-4">
			<Icons.WarningOctagon class="text-destructive size-10" />

			{#if tagsCount > 1}
				<span class="text-center">
					Are the sure you want to delete the following {tagsCount} tags
				</span>
			{:else}
				<div class="flex flex-col items-center gap-3">
					Are you sure you want to delete this tag?

					<span class="text-muted-foreground text-sm">
						{Object.values(tags)[0]}
					</span>
				</div>
			{/if}
		</div>

		{#if tagsCount > 1}
			<div class="flex max-h-[20rem] flex-col gap-2 overflow-hidden overflow-y-auto px-8">
				<Table.Root>
					<Table.Body>
						{#each Object.entries(tags) as [_, t], i (i)}
							<Table.Row
								class={cn(
									'border-alt-1/40 last:border-none',
									tagsCount === 1 && 'hover:bg-inherit'
								)}
							>
								<Table.Cell class="text-muted-foreground select-none text-wrap px-2.5 py-1.5">
									{t}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		{/if}

		<Dialog.Footer
			class="border-alt-1/60 h-14 flex-row items-center justify-end gap-2 border-t px-4"
		>
			<Button
				variant="outline"
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 h-8 w-20"
				on:click={() => {
					dispatch('cancelled');
					open = false;
				}}>Cancel</Button
			>
			<Button
				variant="destructive"
				class="h-8 w-20"
				on:click={async () => {
					await deleteTags();
					dispatch('deleted');
					open = false;
				}}
			>
				Yes
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
