<media-time-slider
	class="group relative mx-[7.5px] inline-flex h-6 w-full cursor-pointer touch-none select-none items-center outline-none aria-hidden:hidden"
>
	<!-- Track -->
	<div
		class="duration-250 relative z-0 h-[3px] w-full rounded-sm bg-white/30 ring-sky-400 transition-[height] group-hover:h-[9px] group-data-[focus]:ring-[3px]"
	>
		<!-- Progress -->
		<div
			class="absolute z-10 h-full w-[var(--slider-progress)] rounded-sm bg-white/50 will-change-[width]"
		/>

		<!-- Fill -->
		<div
			class="absolute z-[11] h-full w-[var(--slider-fill)] rounded-sm bg-primary will-change-[width]"
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
		<media-slider-value class="rounded-sm bg-white px-2 py-px text-[13px] font-medium text-black" />
	</media-slider-preview>
</media-time-slider>
