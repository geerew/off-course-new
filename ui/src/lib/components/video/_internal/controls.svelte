<script setup lang="ts">
	import { Loading } from '$components/generic';
	import { MediaRemoteControl } from 'vidstack';
	import { getCtx } from './context';
	import Fullscreen from './fullscreen.svelte';
	import Play from './play.svelte';
	import Time from './time.svelte';
	import Volume from './volume.svelte';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	const ctx = getCtx();
</script>

<!-- Buffering indicator -->
<div
	class="absolute left-1/2 top-1/2 hidden -translate-x-1/2 -translate-y-1/2 transform group-data-[buffering]/player:inline-flex"
>
	<Loading loaderClass="size-20 text-secondary" />
</div>

<media-controls
	class="media-controls:opacity-100 absolute inset-0 z-50 flex h-full w-full flex-col overflow-hidden opacity-0 transition-opacity"
>
	<!-- Controls -->
	<media-controls-group
		role="presentation"
		class="absolute bottom-0 z-10 flex w-full items-end px-2 pb-2"
		on:mouseenter={() => {
			// Update the video ctx to mark controls as open
			ctx.set({ ...$ctx, controlsOpen: true });
			remote.pauseControls();
		}}
		on:mouseleave={() => {
			// Update the video ctx to mark controls as closed and resume idle tracking (if required)
			ctx.set({ ...$ctx, controlsOpen: false });
			if (!$ctx.controlsOpen && !$ctx.settingsOpen) remote.resumeControls();
		}}
	>
		<div class="flex h-full w-full items-center gap-1.5">
			<Play />

			<div class="flex w-full flex-row items-center gap-2 rounded-sm bg-black px-2 py-1.5">
				<Time />
				<Volume />
				<!-- <Settings /> -->
				<Fullscreen />
			</div>
		</div>
	</media-controls-group>
</media-controls>
