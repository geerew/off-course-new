<script lang="ts">
	import * as Tooltip from '$components/ui/tooltip';
	import { flyAndScale } from '$lib/utils';
	import { Maximize2, Minimize2 } from 'lucide-svelte';
</script>

<!-- Button -->
<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={true}>
	<Tooltip.Trigger class="inline-flex">
		<media-fullscreen-button
			class="ring-media-focus hover:bg-secondary group relative inline-flex cursor-pointer items-center justify-center rounded-md p-1.5 outline-none ring-inset data-[focus]:ring-4"
		>
			<Maximize2 class="media-fullscreen:hidden size-[18px] text-white" />
			<Minimize2 class="media-fullscreen:block hidden size-[18px] text-white" />
		</media-fullscreen-button>
	</Tooltip.Trigger>

	<Tooltip.Content
		class="bg-background text-foreground rounded-sm border-none px-1.5 py-1 text-xs"
		transition={flyAndScale}
		transitionConfig={{ y: 8, duration: 100 }}
		sideOffset={5}
	>
		<Tooltip.Arrow />
		<span class="media-fullscreen:hidden">Enter fullscreen</span>
		<span class="media-fullscreen:block hidden">Exit fullscreen</span>
	</Tooltip.Content>
</Tooltip.Root>
