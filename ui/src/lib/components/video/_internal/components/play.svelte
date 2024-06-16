<script lang="ts">
	import { Pause, Play } from 'lucide-svelte';

	// ----------------------
	// Exports
	// ----------------------
	export let big = false;
</script>

{#if big}
	<media-play-button
		class="hover:text-secondary group relative inline-flex cursor-pointer place-content-center items-center justify-center rounded-full bg-black/40 p-3 outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
	>
		<Play class="group-hover:fill-secondary hidden size-7 fill-white group-data-[paused]:block" />

		<Pause class="group-hover:fill-secondary block size-7 fill-white group-data-[paused]:hidden" />
	</media-play-button>
{:else}
	<media-tooltip class="contents">
		<media-tooltip-trigger>
			<media-play-button
				class="hover:text-secondary group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
			>
				<Play
					class="group-hover:fill-secondary hidden size-6 fill-white group-data-[paused]:block"
				/>

				<Pause
					class="group-hover:fill-secondary block size-6 fill-white group-data-[paused]:hidden"
				/>
			</media-play-button>
		</media-tooltip-trigger>
		<media-tooltip-content
			class="animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm border border-gray-400/20 bg-white px-1 py-0.5 text-sm font-medium text-black"
			placement="top start"
			offset={10}
		>
			<span class="hidden group-data-[paused]:block">Play</span>
			<span class="block group-data-[paused]:hidden">Pause</span>
		</media-tooltip-content>
	</media-tooltip>
{/if}
