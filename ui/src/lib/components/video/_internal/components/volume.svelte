<script lang="ts">
	import { cn } from '$lib/utils';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';
	import { preferences } from '../store';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	// The volume element
	let volumeEl: HTMLDivElement;

	// True when the volume slider is hovered. Used to show/hide the slider
	let isHovered = false;

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(volumeEl);
		if (!player) return;

		// Set the playback rate
		const volumeUnsub = player.subscribe(({ volume, muted }) => {
			preferences.set({ ...$preferences, volume, muted });
		});

		// Unsubscribe
		return () => {
			volumeUnsub();
			remote.resumeControls();
		};
	});
</script>

<div
	bind:this={volumeEl}
	class="inline-flex"
	role="button"
	tabindex="0"
	aria-haspopup="true"
	aria-expanded={isHovered}
	on:mouseenter={() => {
		isHovered = true;
	}}
	on:mouseleave={() => {
		isHovered = false;
	}}
>
	<!-- Volume/mute -->
	<media-mute-button
		class="hover:text-secondary group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
	>
		<!-- Muted -->
		<svg
			width="24"
			height="24"
			viewBox="0 0 24 24"
			fill="none"
			aria-hidden="true"
			xmlns="http://www.w3.org/2000/svg"
			class="group-hover:fill-secondary hidden size-6 fill-white group-data-[state=muted]:block"
		>
			<path
				d="m 13,21.283957 c 0,0.58664 -0.599233,0.923486 -1.022748,0.574846 L 5.2431405,16.315744 c -0.022075,-0.01812 -0.048799,-0.02798 -0.076254,-0.02798 H 0.64154607 C 0.28723157,16.287763 0,15.967748 0,15.573118 V 8.4259112 C 0,8.0311729 0.28723157,7.7111583 0.64154607,7.7111583 H 5.1687344 c 0.027445,0 0.054178,-0.00976 0.076254,-0.027981 L 11.977252,2.1411896 C 12.400671,1.792572 13,2.129386 13,2.7160044 Z"
				fill="currentColor"
				id="path173"
				style="stroke-width:1.01572"
			/>
			<path
				d="m 23.823981,9.875524 c 0.234692,-0.2346997 0.234692,-0.6151403 0,-0.8497497 L 22.97426,8.1760245 c -0.234601,-0.2346996 -0.615119,-0.2346996 -0.849721,0 l -2.039673,2.0397425 c -0.04696,0.04687 -0.123024,0.04687 -0.16998,0 L 17.875483,8.1762049 c -0.234691,-0.2346997 -0.615119,-0.2346997 -0.849811,0 l -0.849721,0.8497497 c -0.234601,0.2346094 -0.234601,0.61505 0,0.8497495 l 2.039493,2.0395629 c 0.04696,0.04687 0.04696,0.123028 0,0.169896 l -2.039042,2.03911 c -0.234602,0.2347 -0.234602,0.615141 0,0.84975 l 0.849721,0.84984 c 0.234692,0.234609 0.615119,0.234609 0.849721,-9e-5 l 2.039042,-2.039021 c 0.04696,-0.04696 0.123024,-0.04696 0.16998,0 l 2.039223,2.039291 c 0.234601,0.23461 0.615119,0.23461 0.849721,0 l 0.849721,-0.84984 c 0.234691,-0.234609 0.234691,-0.61505 0,-0.849749 l -2.039223,-2.039291 c -0.04687,-0.04687 -0.04687,-0.123028 0,-0.169896 z"
				fill="currentColor"
				id="path175"
				style="stroke-width:0.901289"
			/>
		</svg>

		<!-- Volume -->
		<svg
			width="24"
			height="24"
			viewBox="0 0 24 24"
			fill="none"
			aria-hidden="true"
			xmlns="http://www.w3.org/2000/svg"
			class="group-hover:fill-secondary hidden size-6 fill-white group-data-[state=high]:block group-data-[state=low]:block"
		>
			<path
				d="m 13,21.284061 c 0,0.586531 -0.599233,0.92338 -1.022748,0.57474 l -6.7341116,-5.54307 c -0.022075,-0.01812 -0.048799,-0.02787 -0.076254,-0.02787 H 0.64154605 C 0.28723156,16.287866 0,15.96785 0,15.573111 V 8.4258884 C 0,8.0311496 0.28723156,7.7112417 0.64154605,7.7112417 H 5.1687343 c 0.027446,0 0.054179,-0.00986 0.076254,-0.027982 L 11.977252,2.1411899 C 12.40067,1.7925719 13,2.1293863 13,2.7160055 Z"
				fill="currentColor"
				id="path190"
				style="stroke-width:1.01572"
			/>
			<path
				d="M 23.499981,5.5 C 23.776128,5.5 24,5.7910178 24,6.1499738 V 17.849967 C 24,18.208962 23.776128,18.5 23.499981,18.5 H 22.500019 C 22.223872,18.5 22,18.208962 22,17.849967 V 6.1499738 C 22,5.7910178 22.223872,5.5 22.500019,5.5 Z"
				fill="currentColor"
				id="path192"
				style="stroke-width:0.855132"
			/>
			<path
				d="M 18.000038,7.9000002 C 18.276194,7.9000002 18.5,8.1985001 18.5,8.5667 v 6.666701 C 18.5,15.6016 18.276194,15.9 18.000038,15.9 H 16.999962 C 16.723881,15.9 16.5,15.6016 16.5,15.233401 V 8.5667 c 0,-0.3681999 0.223881,-0.6666998 0.499962,-0.6666998 z"
				fill="currentColor"
				id="path194"
				style="stroke-width:0.866042"
			/>
		</svg>
	</media-mute-button>

	<!-- Volume slider -->
	<media-volume-slider
		class={cn(
			'group relative inline-flex w-0 cursor-pointer touch-none select-none items-center outline-none transition-all duration-200',
			isHovered && 'ml-3.5 w-20'
		)}
	>
		<!-- Track -->
		<div
			class="relative z-0 h-[5px] w-full rounded-sm bg-white/30 ring-sky-400 group-data-[focus]:ring-[3px]"
		>
			<!-- Fill -->
			<div
				class="bg-secondary absolute h-full w-[var(--slider-fill)] rounded-sm will-change-[width]"
			/>
		</div>

		<!-- Thumb -->
		<div
			class="absolute left-[var(--slider-fill)] top-1/2 z-20 h-[15px] w-[15px] -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#cacaca] bg-white opacity-0 ring-white/40 transition-opacity will-change-[left] group-data-[active]:opacity-100 group-data-[dragging]:ring-4"
		/>

		<media-slider-preview
			class="pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200 data-[visible]:opacity-100"
			noClamp={false}
		>
			<media-slider-value
				class="rounded-sm bg-white px-2 py-px text-[13px] font-medium text-black"
			/>
		</media-slider-preview>
	</media-volume-slider>
</div>
