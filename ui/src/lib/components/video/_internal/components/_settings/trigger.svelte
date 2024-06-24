<script lang="ts">
	import { Button } from '$components/ui/button';
	import type { Builder } from 'bits-ui';

	// ----------------------
	// Exports
	// ----------------------
	export let builder: Builder;
</script>

<Button
	builders={[builder]}
	variant="ghost"
	class="hover:text-secondary group relative inline-flex h-auto cursor-pointer items-center px-0 py-0 text-white hover:bg-transparent"
>
	<svg
		width="24"
		height="24"
		viewBox="0 0 24 24"
		stroke="currentColor"
		stroke-width="2"
		class="group-hover:fill-secondary size-6 fill-white stroke-[1] duration-200 group-data-[open]:rotate-90"
		xmlns="http://www.w3.org/2000/svg"
	>
		<path
			id="circle980"
			d="m 17,12 a 5,5 0 0 1 -5,5 5,5 0 0 1 -5,-5 5,5 0 0 1 5,-5 5,5 0 0 1 5,5 z M 12.22,2 h -0.44 a 2,2 0 0 0 -2,2 v 0.18 a 2,2 0 0 1 -1,1.73 L 8.35,6.16 a 2,2 0 0 1 -2,0 L 6.2,6.08 A 2,2 0 0 0 3.47,6.81 L 3.25,7.19 a 2,2 0 0 0 0.73,2.73 l 0.15,0.1 a 2,2 0 0 1 1,1.72 v 0.51 a 2,2 0 0 1 -1,1.74 l -0.15,0.09 a 2,2 0 0 0 -0.73,2.73 l 0.22,0.38 a 2,2 0 0 0 2.73,0.73 l 0.15,-0.08 a 2,2 0 0 1 2,0 l 0.43,0.25 a 2,2 0 0 1 1,1.73 V 20 a 2,2 0 0 0 2,2 h 0.44 a 2,2 0 0 0 2,-2 v -0.18 a 2,2 0 0 1 1,-1.73 l 0.43,-0.25 a 2,2 0 0 1 2,0 l 0.15,0.08 a 2,2 0 0 0 2.73,-0.73 l 0.22,-0.39 a 2,2 0 0 0 -0.73,-2.73 l -0.15,-0.08 a 2,2 0 0 1 -1,-1.74 v -0.5 a 2,2 0 0 1 1,-1.74 L 20.02,9.92 A 2,2 0 0 0 20.75,7.19 L 20.53,6.81 A 2,2 0 0 0 17.8,6.08 l -0.15,0.08 a 2,2 0 0 1 -2,0 L 15.22,5.91 a 2,2 0 0 1 -1,-1.73 V 4 a 2,2 0 0 0 -2,-2 z"
		/>
	</svg>
</Button>
