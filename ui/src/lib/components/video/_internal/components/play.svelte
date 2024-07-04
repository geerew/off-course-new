<script lang="ts">
	import { Icons } from '$components/icons';
</script>

<media-tooltip class="contents">
	<media-tooltip-trigger>
		<media-play-button
			class="hover:text-secondary group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
		>
			<Icons.Play
				weight="fill"
				class="group-hover:fill-secondary media-paused:block media-ended:!hidden hidden size-6 fill-white"
			/>
			<Icons.Pause
				weight="fill"
				class="group-hover:fill-secondary media-playing:block hidden size-6 fill-white"
			/>
			<Icons.ArrowCounterClockwise
				weight="bold"
				class="media-ended:block hidden size-6 rotate-12"
			/>
		</media-play-button>
	</media-tooltip-trigger>
	<media-tooltip-content
		class="animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm border border-gray-400/20 bg-white px-1 py-0.5 text-sm font-medium text-black"
		placement="top start"
		offset={10}
	>
		<span class="media-paused:block media-ended:!hidden hidden">Play</span>
		<span class="media-playing:block hidden">Pause</span>
		<span class="media-ended:block hidden">Replay</span>
	</media-tooltip-content>
</media-tooltip>
