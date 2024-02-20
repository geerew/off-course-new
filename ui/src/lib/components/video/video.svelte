<script lang="ts">
	import 'vidstack/icons';
	import 'vidstack/player';
	import 'vidstack/player/styles/base.css';
	import 'vidstack/player/ui';

	import { ASSET_API } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { createEventDispatcher } from 'svelte';
	import type {
		MediaDurationChangeEvent,
		MediaProviderChangeEvent,
		MediaTimeUpdateEvent
	} from 'vidstack';
	import type { MediaPlayerElement } from 'vidstack/elements';
	import Controls from './controls.svelte';
	import Gestures from './gestures.svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let asset: Asset;

	// ----------------------
	// Variables
	// ----------------------

	let player: MediaPlayerElement;

	const statusDispatch = createEventDispatcher<Record<'started' | 'finished', boolean>>();
	const progressDispatch = createEventDispatcher<Record<'progress', number>>();

	// Used to only do stuff when the logged second changes
	let lastLoggedSecond = -1;

	// True when the started/finished events are dispatched
	let startedDispatched = false;
	let finishedDispatched = false;

	// True when the video ends
	let finished = false;

	// Set by the player
	let duration = -1;

	// ----------------------
	// Functions
	// ----------------------

	// Reset some stuff when the src changes
	function srcChange(e: MediaProviderChangeEvent) {
		if (!e.detail) return;

		lastLoggedSecond = -1;
		startedDispatched = false;
		finishedDispatched = false;
		finished = false;

		if (player) player.currentTime = asset.videoPos ?? 0;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the duration. Update when the src changes
	function durationChange(e: MediaDurationChangeEvent) {
		duration = e.detail;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Each second a progress event will be dispatched. Additionally, when the time is > 3 seconds
	// a started event will be dispatched and when the time is > duration - 5 seconds, a finished
	// event will be dispatched
	function timeChange(e: MediaTimeUpdateEvent) {
		if (duration === -1) return;

		const currentSecond = Math.floor(e.detail.currentTime);

		// Do nothing when the currentSecond is 0 or the same as the last logged second
		if (currentSecond === 0 || currentSecond === lastLoggedSecond) return;

		lastLoggedSecond = currentSecond;

		progressDispatch('progress', currentSecond);

		if (!startedDispatched && currentSecond >= 3) {
			startedDispatched = true;
			statusDispatch('started', true);
		}

		if (!finishedDispatched && currentSecond >= duration - 5) {
			finishedDispatched = true;
			statusDispatch('finished', true);
		}
	}
</script>

<media-player
	class="ring-media-focus aspect-video w-full overflow-hidden rounded-md data-[focus]:ring-4"
	title={asset.title}
	src={{
		src: ASSET_API + '/' + asset.id + '/serve',
		type: 'video/' + asset.path.split('.').pop() ?? 'mp4'
	}}
	playsInline={true}
	on:provider-change={srcChange}
	on:time-update={timeChange}
	on:duration-change={durationChange}
	on:ended={() => {
		finished = true;
	}}
	bind:this={player}
>
	<media-provider />
	<Gestures />
	<Controls />
</media-player>
