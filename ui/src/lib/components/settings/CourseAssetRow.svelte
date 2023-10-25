<script lang="ts">
	import { Separator } from '$components';
	import { Icons } from '$components/icons';
	import type { Asset } from '$lib/types/models';
	import AttachmentsPopover from './internal/AttachmentsPopover.svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let asset: Asset;
</script>

<div class="flex w-full flex-col gap-4">
	<!-- Title -->
	<span>{asset.prefix}. {asset.title}</span>

	<div class="grid auto-cols-min grid-flow-col items-center gap-5">
		<!-- Type -->
		<span class="token">
			<svelte:component
				this={asset.assetType === 'video'
					? Icons.fileVideo
					: asset.assetType === 'html'
					? Icons.fileHtml
					: Icons.filePdf}
				class="icon"
			/>
			<span>{asset.assetType}</span>
		</span>

		<Separator orientation="vertical" class=" h-4" />

		<!-- Completion status -->
		{#if asset.finished}
			<span class="token !bg-success">
				<Icons.checkCircle class="icon" />
				completed
			</span>
		{:else if asset.started}
			<span class="token !bg-secondary">
				<Icons.halfCircle class="icon [&>:nth-child(n+2)]:fill-white" />
				started
			</span>
		{:else}
			<span class="token">
				<Icons.circle class="icon" />
				not started
			</span>
		{/if}

		<!-- Attachments -->
		{#if asset.attachments && asset.attachments.length > 0}
			<Separator orientation="vertical" class=" h-4" />

			<AttachmentsPopover attachments={asset.attachments} />
		{/if}
	</div>
</div>

<style lang="postcss">
	.token {
		@apply inline-flex select-none items-center justify-center gap-2 whitespace-nowrap rounded px-2 py-1.5 text-center text-xs;
		@apply bg-accent-1;

		& > :global(.icon) {
			@apply h-4 w-4;
		}
	}
</style>
