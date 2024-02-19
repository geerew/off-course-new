<script lang="ts">
	import * as Tooltip from '$components/ui/tooltip';
	import { flyAndScale } from '$lib/utils';
	import { Pause, Play } from 'lucide-svelte';
</script>

<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={false}>
	<Tooltip.Trigger class="inline-flex">
		<media-play-button
			class="ring-media-focus hover:bg-secondary relative inline-flex cursor-pointer items-center justify-center rounded-md p-1.5 outline-none ring-inset data-[focus]:ring-4"
		>
			<Play class="media-playing:hidden size-[18px] fill-white text-white" />
			<Pause class="media-paused:hidden size-[18px] fill-white text-white" />
		</media-play-button>
	</Tooltip.Trigger>

	<Tooltip.Content
		class="bg-background text-foreground rounded-sm border-none px-1.5 py-1 text-xs"
		transition={flyAndScale}
		transitionConfig={{ y: 8, duration: 100 }}
		sideOffset={5}
	>
		<Tooltip.Arrow />
		<span class="media-paused:block hidden">Play</span>
		<span class="media-paused:hidden">Pause</span>
	</Tooltip.Content>
</Tooltip.Root>
