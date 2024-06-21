<script lang="ts">
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import 'vidstack/bundle';
	import 'vidstack/icons';
	import Fullscreen from './components/fullscreen.svelte';
	import PlayBig from './components/play-big.svelte';
	import SeekBackward from './components/seek-backward.svelte';
	import SeekForward from './components/seek-forward.svelte';
	import Settings from './components/settings.svelte';
	import TimeSlider from './components/time-slider.svelte';
	import Timestamp from './components/timestamp.svelte';

	// ----------------------
	// Variables
	// ----------------------

	// The layout element
	let showLayout = false;

	// The breakpoint for md
	const mdPx = +theme.screens.md.replace('px', '');

	// ----------------------
	// Functions
	// ----------------------

	function updateLayout() {
		showLayout = window.innerWidth < mdPx;
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		updateLayout();
		window.addEventListener('resize', updateLayout);

		return () => {
			window.removeEventListener('resize', updateLayout);
		};
	});
</script>

{#if showLayout}
	<media-controls
		class="pointer-events-none absolute inset-0 z-10 box-border hidden h-full w-full flex-col p-0 opacity-0 transition-opacity duration-200 ease-out data-[visible]:flex data-[show]:opacity-100 data-[visible]:opacity-100 data-[visible]:ease-in"
	>
		<!--  Top 1/3 -->
		<media-controls-group
			class="pointer-events-auto flex w-full basis-3/12 items-start justify-end px-3 pt-2"
		>
			<Settings isMobile={true} />
		</media-controls-group>

		<media-controls-group
			class="pointer-events-auto flex w-full basis-6/12 place-content-center items-center gap-5 sm:gap-8"
		>
			<SeekBackward />
			<PlayBig />
			<SeekForward />
		</media-controls-group>

		<media-controls-group
			class="pointer-events-auto flex w-full basis-3/12 flex-col items-center justify-end px-3 pb-2"
		>
			<div class="flex w-full flex-row justify-between">
				<Timestamp />
				<Fullscreen />
			</div>

			<TimeSlider />
		</media-controls-group>

		<!-- Gradient top -->
		<div
			class="pointer-events-none absolute left-0 top-0 z-[-1] h-[99px] w-full bg-top bg-repeat-x [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)]"
		/>

		<!-- Gradient bottom -->
		<div
			class="pointer-events-none absolute bottom-0 left-0 z-[-1] h-[99px] w-full bg-bottom bg-repeat-x [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)]"
		/>
	</media-controls>
{/if}
