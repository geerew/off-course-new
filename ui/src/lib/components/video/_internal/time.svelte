<script lang="ts">
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	// The input (range) element
	let inputEl: HTMLInputElement;

	// The value (%) of the slider
	let value: number = 0;

	// True when the thumb is being dragged
	let isDragging = false;

	// True when the video is paused
	let isPaused = false;

	// When true, the video should be unpaused after the thumb is released
	let shouldUnpause = false;

	// Used to only do stuff when the logged second changes
	let lastLoggedSecond: number = -1;

	// The time in a human-readable format
	let formattedTime = '';

	// Duration and current time of the video (set by the player)
	let duration = 0;
	let time = 0;

	// Create a throttled version of the seeking function

	// ----------------------
	// Functions
	// ----------------------

	// Calculates the seeking time based on the mouse position and dispatches a seeking event
	function seeking(event: MouseEvent) {
		if (!inputEl || duration <= 0) return;

		const bounds = inputEl.getBoundingClientRect();

		// The thumb is 12 px. So we need to offset the mouse position by 6 px to get the correct
		// position
		const thumbOffset = 6;

		const adjustedWidth = bounds.width - thumbOffset * 2;
		let mouseX = Math.max(0, event.clientX - bounds.left - thumbOffset);

		const percentage = Math.max(0, Math.min((mouseX / adjustedWidth) * 100, 100));

		const timePosition = (percentage / 100) * duration;
		remote.seeking(timePosition);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Format the current time and duration into a human-readable format
	function formatTime(): string {
		// Determine the format based on the total duration
		const totalHours = Math.floor(duration / 3600);

		// Helper function to format time
		function formatDuration(
			seconds: number,
			includeMinutes: boolean,
			includeHours: boolean
		): string {
			const hours = Math.floor(seconds / 3600);
			const minutes = Math.floor((seconds % 3600) / 60);
			const remainingSeconds = Math.floor(seconds % 60);

			let timeString = '';

			if (includeHours) {
				timeString += `${hours.toString().padStart(2, '0')}:`;
			}
			if (includeMinutes || includeHours) {
				timeString += `${minutes.toString().padStart(2, '0')}:`;
			}
			timeString += remainingSeconds.toString().padStart(2, '0');

			return timeString;
		}

		const includeMinutes = duration >= 60 || totalHours > 0;
		const includeHours = totalHours > 0;

		// Format the current time and total duration
		const formattedCurrentTime = formatDuration(time, includeMinutes, includeHours);
		const formattedTotalDuration = formatDuration(duration, includeMinutes, includeHours);

		return `${formattedCurrentTime} of ${formattedTotalDuration}`;
	}

	// ----------------------
	// Reactive
	// ----------------------

	// Any time the value changes, update the fill of the slider
	$: if (inputEl) {
		inputEl.style.setProperty('--slider-fill', value + '%');
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(inputEl);
		if (!player) return;

		// Set the duration when the player is ready
		const durationUnsub = player.subscribe(({ duration: playerDuration }) => {
			if (playerDuration === 0) return;
			duration = playerDuration;

			// Set the human-readable time
			formattedTime = formatTime();
			lastLoggedSecond = 0;
		});

		// Update the time and range value
		const timeUnsub = player.subscribe(({ currentTime }) => {
			// Do nothing when the duration has not been set or the thumb is being dragged
			if (duration === 0 || isDragging) return;

			time = currentTime;
			value = (time / duration ?? 0) * 100;

			const currentSecond = Math.floor(currentTime);
			if (currentSecond !== lastLoggedSecond) {
				lastLoggedSecond = currentSecond;
				formattedTime = formatTime();
			}
		});

		// Update the paused state
		const pausedUnsub = player.subscribe(({ paused }) => {
			isPaused = paused;
		});

		// Unsubscribe
		return () => {
			durationUnsub();
			timeUnsub();
			pausedUnsub();
		};
	});
</script>

<media-controls-group class="flex flex-grow items-center gap-1.5">
	<!-- Range -->
	<input
		bind:this={inputEl}
		bind:value
		type="range"
		min="0"
		max="100"
		step="0.01"
		aria-valuenow={time}
		aria-valuetext={formattedTime}
		on:mousemove={seeking}
		on:pointerdown={(e) => {
			if (e.button !== 0) return;

			isDragging = true;
			shouldUnpause = !isPaused;

			if (!isPaused) remote.pause();
		}}
		on:pointerup={() => {
			if (isDragging) {
				remote.seek((value / 100) * duration);
				isDragging = false;
				if (shouldUnpause) remote.play();
			}
		}}
		on:pointercancel={() => (isDragging = false)}
	/>

	<!-- Time -->
	<div class="font-ubuntu-mono mb-0.5 flex gap-0.5 text-sm font-semibold text-white">
		<media-time class="time" type="current" padMinutes={true}></media-time>
		<span>/</span>
		<media-time class="time" type="duration" padMinutes={true}></media-time>
	</div>
</media-controls-group>

<style lang="postcss">
	input[type='range'] {
		@apply h-4 w-full appearance-none bg-transparent;
	}

	input[type='range']:focus {
		outline: none;
	}

	/* Thumb */
	input[type='range']::-webkit-slider-thumb {
		@apply -mt-1 size-3 appearance-none rounded-full border-none bg-white shadow-[0_0_2px_0px_#000000] transition-all duration-200 ease-in-out;

		&:active {
			@apply shadow-[0_0_2px_0px_#000000,0_0_0_3px_#ffffff40];
		}
	}

	input[type='range']::-moz-range-thumb {
		@apply size-3 appearance-none rounded-full border-none bg-white shadow-[0_0_1px_0px_#000000] transition-all delay-0 duration-200 [animation:ease] [transition-timing-function:ease];

		&:active {
			@apply shadow-[0_0_2px_0px_#000000,0_0_0_3px_#ffffff40];
		}
	}

	/* Track */
	input[type='range']::-webkit-slider-runnable-track {
		@apply h-[5px] rounded-sm shadow-none;
		background:
			linear-gradient(theme(colors.secondary.DEFAULT), theme(colors.secondary.DEFAULT)) 0 /
				var(--slider-fill, 0%) 100% no-repeat,
			#ffffff50;
	}

	input[type='range']::-moz-range-track {
		@apply h-[5px] rounded-sm shadow-none;
		background:
			linear-gradient(theme(colors.secondary.DEFAULT), theme(colors.secondary.DEFAULT)) 0 /
				var(--slider-fill, 0%) 100% no-repeat,
			#ffffff50;
	}
</style>
