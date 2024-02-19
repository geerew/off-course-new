<script lang="ts">
	// Import styles
	import 'vidstack/player/styles/base.css';

	// Register elements
	import 'vidstack/icons';
	import 'vidstack/player';
	import 'vidstack/player/ui';

	import { ASSET_API } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { type MediaCanPlayEvent, type MediaProviderChangeEvent } from 'vidstack';
	import type { MediaPlayerElement } from 'vidstack/elements';
	import ControlsLayout from './controls-layout.svelte';
	import Gestures from './gestures.svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let asset: Asset;

	// ----------------------
	// Variables
	// ----------------------

	let player: MediaPlayerElement;

	// ----------------------
	// Lifecycle
	// ----------------------

	function onProviderChange(event: MediaProviderChangeEvent) {
		const provider = event.detail;
		// console.log('provider changed', '->', provider?.currentSrc);
	}

	// We can listen for the `can-play` event to be notified when the player is ready.
	function onCanPlay(event: MediaCanPlayEvent) {
		// console.log('ready to play', '->', event.detail);
	}
</script>

<!-- crossorigin -->

<media-player
	class="ring-media-focus aspect-video w-full overflow-hidden rounded-md data-[focus]:ring-4"
	title={asset.title}
	src={ASSET_API + '/' + asset.id + '/serve'}
	playsinline
	on:provider-change={onProviderChange}
	on:can-play={onCanPlay}
	bind:this={player}
>
	<media-provider />
	<Gestures />
	<ControlsLayout />
</media-player>
