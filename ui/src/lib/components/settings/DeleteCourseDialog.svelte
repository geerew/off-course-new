<script lang="ts">
	import { Icons } from '$components/icons';
	import { addToast } from '$lib/stores/addToast';
	import { DeleteCourse } from '$lib/utils/api';
	import type { createDialog } from '@melt-ui/svelte';
	import { createEventDispatcher } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	// ----------------------
	// Exports
	// ----------------------
	export let dialog: ReturnType<typeof createDialog>;
	export let id: string;

	// ----------------------
	// Variables
	// ----------------------
	const {
		elements: { portalled, overlay, content, description, close },
		states: { open }
	} = dialog;

	const dispatch = createEventDispatcher();
</script>

<div {...$portalled} use:portalled>
	{#if $open}
		<div
			{...$overlay}
			use:overlay
			transition:fade|local={{ duration: 150 }}
			class="fixed inset-0 bg-black/60 backdrop-blur-sm"
			tabindex="-1"
		/>

		<div
			class="bg-background fixed left-1/2 top-1/2 z-50 w-[90vw] max-w-[30rem] -translate-x-1/2 -translate-y-1/2 rounded-md border"
			transition:fly={{ y: -150, duration: 400 }}
			{...$content}
			use:content
		>
			<div
				{...$description}
				use:description
				class="flex min-w-[20rem] max-w-xl grow flex-col items-center gap-5 overflow-y-scroll p-5"
			>
				<Icons.warningCircle class="text-error h-14 w-14" />
				Do you really want to delete this course and all its data?
			</div>
			<footer class="border-border/50 flex items-center justify-end gap-3 border p-5 font-medium">
				<button
					{...$close}
					use:close
					class="hover:bg-accent-1 inline-flex w-24 select-none items-center justify-center gap-2 whitespace-nowrap rounded border px-3 py-1.5 text-center"
					>No</button
				>

				<button
					class="bg-error inline-flex w-24 select-none items-center justify-center gap-2 whitespace-nowrap rounded border !border-none px-3 py-1.5 text-center text-white hover:brightness-110"
					on:click={async () => {
						await DeleteCourse(id)
							.then(() => {
								$addToast({
									data: {
										message: `Deleted course`,
										status: 'success'
									}
								});

								dispatch('confirmed');
							})
							.catch((err) => {
								console.error(err);
							})
							.finally(() => {
								open.set(false);
							});
					}}
				>
					Yes
				</button>
			</footer>
		</div>
	{/if}
</div>
