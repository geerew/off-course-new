<script lang="ts">
	import { ASSET_API } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { ChevronLeft, ChevronRight, Play } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import videojs from 'video.js';
	import type Player from 'video.js/dist/types/player';
	import 'video.js/dist/video-js.css';

	// ----------------------
	// Exports
	// ----------------------

	// An asset id
	export let id: string;
	export let startTime = 0;
	export let prevVideo: Asset | null = null;
	export let nextVideo: Asset | null = null;

	// ----------------------
	// Variables
	// ----------------------

	// Types
	//  - When the video ends (- 5 seconds) a 'finished' event is fired
	//  - When the previous button is clicked, a 'previous' event is fired
	//  - When the next button is clicked, a 'next' event is fired
	const dispatch =
		createEventDispatcher<Record<'started' | 'finished' | 'previous' | 'next', boolean>>();

	const progressDispatch = createEventDispatcher<Record<'progress', number>>();

	// The video element
	let videoEl: HTMLVideoElement;

	// The video.js player
	let player: Player;

	// Used to only do stuff when the logged second changes
	let lastLoggedSecond: number = -1;

	// True when the videos moves past 5 seconds before the end
	let endedTrigger = false;

	// Set to true when `useractive` fires and false when `userinactive` fires
	let showPrevNext = true;

	// True while the mouse is over the prev/next button. This causes the `userinactive` to reset
	// to `useractive`
	let prevNextActive = false;

	// True when the video ends
	let videoEnded = false;

	// ----------------------
	// Functions
	// ----------------------

	// Manage the video time updating
	const handleTimeChange = () => {
		// Get the current time and duration
		const currentTime = player.currentTime();
		const videoDuration = player.duration();
		if (typeof currentTime === 'undefined' || typeof videoDuration === 'undefined') return;

		const currentSecond = Math.floor(currentTime);

		// Do nothing when the currentSecond is 0
		if (currentSecond === 0) return;

		if (currentSecond !== lastLoggedSecond) {
			// Update the current progress
			if (currentSecond % 3 === 0) progressDispatch('progress', currentSecond);

			// When the currentSecond is greater than 0 and the currentSecond is greater than the
			// duration - 5, dispatch the finished event. This will mark the video as finished
			if (currentSecond > 0 && currentSecond >= videoDuration - 10 && !endedTrigger) {
				dispatch('finished', true);
				endedTrigger = true;
			}

			lastLoggedSecond = currentSecond;
		}
	};

	// Fired when mouse enters prev/next button
	const handleMouseEnterPrevNextButton = () => {
		prevNextActive = true;
		player.userActive(true);
	};

	// Fired when mouse leaves prev/next button
	const handleMouseLeavePrevNextButton = () => {
		prevNextActive = false;
	};

	// ----------------------
	// Reactive
	// ----------------------

	// Reset things when the video changes
	$: {
		if (id) {
			endedTrigger = false;
			lastLoggedSecond = -1;
			videoEnded = false;
		}
	}

	// ----------------------
	// Lifecycle
	// ----------------------

	// When we have the video element, initialize the player
	onMount(() => {
		player = videojs(videoEl, {
			controls: true,
			autoplay: false,
			preload: 'auto',
			fluid: true,

			playbackRates: [0.5, 1, 1.5, 2]
		});

		// Manage the time update
		player.on('timeupdate', handleTimeChange);

		// Show spinner and hide big play button while loading
		player.on('loadstart', () => {
			player.getChild('BigPlayButton')!.hide();
			player.addClass('vjs-waiting');
		});

		// When the video metadata is loaded, set the start time
		player.on('loadedmetadata', function () {
			player.currentTime(startTime);
		});

		// Hide spinner and show big play button when the video can play
		player.on('canplay', () => {
			player.removeClass('vjs-waiting');
		});

		// Show the prev/next buttons when the user is active
		player.on('useractive', () => {
			showPrevNext = true;
		});

		// Hide the prev/next buttons when the user is inactive (and not over either button)
		player.on('userinactive', () => {
			prevNextActive ? player.userActive(true) : (showPrevNext = false);
		});

		// If there is a next video, toggle `videoEnded` variable when the current video ends
		if (nextVideo) {
			// Event listener for when the video ends
			player.on('ended', () => {
				videoEnded = true;
				setTimeout(() => {
					dispatch('next', true);
				}, 5500);
			});
		}

		return () => {
			player.dispose();
		};
	});
</script>

<!-- svelte-ignore a11y-media-has-caption -->
<div class="relative h-fit w-full">
	<video bind:this={videoEl} class="video-js vjs-fluid" src={ASSET_API + '/' + id + '/serve'} />

	{#if prevVideo}
		<button
			class="border-foreground-muted hover:bg-primary hover:border-primary absolute left-0 top-1/2 z-10 -translate-y-1/2 transform
		rounded-r-md border-y border-r bg-black/30 px-2 py-4 transition duration-700 hover:duration-200"
			class:opacity-100={showPrevNext}
			class:opacity-0={!showPrevNext}
			on:click={() => {
				if (prevVideo) dispatch('previous', true);
			}}
			on:mouseenter={handleMouseEnterPrevNextButton}
			on:mouseleave={handleMouseLeavePrevNextButton}
		>
			<ChevronLeft class="text-white" />
		</button>
	{/if}

	{#if nextVideo}
		<button
			class="border-foreground-muted hover:bg-primary hover:border-primary absolute right-0 top-1/2 z-10 -translate-y-1/2 transform
		rounded-l-md border-y border-l bg-black/30 px-2 py-4 transition duration-700"
			class:opacity-100={showPrevNext}
			class:opacity-0={!showPrevNext}
			on:click={() => {
				if (nextVideo) dispatch('next', true);
			}}
			on:mouseenter={handleMouseEnterPrevNextButton}
			on:mouseleave={handleMouseLeavePrevNextButton}
		>
			<ChevronRight class="text-white" />
		</button>
	{/if}

	{#if videoEnded && nextVideo}
		<div class="absolute left-0 top-0 z-[50] h-full w-full bg-gray-700 py-3 dark:bg-gray-800">
			<div class="flex h-full w-full flex-col place-content-center items-center gap-2.5 text-white">
				<div>Up next</div>
				<button
					class="hover:text-primary max-w-lg overflow-hidden text-xl duration-200 md:text-2xl lg:max-h-none"
					on:click={() => {
						dispatch('next', true);
					}}
				>
					<span>
						{nextVideo.prefix}. {nextVideo.title}
					</span>
				</button>

				<button
					class="group relative mt-2.5"
					on:click={() => {
						dispatch('next', true);
					}}
				>
					<svg width="100" height="100" viewBox="0 0 100 100">
						<circle
							class="progress-ring__background"
							stroke="#d1d5db"
							stroke-width="8"
							fill="transparent"
							r="46"
							cx="50"
							cy="50"
							shape-rendering="gemetricPrecision"
						/>
						<circle
							class={videoEnded ? 'progress-ring__circle animate-ring' : 'progress-ring__circle'}
							stroke="#3b82f6"
							stroke-width="8"
							fill="transparent"
							r="46"
							cx="50"
							cy="50"
							shape-rendering="geometricPrecision"
						/>
					</svg>

					<div class="absolute left-1/2 top-1/2 ml-1 -translate-x-1/2 -translate-y-1/2 transform">
						<Play
							class="group-hover:fill-primary group-hover:stroke-primary h-10 w-10 fill-white duration-200"
						/>
					</div>
				</button>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	:global(.vjs-paused.vjs-has-started .vjs-big-play-button) {
		display: block;
	}

	:global(.video-js) {
		/* Big Play button */
		:global(.vjs-big-play-button) {
			@apply !ml-0 !mt-0 !h-auto !w-auto -translate-x-1/2 -translate-y-1/2 !rounded-full !border-none
			!leading-3;

			:global(.vjs-icon-placeholder::before) {
				@apply bg-primary !relative !mt-0 !flex h-auto w-auto place-content-center !rounded-full p-7 !text-5xl 
				leading-7 duration-200 hover:brightness-110;
			}
		}

		/* Control Bar */
		:global(.vjs-control-bar) {
			@apply !mb-0 !h-10 bg-black;

			/* Play */
			:global(.vjs-play-control) {
				@apply !flex place-content-center !items-center;

				:global(.vjs-icon-placeholder::before) {
					@apply hover:bg-primary !relative !mt-0 !flex h-auto w-auto place-content-center !rounded-md px-1 
					py-1 leading-5 duration-200;
				}
			}

			/* Volume */
			:global(.vjs-volume-panel) {
				@apply !flex !items-center;

				:global(.vjs-icon-placeholder::before) {
					@apply !relative !mt-0 !flex place-content-center;
				}

				:global(.vjs-volume-level) {
					@apply !bg-primary;
				}
			}

			/* Progress */
			:global(.vjs-play-progress) {
				@apply !bg-primary;
			}

			/* Remaining time */
			:global(.vjs-remaining-time) {
				@apply !flex place-content-center !items-center !text-xs;
			}

			/* Playback Speed */
			:global(.vjs-playback-rate.vjs-control) {
				:global(.vjs-playback-rate-value) {
					@apply !relative flex place-content-center items-center !text-xs leading-3;
				}

				:global(.vjs-playback-rate-value) {
					@apply !relative flex place-content-center items-center !text-xs leading-3;
				}

				:global(.vjs-menu) {
					@apply !-left-2 !mb-[25px] !w-14;

					:global(.vjs-menu-content) {
						/* @apply !bg-accent-1; */

						:global(.vjs-menu-item) {
							/* @apply !bg-accent-1 hover:!bg-primary !text-sm focus-visible:bg-inherit focus-visible:!outline-none; */
						}

						:global(.vjs-selected) {
							@apply !bg-primary !text-primary-foreground focus-visible:!bg-primary focus-visible:!outline-none;
						}
					}
				}
			}

			/* Fullscreen */
			:global(.vjs-fullscreen-control) {
				@apply !flex place-content-center !items-center;

				:global(.vjs-icon-placeholder::before) {
					@apply hover:bg-primary !relative !mt-0 !flex h-auto w-auto place-content-center !rounded-md px-1 
					py-1 leading-5 duration-200;
				}
			}
		}
	}

	.progress-ring__background {
		stroke-dasharray: 290;
		stroke-dashoffset: 0;
	}

	.progress-ring__circle {
		transition: stroke-dashoffset 0.35s;
		transform: rotate(-90deg);
		transform-origin: 50% 50%;
		stroke-dasharray: 290;
		stroke-dashoffset: 290;
	}

	/* Trigger the animation when video ends */
	.animate-ring {
		animation: fillRing 5s linear forwards;
	}

	@keyframes fillRing {
		to {
			stroke-dashoffset: 0;
		}
	}
</style>
