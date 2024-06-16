<script lang="ts">
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import Fullscreen from './components/fullscreen.svelte';
	import Play from './components/play.svelte';
	import Settings from './components/settings.svelte';
	import TimeSlider from './components/time-slider.svelte';
	import Timestamp from './components/timestamp.svelte';
	import Volume from './components/volume.svelte';

	// ----------------------
	// Variables
	// ----------------------

	let loadedSize = false;

	// The breakpoint for md
	const mdPx = +theme.screens.md.replace('px', '');

	let largeLayout = true;

	onMount(() => {
		largeLayout = window.innerWidth >= mdPx;
		window.addEventListener('resize', () => {
			largeLayout = window.innerWidth >= mdPx;
		});

		loadedSize = true;
	});
</script>

{#if loadedSize}
	{#if !largeLayout}
		<!-- sm- -->
		<media-controls
			class="pointer-events-none absolute inset-0 z-10 flex h-full w-full flex-col opacity-0 transition-opacity data-[visible]:opacity-100 md:hidden"
			role="group"
			data-visible=""
		>
			<div class="basis-1/3 bg-gradient-to-b from-black/30 to-transparent pt-2">
				<media-controls-group class="pointer-events-auto flex w-full items-center justify-end px-3">
					<Settings side="bottom" />
				</media-controls-group>
			</div>

			<media-controls-group
				class="pointer-events-auto flex w-full basis-1/3 place-content-center items-center"
			>
				<Play big={true} />
			</media-controls-group>

			<div
				class="flex basis-1/3 flex-col justify-end bg-gradient-to-t from-black/30 to-transparent"
			>
				<media-controls-group class="pointer-events-auto flex w-full items-center gap-3 px-5">
					<Timestamp />
					<div class="flex-1" />
					<Fullscreen />
				</media-controls-group>

				<!-- Time Slider -->
				<media-controls-group class="pointer-events-auto flex w-full items-center px-3">
					<TimeSlider />
				</media-controls-group>
			</div>
		</media-controls>
	{:else}
		<!-- md+ -->
		<media-controls
			class="pointer-events-none absolute inset-0 z-10 hidden h-full w-full flex-col bg-gradient-to-t from-black/30 to-transparent opacity-0 transition-opacity data-[visible]:opacity-100 md:flex"
			role="group"
			data-visible=""
		>
			<div class="flex-1"></div>

			<!-- Time Slider -->
			<media-controls-group class="pointer-events-auto flex w-full items-center px-3">
				<TimeSlider />
			</media-controls-group>

			<!-- Controls -->
			<media-controls-group
				class="pointer-events-auto flex w-full items-center gap-4 pb-4 pl-4 pr-5 pt-1"
			>
				<Play />
				<Volume />
				<Timestamp />

				<div class="flex-1" />

				<Settings />
				<Fullscreen />
			</media-controls-group>
		</media-controls>
	{/if}
{/if}
