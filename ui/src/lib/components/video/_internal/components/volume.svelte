<script lang="ts">
	import { Icons } from '$components/icons';
	import { cn } from '$lib/utils';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import theme from 'tailwindcss/defaultTheme';
	import { MediaRemoteControl, type SliderOrientation } from 'vidstack';
	import type { MediaVolumeSliderElement } from 'vidstack/elements';
	import { preferences } from '../store';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	// The volume element
	let volumeEl: HTMLDivElement;

	// True when the volume slider should be shown
	let show = false;

	// Used to determine the orientation of the slider
	const mdPx = +theme.screens.md.replace('px', '');

	// The orientation of the slider. Based on the window width this will either be 'vertical'
	// or 'horizontal'
	let orientation: SliderOrientation;

	let verticalSliderEl: MediaVolumeSliderElement;

	// ----------------------
	// Functions
	// ----------------------

	// Hide the vertical slider when the user clicks outside of the volume element
	function hideVerticalSlider(e: MouseEvent) {
		if (!volumeEl || !verticalSliderEl) return;

		if (!verticalSliderEl.hasAttribute('data-dragging') && !volumeEl.contains(e.target as Node)) {
			show = false;
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		orientation = window.innerWidth < mdPx ? 'vertical' : 'horizontal';

		// Find the player
		const player = remote.getPlayer(volumeEl);
		if (!player) return;

		// Set the playback rate
		const volumeUnsub = player.subscribe(({ volume, muted }) => {
			preferences.set({ ...$preferences, volume, muted });
		});

		document.addEventListener('mouseup', hideVerticalSlider);

		// Unsubscribe
		return () => {
			volumeUnsub();
			document.removeEventListener('mouseup', hideVerticalSlider);
		};
	});
</script>

<!-- Update the orientation as the window size changes -->
<svelte:window
	on:resize={() => {
		orientation = window.innerWidth < mdPx ? 'vertical' : 'horizontal';
	}}
/>

<div
	bind:this={volumeEl}
	class={cn('relative inline-flex group-data-[pointer=coarse]/player:hidden')}
	role="button"
	tabindex="0"
	aria-haspopup="true"
	aria-expanded={show}
	on:mouseenter={() => {
		show = true;
	}}
	on:mouseleave={() => {
		if (orientation === 'horizontal') show = false;
		else {
			if (verticalSliderEl && !verticalSliderEl.hasAttribute('data-dragging')) show = false;
		}
	}}
>
	<!-- Volume/mute -->
	<media-mute-button
		class="group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400 hover:text-secondary data-[focus]:ring-4"
	>
		<Icons.Mute
			weight="fill"
			class="hidden size-6 fill-white group-hover:fill-secondary group-data-[state=muted]:block"
		/>
		<Icons.VolumeLow
			weight="fill"
			class="hidden size-6 fill-white group-hover:fill-secondary group-data-[state=low]:block"
		/>
		<Icons.VolumeHigh
			weight="fill"
			class="hidden size-6 fill-white group-hover:fill-secondary group-data-[state=high]:block"
		/>
	</media-mute-button>

	{#if orientation === 'horizontal'}
		<media-volume-slider
			class={cn(
				'group relative inline-flex w-0 cursor-pointer touch-none select-none items-center outline-none transition-all duration-200',
				show && 'ml-3.5 w-20'
			)}
			orientation="horizontal"
		>
			<!-- Track -->
			<div
				class="relative z-0 h-[5px] w-full rounded-sm bg-white/30 ring-sky-400 group-data-[focus]:ring-[3px]"
			>
				<!-- Fill -->
				<div
					class="absolute h-full w-[var(--slider-fill)] rounded-sm bg-secondary will-change-[width]"
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
	{:else if show}
		<!-- gap -->
		<div
			class={cn(
				'after:absolute after:-bottom-2 after:left-1/2 after:h-3 after:w-10 after:-translate-x-1/2 after:cursor-auto'
			)}
		/>
		<!-- Slider Popover -->
		<div
			class={cn(
				'absolute left-1/2 top-8 flex h-28 w-10 -translate-x-1/2 cursor-auto place-content-center place-items-center items-center rounded-md border bg-background py-3'
			)}
			transition:fade={{ duration: 150 }}
		>
			<media-volume-slider
				class="group relative my-[7.5px] inline-flex h-full w-full cursor-pointer touch-none select-none place-content-center items-center outline-none aria-hidden:hidden"
				orientation="vertical"
				bind:this={verticalSliderEl}
			>
				<!-- Track -->
				<div
					class="relative z-0 h-full w-[5px] rounded-sm bg-white/30 ring-sky-400 group-data-[focus]:ring-[3px]"
				>
					<!-- Fill -->
					<div
						class="absolute bottom-0 h-[var(--slider-fill)] w-full rounded-sm bg-secondary will-change-[height]"
					/>
				</div>

				<!-- Thumb -->
				<div
					class="absolute bottom-[var(--slider-fill)] left-1/2 z-20 h-[15px] w-[15px] -translate-x-1/2 translate-y-1/2 rounded-full border border-[#cacaca] bg-white opacity-0 ring-white/40 transition-opacity will-change-[bottom] group-data-[active]:opacity-100 group-data-[dragging]:ring-4"
				/>

				<media-slider-preview
					class="pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200 data-[visible]:opacity-100"
					noClamp
					offset={-85}
				>
					<media-slider-value
						class="rounded-sm bg-white px-2 py-px text-[13px] font-medium text-black"
					/>
				</media-slider-preview>
			</media-volume-slider>
		</div>
	{/if}
</div>
