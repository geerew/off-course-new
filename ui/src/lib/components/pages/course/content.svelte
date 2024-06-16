<script lang="ts">
	import { Video } from '$components/video';
	import { ASSET_API, GetBackendUrl } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { UpdateQueryParam } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import PrevNext from './_internal/prev-next.svelte';
	import Title from './_internal/title.svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let selectedAsset: Asset | null;
	export let prevAsset: Asset | null;
	export let nextAsset: Asset | null;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();
</script>

<div class="w-full px-4 md:px-8 lg:px-0">
	<div class="flex h-full w-full flex-col gap-5 pb-8">
		{#if selectedAsset}
			<Title
				asset={selectedAsset}
				on:complete={() => {
					if (!selectedAsset) return;
					selectedAsset.completed = true;
					dispatch('update');
				}}
				on:incomplete={() => {
					if (!selectedAsset) return;
					selectedAsset.completed = false;
					dispatch('update');
				}}
			/>

			{#if selectedAsset.assetType === 'video'}
				<Video
					title={selectedAsset.title}
					src={selectedAsset.id}
					startTime={selectedAsset.videoPos}
					{nextAsset}
					on:progress={(e) => {
						if (!selectedAsset) return;
						selectedAsset.videoPos = e.detail;
						dispatch('update');
					}}
					on:complete={(e) => {
						if (!selectedAsset) return;
						selectedAsset.videoPos = e.detail;
						selectedAsset.completed = true;
						dispatch('update');
					}}
					on:next={() => {
						if (!nextAsset) return;
						UpdateQueryParam('a', nextAsset.id, false);
					}}
				/>
			{:else if selectedAsset.assetType === 'html'}
				<iframe
					src="{GetBackendUrl(ASSET_API)}/{selectedAsset.id}/serve"
					class="h-full w-full"
					title={selectedAsset.title}
				/>
			{/if}

			<PrevNext
				{prevAsset}
				{nextAsset}
				on:prev={() => {
					if (!prevAsset) return;
					UpdateQueryParam('a', prevAsset.id, false);
				}}
				on:next={() => {
					if (!nextAsset) return;
					UpdateQueryParam('a', nextAsset.id, false);
				}}
			/>
		{/if}
	</div>
</div>
