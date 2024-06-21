<script lang="ts">
	import { Pause, Play } from 'lucide-svelte';
</script>

<media-play-button
	class="hover:text-secondary group relative inline-flex cursor-pointer place-content-center items-center justify-center rounded-full bg-black/60 p-3 outline-none ring-inset ring-sky-400 data-[focus]:ring-4"
>
	<Play
		class="group-hover:fill-secondary media-paused:block media-ended:!hidden hidden size-10 fill-white sm:size-14"
	/>
	<Pause
		class="group-hover:fill-secondary media-playing:block hidden size-10 fill-white sm:size-14"
	/>

	<!-- Replay -->
	<svg
		width="24"
		height="24"
		viewBox="0 0 24 24"
		xmlns="http://www.w3.org/2000/svg"
		class="media-ended:block hidden size-10 sm:size-14"
	>
		<path
			d="m 11.593547,8.7211656 c 0.4401,0.3252 1.0629,0.011 1.0629,-0.5361 v -1.4997 c 0,-0.0788 0.0684,-0.1405 0.1467,-0.131 3.2982,0.4015 5.8533,3.2114 5.8533,6.6180004 0,3.6822 -2.985,6.6673 -6.6672,6.6673 -3.4575004,0 -6.3002304,-2.6318 -6.6344104,-6.0012 -0.03634,-0.3665 -0.33138,-0.6658 -0.69961,-0.6658 l -1.99893,4e-4 c -0.36814,0 -0.66886,0.2994 -0.64468,0.6668 0.34307,5.212 4.67933,9.3333 9.9781304,9.3333 5.5228,0 10,-4.4772 10,-10 0,-5.2564004 -4.0556,-9.5655804 -9.2084,-9.9691304 -0.07,-0.00548 -0.1249,-0.06361 -0.1249,-0.13385 v -1.57548 c 0,-0.54717999 -0.6228,-0.86134999 -1.0629,-0.53617999 L 7.0662466,4.3038456 c -0.3606,0.26649 -0.3606,0.80588 10e-5,1.07235 z"
			fill="currentColor"
			id="path42"
		/>
	</svg>
</media-play-button>
