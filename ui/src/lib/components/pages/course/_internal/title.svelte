<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import * as Tooltip from '$components/ui/tooltip';
	import { ATTACHMENT_API } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { cn, flyAndScale } from '$lib/utils';
	import { CircleCheck, Download, FileCode, FileText, FileVideo, Paperclip } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let asset: Asset;

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();
</script>

<div class="flex flex-row items-center justify-between pt-5">
	<!-- Title -->
	<div class="flex flex-row items-center gap-2.5">
		<svelte:component
			this={asset.assetType === 'video'
				? FileVideo
				: asset.assetType === 'html'
					? FileCode
					: FileText}
			class="size-6 stroke-[1]"
		/>
		<span class="text-lg font-medium">{asset.title}</span>
	</div>

	<!-- Actions -->
	<div class="flex flex-row gap-2">
		<!-- Attachments -->
		{#if asset.attachments && asset.attachments.length > 0}
			<DropdownMenu.Root closeOnItemClick={false}>
				<DropdownMenu.Trigger asChild let:builder>
					<Button
						builders={[builder]}
						variant="ghost"
						class="group h-auto items-center gap-1 text-xs"
						on:click={(e) => {
							e.stopPropagation();
						}}
					>
						<Paperclip />
					</Button>
				</DropdownMenu.Trigger>

				<DropdownMenu.Content
					class="bg-foreground text-background flex  w-auto max-w-xs flex-col md:max-w-sm"
					fitViewport={true}
				>
					<div class=" max-h-[10rem] overflow-y-scroll">
						{#each asset.attachments as attachment, i}
							{@const lastAttachment = asset.attachments.length - 1 == i}
							<DropdownMenu.Item
								class="data-[highlighted]:text-background cursor-pointer justify-between gap-3 text-xs data-[highlighted]:bg-transparent data-[highlighted]:underline"
								href={ATTACHMENT_API + '/' + attachment.id + '/serve'}
								download
							>
								<div class="flex flex-row gap-1.5">
									<span class="grow">{attachment.title}</span>
								</div>

								<Download class="flex h-3 w-3 shrink-0" />
							</DropdownMenu.Item>

							{#if !lastAttachment}
								<DropdownMenu.Separator class="bg-muted my-1 -ml-1 -mr-1 block h-px" />
							{/if}
						{/each}
					</div>
					<DropdownMenu.Arrow />
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{/if}

		<!-- Complete/incomplete -->
		<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
			<Tooltip.Trigger asChild let:builder>
				<Button
					builders={[builder]}
					variant="ghost"
					class="h-auto"
					on:click={() => (asset.completed ? dispatch('incomplete') : dispatch('complete'))}
				>
					<CircleCheck
						class={cn(asset.completed && 'fill-success text-success [&>:nth-child(2)]:text-white')}
					/>
				</Button>
			</Tooltip.Trigger>

			<Tooltip.Content
				class="bg-foreground text-background select-none rounded-sm border-none px-1.5 py-1 text-xs"
				transition={flyAndScale}
				transitionConfig={{ y: 8, duration: 100 }}
				side="bottom"
			>
				{#if asset.completed}
					Mark as incomplete
				{:else}
					Mark as complete
				{/if}
				<Tooltip.Arrow class="bg-background" />
			</Tooltip.Content>
		</Tooltip.Root>
	</div>
</div>
