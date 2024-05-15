<script lang="ts">
	import * as Tooltip from '$components/ui/tooltip';
	import { flyAndScale } from '$lib/utils';
</script>

<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
	<Tooltip.Trigger
		class="hover:text-secondary inline-flex h-auto cursor-pointer items-center text-white"
	>
		<media-mute-button
			class="hover:text-secondary =h-auto group inline-flex cursor-pointer items-center text-white"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 24 24"
				fill="currentColor"
				class="hidden size-[22px] group-data-[state='muted']:block"
			>
				<path
					d="M13.5 4.06c0-1.336-1.616-2.005-2.56-1.06l-4.5 4.5H4.508c-1.141 0-2.318.664-2.66 1.905A9.76 9.76 0 0 0 1.5 12c0 .898.121 1.768.35 2.595.341 1.24 1.518 1.905 2.659 1.905h1.93l4.5 4.5c.945.945 2.561.276 2.561-1.06V4.06ZM17.78 9.22a.75.75 0 1 0-1.06 1.06L18.44 12l-1.72 1.72a.75.75 0 1 0 1.06 1.06l1.72-1.72 1.72 1.72a.75.75 0 1 0 1.06-1.06L20.56 12l1.72-1.72a.75.75 0 1 0-1.06-1.06l-1.72 1.72-1.72-1.72Z"
				/>
			</svg>

			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 24 24"
				fill="currentColor"
				class="hidden size-[22px] group-data-[state='high']:block group-data-[state='low']:block"
			>
				<path
					d="M13.5 4.06c0-1.336-1.616-2.005-2.56-1.06l-4.5 4.5H4.508c-1.141 0-2.318.664-2.66 1.905A9.76 9.76 0 0 0 1.5 12c0 .898.121 1.768.35 2.595.341 1.24 1.518 1.905 2.659 1.905h1.93l4.5 4.5c.945.945 2.561.276 2.561-1.06V4.06ZM18.584 5.106a.75.75 0 0 1 1.06 0c3.808 3.807 3.808 9.98 0 13.788a.75.75 0 0 1-1.06-1.06 8.25 8.25 0 0 0 0-11.668.75.75 0 0 1 0-1.06Z"
				/>
				<path
					d="M15.932 7.757a.75.75 0 0 1 1.061 0 6 6 0 0 1 0 8.486.75.75 0 0 1-1.06-1.061 4.5 4.5 0 0 0 0-6.364.75.75 0 0 1 0-1.06Z"
				/>
			</svg>
		</media-mute-button>
	</Tooltip.Trigger>

	<Tooltip.Content
		class="bg-background text-foreground rounded-sm border-none px-1.5 py-1 text-xs"
		transition={flyAndScale}
		transitionConfig={{ y: 8, duration: 100 }}
		sideOffset={8}
	>
		<span class="media-muted:hidden">Mute</span>
		<span class="media-muted:block hidden">Unmute</span>
	</Tooltip.Content>
</Tooltip.Root>

<media-volume-slider
	class="group relative inline-flex h-[22px] w-full max-w-[80px] cursor-pointer touch-none select-none items-center outline-none aria-hidden:hidden"
>
	<!-- Track -->
	<div
		class="relative z-0 h-[5px] w-full rounded-sm bg-white/30 ring-sky-400 group-data-[focus]:ring-[3px]"
	>
		<!-- Track Fill -->
		<div
			class="bg-secondary absolute h-full w-[var(--slider-fill)] rounded-sm will-change-[width]"
		></div>
	</div>

	<!-- Preview -->
	<media-slider-preview
		class="pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200 data-[visible]:opacity-100"
	>
		<media-slider-value class="rounded-sm bg-white px-2 py-px text-sm font-medium text-black"
		></media-slider-value>
	</media-slider-preview>
</media-volume-slider>
