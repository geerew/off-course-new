<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import * as DropdownMenu from '$components/ui/dropdown-menu';
	import * as Tooltip from '$components/ui/tooltip';
	import { ATTACHMENT_API, GetBackendUrl } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { cn } from '$lib/utils';
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
				? Icons.FileVideo
				: asset.assetType === 'html'
					? Icons.FileCode
					: Icons.FileText}
			class="size-6 shrink-0 stroke-[1]"
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
						class="group flex h-auto items-center gap-1 px-0 py-0 text-xs hover:bg-transparent"
					>
						<Icons.Paperclip class="size-6" />
					</Button>
				</DropdownMenu.Trigger>

				<DropdownMenu.Content
					class="flex max-h-[10rem] w-auto max-w-xs flex-col overflow-y-scroll md:max-w-sm"
					fitViewport={true}
					sideOffset={-5}
				>
					{#each asset.attachments as attachment, i}
						{@const lastAttachment = asset.attachments.length - 1 == i}
						<DropdownMenu.Item
							class="cursor-pointer justify-between gap-3 text-xs"
							href={GetBackendUrl(ATTACHMENT_API) + '/' + attachment.id + '/serve'}
							download
						>
							<div class="flex flex-row gap-1.5">
								<span class="shrink-0">{i + 1}.</span>
								<span class="grow">{attachment.title}</span>
							</div>

							<Icons.Download class="flex size-3 shrink-0" />
						</DropdownMenu.Item>

						{#if !lastAttachment}
							<DropdownMenu.Separator class="my-1 -ml-1 -mr-1 block h-px bg-muted" />
						{/if}
					{/each}
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
					<Icons.CircleCheck
						class={cn(asset.completed && 'fill-success text-success [&>:nth-child(2)]:text-white')}
					/>
				</Button>
			</Tooltip.Trigger>

			<Tooltip.Content
				class="select-none rounded-sm border-none bg-foreground px-1.5 py-1 text-xs text-background"
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
