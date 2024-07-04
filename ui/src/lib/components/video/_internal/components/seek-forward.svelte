<script lang="ts">
	import { Icons } from '$components/icons';
</script>

<media-seek-button
	class="hover:text-secondary media-ended:hidden group relative inline-flex cursor-pointer place-content-center items-center justify-center rounded-full bg-black/60 p-2 outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
	seconds={10}
>
	<Icons.ArrowClockwise weight="bold" class="size-6 -rotate-45 sm:size-7" />
	<span class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-[8px] font-bold">
		10
	</span>
</media-seek-button>
