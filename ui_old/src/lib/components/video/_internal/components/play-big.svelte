<script lang="ts">
	import { Icons } from '$components/icons';
</script>

<media-play-button
	class="group relative inline-flex cursor-pointer place-content-center items-center justify-center rounded-full bg-black/60 p-3 outline-none ring-inset ring-sky-400 hover:text-secondary data-[focus]:ring-4"
>
	<Icons.Play
		weight="fill"
		class="hidden size-10 fill-white group-hover:fill-secondary media-ended:!hidden media-paused:block sm:size-14"
	/>
	<Icons.Pause
		weight="fill"
		class="hidden size-10 fill-white group-hover:fill-secondary media-playing:block sm:size-14"
	/>
	<Icons.ArrowCounterClockwise
		weight="bold"
		class="hidden size-10 rotate-12 media-ended:block sm:size-14"
	/>
</media-play-button>
