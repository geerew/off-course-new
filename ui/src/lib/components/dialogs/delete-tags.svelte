<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import { DeleteTag } from '$lib/api';
	import { AlertOctagon } from 'lucide-svelte';
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
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="min-w-[20rem] max-w-[425px] md:max-w-[30rem]">
		<div class="flex flex-col gap-2 overflow-y-scroll px-4 py-2">
			<AlertOctagon class="text-destructive size-10 w-full text-center" />
			<span>Do you really want to delete the following tags?</span>
		</div>

		<div class="flex max-h-[20rem] flex-col gap-2 overflow-hidden overflow-y-auto px-4">
			<ul class="list-inside">
				{#each Object.entries(tags) as [id, name]}
					<li class="text-muted-foreground list-disc">
						<span class="text-muted-foreground select-none">{name}</span>
					</li>
				{/each}
			</ul>
		</div>

		<Dialog.Footer class="gap-2">
			<Button variant="outline" class="px-6" on:click={() => (open = false)}>No</Button>
			<Button
				variant="destructive"
				class="px-6"
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
