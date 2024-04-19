<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import { DeleteCourse } from '$lib/api';
	import { AlertOctagon } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;
	export let open = false;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	async function deleteCourse() {
		try {
			await DeleteCourse(courseId);
			dispatch('deleted');
			toast.success('Deleted course');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		} finally {
			open = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-[425px]">
		<div class="flex min-w-[20rem] grow flex-col items-center gap-5 overflow-y-scroll p-5">
			<AlertOctagon class="text-destructive h-14 w-14" />
			<span> Do you really want to delete this course and all its data? </span>
		</div>

		<Dialog.Footer>
			<Button variant="outline" class="px-6" on:click={() => (open = false)}>No</Button>
			<Button variant="destructive" class="px-6" on:click={deleteCourse}>Yes</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
