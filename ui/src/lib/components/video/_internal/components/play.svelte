<script lang="ts">
	import { Icons } from '$components/icons';
</script>

<media-tooltip class="contents">
	<media-tooltip-trigger>
		<media-play-button
			class="group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400 hover:text-secondary data-[focus]:ring-4"
		>
			<Icons.Play
				weight="fill"
				class="hidden size-6 fill-white group-hover:fill-secondary media-ended:!hidden media-paused:block"
			/>
			<Icons.Pause
				weight="fill"
				class="hidden size-6 fill-white group-hover:fill-secondary media-playing:block"
			/>
			<Icons.ArrowCounterClockwise
				weight="bold"
				class="hidden size-6 rotate-12 media-ended:block"
			/>
		</media-play-button>
	</media-tooltip-trigger>
	<media-tooltip-content
		class="z-10 rounded-sm border border-gray-400/20 bg-white px-1 py-0.5 text-sm font-medium text-black animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4"
		placement="top start"
		offset={10}
	>
		<span class="hidden media-ended:!hidden media-paused:block">Play</span>
		<span class="hidden media-playing:block">Pause</span>
		<span class="hidden media-ended:block">Replay</span>
	</media-tooltip-content>
</media-tooltip>
