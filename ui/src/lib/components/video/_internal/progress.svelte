<script lang="ts">
	import { Play } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import Button from '../../ui/button/button.svelte';

	// ----------------------
	// Exports
	// ----------------------

	export let progress: number = 0;
	export let stroke: number = 8;
	export let viewBox: number = 130;
	export let duration: number = 5;

	// ----------------------
	// Variables
	// ----------------------

	let radius: number;
	let normalizedRadius: number;
	let calculatedCircumference: number;
	let interval: ReturnType<typeof setInterval>;

	const dispatch = createEventDispatcher();

	// ----------------------
	// Functions
	// ----------------------

	function startProgress(): void {
		let startTime: number = Date.now();
		let endTime: number = startTime + duration * 1000;

		interval = setInterval(() => {
			let now: number = Date.now();
			let elapsedTime: number = now - startTime;
			let newProgress: number = (elapsedTime / (duration * 1000)) * 100;

			if (now >= endTime) {
				newProgress = 100;
				clearInterval(interval);

				setTimeout(() => {
					dispatch('completed');
				}, 1000);
			}

			progress = newProgress;
		}, 1000 / 60);
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		radius = viewBox / 2;
		normalizedRadius = radius - stroke;
		calculatedCircumference = normalizedRadius * 2 * Math.PI;
		startProgress();

		return () => clearInterval(interval);
	});
</script>

<div
	class=""
	role="progressbar"
	aria-label="Current progress"
	aria-valuemax="100"
	aria-valuenow={progress}
>
	<div class="grid place-items-center [grid-template-areas:_stack] *:[grid-area:_stack]">
		<svg
			fill="none"
			viewBox={`0 0 ${viewBox} ${viewBox}`}
			width={viewBox}
			height={viewBox}
			focusable="false"
		>
			<circle
				r={normalizedRadius}
				cx={radius}
				cy={radius}
				stroke-width={stroke}
				class="stroke-white"
			/>
			<circle
				r={normalizedRadius}
				cx={radius}
				cy={radius}
				stroke-dasharray={`${calculatedCircumference} ${calculatedCircumference}`}
				stroke-width={stroke}
				class="stroke-secondary -rotate-90 [transform-origin:_50%_50%] [transition:_stroke-dashoffset_200ms_linear]"
				style={`stroke-dashoffset: ${calculatedCircumference - (progress / 100) * calculatedCircumference};`}
			/>
		</svg>

		<div>
			<Button
				variant="ghost"
				class="group h-full bg-inherit hover:bg-transparent"
				on:click={() => dispatch('completed')}
			>
				<Play
					class="group-hover:fill-primary group-hover:stroke-primary size-12  fill-white stroke-white duration-200"
				/>
			</Button>
		</div>
	</div>
</div>
