<script setup lang="ts">
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';
	import { MediaControlsGroupElement } from 'vidstack/elements';
	import { getCtx, setCtx } from './_internal/context';
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

	setCtx();
	const ctx = getCtx();

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(groupEl);
		if (!player) return;

		// Listen for the player to be ready
		// const unsub = player.subscribe((e) => {
		// console.log('data', e.bufferedStart, e.bufferedEnd, e.currentTime);
		// });

		// Unsubscribe
		return () => {
			// unsub();
		};
	});
</script>

<media-controls
	class="media-controls:opacity-100 absolute inset-0 z-50 flex h-full w-full flex-col overflow-hidden opacity-0 transition-opacity"
>
	<media-controls-group class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
		<Play type="big" />
	</media-controls-group>

	<!-- Bottom gradient -->
	<div class="absolute bottom-0 left-0 h-32 w-full bg-[linear-gradient(#0000,#000000bf)]" />

	<!-- Controls -->
	<media-controls-group
		role="presentation"
		bind:this={groupEl}
		class="absolute bottom-0 z-10 flex w-full items-end px-2 pb-3"
		on:mouseenter={() => {
			// Update the ctx to mark controls as true
			ctx.set({ ...$ctx, controls: true });
			remote.pauseControls();
		}}
		on:mouseleave={() => {
			// Update the ctx to mark controls as false and resume idle tracking if possible
			ctx.set({ ...$ctx, controls: false });
			if (!$ctx.controls && !$ctx.settings) remote.resumeControls();
		}}
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
