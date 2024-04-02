<script lang="ts">
	import { Throttle } from '$lib/utils';
	import { createEventDispatcher, onMount } from 'svelte';
	import {
		type MediaDurationChangeEvent,
		type MediaSourceChangeEvent,
		type MediaTimeUpdateEvent
	} from 'vidstack';
	import type { MediaPlayerElement } from 'vidstack/elements';
	import Controls from './controls.svelte';
	import Gestures from './gestures.svelte';

	import Progress from '$components/Progress.svelte';
	import { ASSET_API } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import 'vidstack/player';
	import 'vidstack/player/styles/base.css';
	import 'vidstack/player/ui';
	import { getCtx, setCtx } from './_internal/context';

	// ----------------------
	// Exports
	// ----------------------
	export let title: string;
	export let src: string;
	export let startTime = 0;

	export let nextAsset: Asset | null = null;

	// ----------------------
	// Variables
	// ----------------------

	// The player element
	let player: MediaPlayerElement;

	const dispatch = createEventDispatcher<Record<'progress' | 'finished' | 'next', number>>();

	// Used to only do stuff when the logged second changes
	let lastLoggedSecond = -1;

	// True when the started/finished events are dispatched
	let finishedDispatched = false;

	// Set by the player
	let duration = -1;

	// Video context
	setCtx();
	const ctx = getCtx();

	// A throttle for seeking forward and backward
	const [throttleSeekFw, resetSeekFw] = Throttle(() => {
		player.currentTime += 10;
	}, 200);

	const [throttleSeekBw, resetSeekBw] = Throttle(() => {
		player.currentTime -= 10;
	}, 100);

	// ----------------------
	// Functions
	// ----------------------

	// Reset some stuff when the src changes
	function srcChange(e: MediaSourceChangeEvent) {
		if (!e.detail) return;

		lastLoggedSecond = -1;
		finishedDispatched = false;
		ctx.set({ ...$ctx, ended: false });

		if (!player) return;
		if (Math.floor(startTime) == Math.floor(duration)) {
			player.currentTime = 0;
		} else {
			player.currentTime = startTime ?? 0;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the currentTime when the video can play
	function canPlay() {
		if (!player) return;

		if (Math.floor(startTime) == Math.floor(duration)) {
			player.currentTime = 0;
		} else {
			player.currentTime = startTime ?? 0;
		}
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
	//
	// When the video is paused, nothing will happen
	function timeChange(e: MediaTimeUpdateEvent) {
		if (duration === -1) return;

		const currentSecond = Math.floor(e.detail.currentTime);

		// Do nothing when we have already processed this second
		if (currentSecond === 0 || currentSecond === lastLoggedSecond) return;
		lastLoggedSecond = currentSecond;

		// For the last 5 seconds of the video, dispatch the finished event. After dispatching the
		// event, finishedDispatched will be set to true, so we do not dispatch the event again.
		// Prior dispatch common progress events. This will set finishedDispatched to false
		if (currentSecond >= duration - 5) {
			if (finishedDispatched) return;
			dispatch('finished', Math.floor(duration));
		} else {
			finishedDispatched = false;
			dispatch('progress', currentSecond);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Seek on arrow right/left. This is called on keydown
	function keyboardSeek(e: KeyboardEvent) {
		if (!player) return;

		if (e.key === 'ArrowRight') {
			throttleSeekFw();
		} else if (e.key === 'ArrowLeft') {
			throttleSeekBw();
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Reset the seek throttle. This is called on keyup
	function keyboardReset(e: KeyboardEvent) {
		if (!player) return;

		if (e.key === 'ArrowRight') {
			resetSeekFw();
		} else if (e.key === 'ArrowLeft') {
			resetSeekBw();
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		document.addEventListener('keydown', keyboardSeek);
		document.addEventListener('keyup', keyboardReset);
		return () => {
			document.removeEventListener('keydown', keyboardSeek);
			document.removeEventListener('keyup', keyboardReset);
		};
	});
</script>

<media-player
	bind:this={player}
	class="ring-media-focus group/player aspect-video w-full overflow-hidden rounded-md data-[focus]:ring-4"
	{title}
	src={{
		src: ASSET_API + '/' + src + '/serve',
		type: 'video/mp4'
	}}
	playsInline={true}
	on:source-change={srcChange}
	on:can-play={canPlay}
	on:time-update={timeChange}
	on:duration-change={durationChange}
	on:ended={() => {
		ctx.set({ ...$ctx, ended: true });
	}}
>
	<media-provider />
	<Gestures />
	<Controls />

	{#if $ctx.ended && nextAsset}
		<div class="absolute left-0 top-0 h-full w-full bg-gray-700 py-3 dark:bg-gray-800">
			<div class="flex h-full w-full flex-col place-content-center items-center gap-2.5 text-white">
				<div class="text-muted-foreground">Up next</div>
				<button
					class="hover:text-primary max-w-lg overflow-hidden text-xl duration-200 md:text-2xl lg:max-h-none"
					on:click={() => {
						dispatch('next', 1);
					}}
				>
					<span>
						{nextAsset.prefix}. {nextAsset.title}
					</span>
				</button>

				<Progress duration={8} on:completed={() => dispatch('next', 1)} />
			</div>
		</div>
	{/if}
</media-player>
