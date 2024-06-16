<div>
	<!-- On single click, play/pause -->
	<media-gesture
		class="pointer-coarse:hidden absolute inset-0 z-0 block h-full w-full"
		event="pointerup"
		action="toggle:paused"
		data-media-gesture=""
		style="pointer-events: none;"
	/>

	<!-- On single click, show controls -->
	<media-gesture
		class="pointer-fine:hidden absolute inset-0 z-0 block h-full w-full"
		event="pointerup"
		action="toggle:controls"
		data-media-gesture=""
		style="pointer-events: none;"
	/>

	<!-- On double click (center), fullscreen -->
	<media-gesture
		class="absolute inset-0 z-0 block h-full w-full"
		event="dblpointerup"
		action="toggle:fullscreen"
		data-media-gesture=""
		style="pointer-events: none;"
	/>

	<!-- On double click (left), seek -10s -->
	<media-gesture
		class="absolute inset-0 z-10 block h-full w-2/6"
		event="dblpointerup"
		action="seek:-10"
		data-media-gesture=""
		style="pointer-events: none;"
	/>

	<!-- On double click (right), seek 10s -->
	<media-gesture
		class="absolute right-0 top-0 z-10 block h-full w-2/6"
		event="dblpointerup"
		action="seek:10"
		data-media-gesture=""
		style="pointer-events: none;"
	/>
</div>
