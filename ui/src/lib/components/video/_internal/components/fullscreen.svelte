<script lang="ts">
	import { Maximize, Minimize } from 'lucide-svelte';
</script>

<media-tooltip class="contents">
	<media-tooltip-trigger>
		<media-fullscreen-button
			class="hover:text-secondary relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
		>
			<Maximize class="size-5 stroke-2 group-data-[fullscreen]/player:hidden sm:size-6" />
			<Minimize class="hidden size-5 stroke-2 group-data-[fullscreen]/player:block sm:size-6" />
		</media-fullscreen-button>
	</media-tooltip-trigger>
	<media-tooltip-content
		class="animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm border border-gray-400/20 bg-white px-1 py-0.5 text-sm font-medium text-black"
		placement="top end"
		offset={10}
	>
		<span class="group-data-[fullscreen]/player:hidden">Fullscreen</span>
		<span class="hidden group-data-[fullscreen]/player:block">Exit Fullscreen</span>
	</media-tooltip-content>
</media-tooltip>
