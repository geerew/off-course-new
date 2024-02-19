<script setup lang="ts">
	import { Play as PlayIcon } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';
	import { MediaControlsGroupElement } from 'vidstack/elements';
	import Fullscreen from './_internal/fullscreen.svelte';
	import Play from './_internal/play.svelte';
	import Settings from './_internal/settings.svelte';
	import Time from './_internal/time.svelte';
	import Volume from './_internal/volume.svelte';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();
	let groupEl: MediaControlsGroupElement;

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(groupEl);
		if (!player) return;

		// Unsubscribe
		return () => {};
	});
</script>

<media-controls
	class="media-controls:opacity-100 absolute inset-0 z-50 flex h-full w-full flex-col overflow-hidden opacity-0 transition-opacity"
>
	<media-controls-group class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
		<media-play-button
			class="media-playing:hidden bg-secondary flex size-20 cursor-pointer items-center justify-center rounded-full hover:brightness-110"
		>
			<PlayIcon class="ml-1 size-10 fill-white text-white" />
		</media-play-button>
	</media-controls-group>

	<!-- Bottom gradient -->
	<div class="absolute bottom-0 left-0 h-32 w-full bg-[linear-gradient(#0000,#000000bf)]" />

	<!-- Controls -->
	<media-controls-group
		role="presentation"
		bind:this={groupEl}
		class="absolute bottom-0 z-10 flex w-full items-end px-2 pb-3"
		on:mouseenter={() => remote.pauseControls()}
		on:mouseleave={() => remote.resumeControls()}
	>
		<div class="flex w-full items-center gap-1.5">
			<Play />
			<Time />
			<Volume />
			<Settings />
			<Fullscreen />
		</div>
	</media-controls-group>
</media-controls>
