<!-- TODO: Handle adding a tag that goes off screen (scroll to it?) -->
<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Drawer from '$components/ui/drawer';
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import AddTagsContent from './_internal/add-tags-content.svelte';

	// ----------------------
	// Variables
	// ----------------------

	// True when the dialog is open
	let isOpen = false;

	// An array of tags that should be added to the course. When first opened this will be empty and
	// when the dialog is closed it will be reset to empty. If the user resizes the window while the dialog
	// is open, the tags will remain in the array and passed to the new dialog
	let toAdd: string[] = [];

	// The breakpoint for md
	const mdPx = +theme.screens.md.replace('px', '');

	// True when the window size is < md. Set once the window size is known, which happens in onMount
	let isMobile: boolean | null = null;

	// ----------------------
	// Reactive
	// ----------------------

	$: if (!isOpen) {
		toAdd = [];
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		isMobile = window.innerWidth < mdPx;
		window.addEventListener('resize', () => {
			isMobile = window.innerWidth < mdPx;
		});
	});
</script>

<Button
	variant="outline"
	class="group flex h-8 gap-1.5 bg-primary hover:bg-primary hover:brightness-110"
	on:click={() => (isOpen = true)}
>
	<Icons.Tag class="size-4" />
	<span>Add Tags</span>
</Button>

{#if isMobile !== null}
	{#if isMobile}
		<Drawer.Root bind:open={isOpen} closeOnOutsideClick={false} closeOnEscape={false}>
			<Drawer.Content class="mx-auto w-full max-w-xl p-0">
				<div class="flex h-full w-full flex-col px-0">
					<div class="mx-auto mt-4 h-2 w-[100px] shrink-0 rounded-full bg-muted"></div>
					<div class="flex h-full w-full flex-col gap-4 px-0">
						<AddTagsContent bind:isOpen bind:toAdd on:added />
					</div>
				</div>
			</Drawer.Content>
		</Drawer.Root>
	{:else}
		<Dialog.Root bind:open={isOpen} closeOnEscape={false} closeOnOutsideClick={false}>
			<Dialog.Content
				class="top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md bg-muted px-0 py-0 duration-200 md:max-w-xl [&>button[data-dialog-close]]:hidden"
			>
				<AddTagsContent bind:isOpen bind:toAdd on:added />
			</Dialog.Content>
		</Dialog.Root>
	{/if}
{/if}
