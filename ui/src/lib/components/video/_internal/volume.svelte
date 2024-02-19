<script lang="ts">
	import * as Tooltip from '$components/ui/tooltip';
	import { flyAndScale } from '$lib/utils';
	import { Volume2, VolumeX } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	let inputEl: HTMLInputElement;
	let value: number;
	let preMutedVolume: number;

	// ----------------------
	// Functions
	// ----------------------

	// Changes the value and updates the fill
	function update(value: number) {
		if (value < 0 || value > 1 || !inputEl) return;
		remote.changeVolume(value);
		inputEl.style.setProperty('--slider-fill', (value / +inputEl.max) * 100 + '%');
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When the value changes, update the fill
	$: update(value);

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(inputEl);

		if (!player) return;

		// Default to the player's value
		value = player.volume ?? 1;
		preMutedVolume = value;

		// When the muted state changes, update the value
		const mutedUnsub = player.subscribe(({ muted }) => {
			if (muted) {
				preMutedVolume = value;
				value = 0;
			} else {
				value = preMutedVolume === 0 ? 0.25 : preMutedVolume;
			}
		});

		// Unsubscribe
		return () => {
			mutedUnsub();
		};
	});
</script>

<media-controls-group class="flex shrink-0 items-center gap-1.5">
	<!-- Button -->
	<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={false}>
		<Tooltip.Trigger class="inline-flex">
			<media-mute-button
				class="ring-media-focus hover:bg-secondary group relative inline-flex cursor-pointer items-center justify-center rounded-md p-1.5 text-white outline-none ring-inset data-[focus]:ring-4"
			>
				<VolumeX
					class="hidden size-[18px] group-data-[state='muted']:block [&>:nth-child(1)]:fill-white"
				/>
				<Volume2
					class="hidden size-[18px] group-data-[state='high']:block group-data-[state='low']:block [&>:nth-child(1)]:fill-white"
				/>
			</media-mute-button>
		</Tooltip.Trigger>

		<Tooltip.Content
			class="bg-background text-foreground rounded-sm border-none px-1.5 py-1 text-xs"
			transition={flyAndScale}
			transitionConfig={{ y: 8, duration: 100 }}
			sideOffset={5}
		>
			<Tooltip.Arrow />
			<span class="media-muted:hidden">Mute</span>
			<span class="media-muted:block hidden">Unmute</span>
		</Tooltip.Content>
	</Tooltip.Root>

	<!-- Slider -->
	<input
		bind:this={inputEl}
		bind:value
		type="range"
		min="0"
		max="1"
		step="0.01"
		aria-valuenow={value * 100}
		aria-valuetext={Math.round(value * 100) + '%'}
	/>
</media-controls-group>

<style lang="postcss">
	input[type='range'] {
		@apply h-4 w-20 appearance-none bg-transparent;
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
