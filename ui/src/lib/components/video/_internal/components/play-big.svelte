<script lang="ts">
	import { Pause, Play } from 'lucide-svelte';
</script>

<media-play-button
	class="hover:text-secondary group relative inline-flex cursor-pointer place-content-center items-center justify-center rounded-full bg-black/60 p-3 outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
>
	<Play
		class="group-hover:fill-secondary hidden size-7 fill-white group-data-[paused]:block sm:size-12"
	/>
	<Pause
		class="group-hover:fill-secondary block size-7 fill-white group-data-[paused]:hidden sm:size-12"
	/>
</media-play-button>
