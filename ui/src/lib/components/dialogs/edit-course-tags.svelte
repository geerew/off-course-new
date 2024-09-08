<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import * as Drawer from '$components/ui/drawer';
	import * as Tooltip from '$components/ui/tooltip';
	import type { CourseTag } from '$lib/types/models';
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import EditCourseTagsContent from './_internal/edit-course-tags-content.svelte';

	// ----------------------
	// Exports
	// ----------------------

	// The course id to add tags to
	export let courseId: string;

	// The existing tags for the course
	export let existingTags: CourseTag[] = [];

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

	$: if (!open) {
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

<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
	<Tooltip.Trigger asChild let:builder>
		<Button
			builders={[builder]}
			variant="ghost"
			class="h-auto cursor-pointer px-2.5 py-1 text-muted-foreground hover:text-foreground"
			on:click={() => {
				isOpen = true;
			}}
		>
			<Icons.Edit class="size-[18px]" weight="bold" />
		</Button>
	</Tooltip.Trigger>

	<Tooltip.Content
		class="select-none rounded-sm border-none bg-foreground px-1.5 py-1 text-xs text-background"
		transitionConfig={{ y: 8, duration: 100 }}
		side="bottom"
	>
		Edit
		<Tooltip.Arrow class="bg-background" />
	</Tooltip.Content>
</Tooltip.Root>

{#if isMobile !== null}
	{#if isMobile}
		<Drawer.Root bind:open={isOpen} closeOnOutsideClick={false} closeOnEscape={false}>
			<Drawer.Content class="mx-auto w-full max-w-xl p-0">
				<div class="flex h-full w-full flex-col px-0">
					<div class="mx-auto mt-4 h-2 w-[100px] shrink-0 rounded-full bg-muted"></div>
					<div class="flex h-full w-full flex-col gap-4 px-0">
						<EditCourseTagsContent bind:isOpen {courseId} {existingTags} bind:toAdd on:updated />
					</div>
				</div>
			</Drawer.Content>
		</Drawer.Root>
	{:else}
		<Dialog.Root bind:open={isOpen} closeOnEscape={false} closeOnOutsideClick={false}>
			<Dialog.Content
				class="top-20 min-w-[20rem] max-w-[26rem] translate-y-0 rounded-md bg-muted px-0 py-0 duration-200 md:max-w-xl [&>button[data-dialog-close]]:hidden"
			>
				<EditCourseTagsContent bind:isOpen {courseId} {existingTags} bind:toAdd on:updated />
			</Dialog.Content>
		</Dialog.Root>
	{/if}
{/if}
