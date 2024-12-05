<script lang="ts">
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import 'vidstack/bundle';
	import 'vidstack/icons';
	import Fullscreen from './components/fullscreen.svelte';
	import Play from './components/play.svelte';
	import Settings from './components/settings.svelte';
	import TimeSlider from './components/time-slider.svelte';
	import Timestamp from './components/timestamp.svelte';
	import Volume from './components/volume.svelte';

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
		showLayout = window.innerWidth >= mdPx;
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
		<div class="flex-1"></div>

		<media-controls-group class="pointer-events-auto flex w-full items-center px-3">
			<TimeSlider />
		</media-controls-group>

		<media-controls-group
			class="pointer-events-auto relative z-20 flex w-full items-center gap-5 pb-3 pl-4 pr-5 pt-1"
		>
			<Play />
			<Volume />
			<Timestamp />

			<div class="flex-1" />

			<Settings isMobile={false} />
			<Fullscreen />
		</media-controls-group>

		<!-- Gradient bottom -->
		<div
			class="pointer-events-none absolute bottom-0 left-0 z-[-1] h-[99px] w-full bg-bottom bg-repeat-x [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)]"
		/>
	</media-controls>
{/if}
