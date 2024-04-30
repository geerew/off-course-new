<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Table from '$components/ui/table';
	import { DeleteCourse } from '$lib/api';
	import { AlertOctagon } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------
	export let courses: Record<string, string>;
	export let open = false;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	async function deleteCourses() {
		try {
			const ids = Object.keys(courses);

			await Promise.all(
				ids.map(async (id) => {
					try {
						await DeleteCourse(id);
					} catch (error) {
						toast.error('Failed to delete course: ' + courses[id]);
					}
				})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="min-w-[20rem] max-w-[425px] px-0 py-4 md:top-20 md:max-w-lg md:translate-y-0"
	>
		<div class="flex flex-col items-center gap-5 overflow-y-scroll px-8 pt-4">
			<AlertOctagon class="text-destructive size-10" />
			<span>Do you really want to delete the following courses?</span>
		</div>

		<div class="flex max-h-[20rem] flex-col gap-2 overflow-hidden overflow-y-auto px-8">
			<Table.Root>
				<Table.Body>
					{#each Object.entries(courses) as [_, c], i (i)}
						<Table.Row class="last:border-none">
							<Table.Cell class="text-muted-foreground select-none px-4 py-1.5">{c}</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		<Dialog.Footer class="gap-2 border-t px-4 pt-4">
			<Button
				variant="outline"
				class="w-20"
				on:click={() => {
					dispatch('cancelled');
					open = false;
				}}
			>
				Cancel
			</Button>

			<Button
				variant="destructive"
				class="w-20"
				on:click={async () => {
					await deleteCourses();
					dispatch('deleted');
					open = false;
				}}
			>
				Yes
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
