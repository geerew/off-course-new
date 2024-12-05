<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Table from '$components/ui/table';
	import { DeleteCourse } from '$lib/api';
	import { cn } from '$lib/utils';
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

	// ----------------------
	// Reactive
	// ----------------------

	$: coursesCount = Object.keys(courses).length;
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md bg-muted px-0 py-0 duration-200 md:max-w-xl [&>button[data-dialog-close]]:hidden"
	>
		<div class="flex flex-col items-center gap-5 overflow-y-scroll px-8 pt-4">
			<Icons.WarningOctagon class="size-10 text-destructive" />

			{#if coursesCount > 1}
				<span class="text-center">
					Are the sure you want to delete the following {coursesCount} courses?
				</span>
			{:else}
				<div class="flex flex-col items-center gap-3">
					Are you sure you want to delete this course?

					<span class="text-sm text-muted-foreground">
						{Object.values(courses)[0]}
					</span>
				</div>
			{/if}
		</div>

		{#if coursesCount > 1}
			<div class="flex max-h-[20rem] flex-col gap-2 overflow-hidden overflow-y-auto px-8">
				<Table.Root>
					<Table.Body>
						{#each Object.entries(courses) as [_, c], i (i)}
							<Table.Row
								class={cn(
									'border-alt-1/40 last:border-none',
									coursesCount === 1 && 'hover:bg-inherit'
								)}
							>
								<Table.Cell class="select-none text-wrap px-2.5 py-1.5 text-muted-foreground">
									{c}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		{/if}

		<Dialog.Footer
			class="h-14 flex-row items-center justify-end gap-2 border-t border-alt-1/60 px-4"
		>
			<Button
				variant="outline"
				class="h-8 w-20 border-alt-1/60 bg-muted hover:bg-alt-1/60"
				on:click={() => {
					dispatch('cancelled');
					open = false;
				}}>Cancel</Button
			>

			<Button
				variant="destructive"
				class="h-8 w-20"
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
