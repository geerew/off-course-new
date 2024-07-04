<script lang="ts">
	import { Icons } from '$components/icons';
</script>

<media-play-button
	class="hover:text-secondary group relative inline-flex cursor-pointer place-content-center items-center justify-center rounded-full bg-black/60 p-3 outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
>
	<Icons.Play
		weight="fill"
		class="group-hover:fill-secondary media-paused:block media-ended:!hidden hidden size-10 fill-white sm:size-14"
	/>
	<Icons.Pause
		weight="fill"
		class="group-hover:fill-secondary media-playing:block hidden size-10 fill-white sm:size-14"
	/>
	<Icons.ArrowCounterClockwise
		weight="bold"
		class="media-ended:block hidden size-10 rotate-12 sm:size-14"
	/>
</media-play-button>
