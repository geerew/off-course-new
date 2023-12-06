<script lang="ts">
	import { Icons } from '$components/icons';
	import type { Attachment } from '$lib/types/models';
	import { ATTACHMENT_API } from '$lib/utils/api';
	import { cn } from '$lib/utils/general';
	import { createPopover } from '@melt-ui/svelte';
	import { fade } from 'svelte/transition';

	// ----------------------
	// Exports
	// ----------------------

	export let attachments: Attachment[];

	export let showIcon = true;
	export let showCount = true;

	// ----------------------
	// Variables
	// ----------------------

	const {
		elements: { trigger, content },
		states: { open }
	} = createPopover({
		forceVisible: true
	});
</script>

<button
	class={cn('token hover:bg-accent-1', $open && 'bg-accent-1')}
	{...$trigger}
	use:trigger
	on:click|stopPropagation
>
	{#if showIcon}
		<Icons.attachment class="icon" />
	{/if}
	<div class="flex flex-row items-center gap-1">
		{#if showCount}
			{attachments.length} attachment{attachments.length > 1 ? 's' : ''}
		{:else}
			Attachments
		{/if}
		<Icons.chevronRight class="h-3 w-3 duration-200 {$open ? 'rotate-90' : ''}" />
	</div>
</button>

{#if $open}
	<div
		{...$content}
		use:content
		transition:fade={{ duration: 100 }}
		class="bg-background z-10 flex max-h-[10rem] min-w-[7rem] max-w-sm flex-col overflow-y-scroll rounded-lg border p-1.5 shadow-sm"
	>
		{#each attachments as attachment, i}
			{@const lastAttachment = attachments.length - 1 == i}
			<div
				class={cn(!lastAttachment && 'border-b', 'flex flex-row items-center px-2 py-1.5 text-xs')}
			>
				<span class="text-foreground-muted shrink pr-2">{i + 1}</span>
				<span class="grow pr-1.5">{attachment.title}</span>
				<a href={ATTACHMENT_API + '/' + attachment.id + '/serve'} download>
					<Icons.download class="hover:text-primary h-3 w-3" />
				</a>
			</div>
		{/each}
	</div>
{/if}

<style lang="postcss">
	.token {
		@apply inline-flex select-none items-center justify-center gap-2 whitespace-nowrap rounded border px-2 py-1 text-center text-xs;

		& > :global(.icon) {
			@apply h-4 w-4;
		}
	}
</style>
