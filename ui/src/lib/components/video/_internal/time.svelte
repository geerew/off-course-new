<script lang="ts">
	import { cn } from '$lib/utils';
	import { getCtx } from './context';

	// ----------------------
	// Variables
	// ----------------------

	let isPointering = false;

	let timeBubbleEl: HTMLDivElement;
	let timeBubbleWidth: number = 0;

	// Video context
	const ctx = getCtx();

	// ----------------------
	// Reactive
	// ----------------------

	// Get the time bubble width
	$: timeBubbleWidth = timeBubbleEl?.getBoundingClientRect().width;
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<media-time-slider
	class="group relative inline-flex h-[22px] w-full cursor-pointer touch-none select-none items-center outline-none aria-hidden:hidden"
	on:mouseenter={() => {
		isPointering = true;
	}}
	on:mouseleave={() => {
		isPointering = false;
	}}
	on:pointerdown={() => {
		ctx.set({ ...$ctx, draggingTimeSlider: true });
	}}
>
	<!-- Track and fill -->
	<div
		class="relative z-0 h-1 w-full rounded-sm bg-white/30 ring-sky-400 transition-[height] duration-200 ease-in-out group-hover:h-1.5 group-data-[focus]:ring-[3px]"
	>
		<div
			class="bg-secondary absolute h-full w-[var(--slider-fill)] rounded-sm will-change-[width]"
		/>
	</div>

	<!-- Actual time -->
	<div
		class={cn(
			'pointer-events-none absolute z-20 flex flex-col items-center transition-opacity duration-200',
			$ctx.draggingTimeSlider || isPointering ? 'opacity-0' : 'opacity-100'
		)}
		bind:this={timeBubbleEl}
		style={`position: absolute; left: min(max(0px, calc(var(--slider-fill) - ${timeBubbleWidth / 2}px)), calc(100% - ${timeBubbleWidth}px)); width: max-content; bottom: calc(100% + var(--media-slider-preview-offset, 0px));`}
	>
		<media-slider-value
			type="current"
			class="bg-foreground text-background rounded-sm px-2 py-px text-sm font-medium"
		/>
	</div>

	<!-- Pointer time -->
	<media-slider-preview
		class="pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200 data-[visible]:opacity-100"
	>
		<media-slider-value
			class="bg-foreground text-background rounded-sm px-2 py-px text-sm font-medium"
		/>
	</media-slider-preview>
</media-time-slider>
