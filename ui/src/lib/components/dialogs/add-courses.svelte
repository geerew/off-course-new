<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Drawer from '$lib/components/ui/drawer';
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import AddCoursesContent from './_internal/add-courses-content.svelte';

	// ----------------------
	// Interfaces
	// ----------------------

	interface $$Slots {
		default: never;
		// the named slot exposes no variables (use an empty object)
		named: object;
		// we have to use the `$$Slots` interface if we have two slots with the same name exposing differently typed props
		trigger: { open: () => void };
	}

	// ----------------------
	// Variables
	// ----------------------

	// True when the sheet is open. This is used to reset stuff on open
	let open = false;

	// The breakpoint for md
	const mdPx = +theme.screens.md.replace('px', '');

	// True when the window size is < md. Set once the window size is known, which happens in onMount
	let isMobile: boolean | null = null;

	// ----------------------
	// Functions
	// ----------------------

	function doOpen() {
		open = true;
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

<slot name="trigger" open={doOpen}>
	<Button
		variant="outline"
		class="group flex h-8 w-36 gap-1.5 bg-primary hover:bg-primary hover:brightness-110"
		on:click={async () => {
			open = true;
		}}
	>
		<Icons.StackPlus class="size-4" />
		<span>Add Courses</span>
	</Button>
</slot>

{#if isMobile !== null}
	{#if isMobile}
		<Drawer.Root bind:open closeOnOutsideClick={false} closeOnEscape={false}>
			<Drawer.Content class="mx-auto h-[90%] w-full p-0">
				<div class="flex h-full w-full flex-col px-0">
					<div class="mx-auto mt-4 h-2 w-[100px] shrink-0 rounded-full bg-muted"></div>
					<AddCoursesContent bind:open on:added />
				</div>
			</Drawer.Content>
		</Drawer.Root>
	{:else}
		<Dialog.Root bind:open closeOnEscape={false} closeOnOutsideClick={false}>
			<Dialog.Content
				class="top-20 max-w-[calc(100vw-4rem)] translate-y-0 overflow-hidden rounded-md bg-muted px-0 py-0 sm:max-w-xl [&>button[data-dialog-close]]:hidden"
			>
				<div class="flex h-[min(calc(100vh-10rem),50rem)] flex-col" data-vaul-no-drag="">
					<AddCoursesContent bind:open on:added />
				</div>
			</Dialog.Content>
		</Dialog.Root>
	{/if}
{/if}
