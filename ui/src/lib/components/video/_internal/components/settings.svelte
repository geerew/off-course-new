<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Label } from '$lib/components/ui/label';
	import { cn } from '$lib/utils';
	import { ArrowLeft, ChevronRight, ChevronsLeft, ChevronsRight, CirclePlay } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';
	import { MediaRemoteControl } from 'vidstack';
	import { preferences } from '../store';

	//TMP
	export let side: 'top' | 'bottom' = 'top';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();

	// The settings menu element
	let menuEl: HTMLDivElement;

	let open = false;

	// The current section of the menu
	let section: 'top' | 'playback' = 'top';

	// The current playback rate the the available playback rates
	let playbackRate = 1;

	// The breakpoint for md
	const mdPx = +theme.screens.md.replace('px', '');

	// ----------------------
	// Reactive
	// ----------------------

	// Toggle the idle tracking of controls
	function controls(open: boolean) {
		if (open) {
			// Update the video ctx to mark settings as open
			remote.pauseControls();
		} else {
			// Update the video ctx to mark settings as closed and resume idle tracking
			remote.resumeControls();
		}
	}

	// ----------------------
	// Reactive
	// ----------------------
	$: controls(open);

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(menuEl);
		if (!player) return;

		// Set the playback rate
		const playbackRateUnsub = player.subscribe(({ playbackRate: playerPlaybackRate }) => {
			playbackRate = playerPlaybackRate;
			// preferences.set({ playbackRate: playerPlaybackRate });
		});

		// Keep the controls open while the menu is open
		const controlsUnsub = player.subscribe(({ controlsVisible }) => {
			if (open && !controlsVisible) {
				remote.toggleControls();
			}
		});

		// Unsubscribe
		return () => {
			playbackRateUnsub();
			controlsUnsub();
		};
	});
</script>

<!-- Close the settings menu if the window is resized to sm- -->
<svelte:window
	on:resize={() => {
		if (window.innerWidth < mdPx) open = false;
	}}
/>

<div bind:this={menuEl} class="inline-flex">
	<DropdownMenu.Root
		bind:open
		closeOnItemClick={false}
		typeahead={false}
		onOpenChange={(o) => {
			if (!o) section = 'top';
		}}
	>
		<DropdownMenu.Trigger asChild let:builder>
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
					class="group-hover:fill-secondary size-7 fill-white stroke-[1] duration-200 group-data-[open]:rotate-90"
					xmlns="http://www.w3.org/2000/svg"
				>
					<path
						id="circle980"
						d="m 17,12 a 5,5 0 0 1 -5,5 5,5 0 0 1 -5,-5 5,5 0 0 1 5,-5 5,5 0 0 1 5,5 z M 12.22,2 h -0.44 a 2,2 0 0 0 -2,2 v 0.18 a 2,2 0 0 1 -1,1.73 L 8.35,6.16 a 2,2 0 0 1 -2,0 L 6.2,6.08 A 2,2 0 0 0 3.47,6.81 L 3.25,7.19 a 2,2 0 0 0 0.73,2.73 l 0.15,0.1 a 2,2 0 0 1 1,1.72 v 0.51 a 2,2 0 0 1 -1,1.74 l -0.15,0.09 a 2,2 0 0 0 -0.73,2.73 l 0.22,0.38 a 2,2 0 0 0 2.73,0.73 l 0.15,-0.08 a 2,2 0 0 1 2,0 l 0.43,0.25 a 2,2 0 0 1 1,1.73 V 20 a 2,2 0 0 0 2,2 h 0.44 a 2,2 0 0 0 2,-2 v -0.18 a 2,2 0 0 1 1,-1.73 l 0.43,-0.25 a 2,2 0 0 1 2,0 l 0.15,0.08 a 2,2 0 0 0 2.73,-0.73 l 0.22,-0.39 a 2,2 0 0 0 -0.73,-2.73 l -0.15,-0.08 a 2,2 0 0 1 -1,-1.74 v -0.5 a 2,2 0 0 1 1,-1.74 L 20.02,9.92 A 2,2 0 0 0 20.75,7.19 L 20.53,6.81 A 2,2 0 0 0 17.8,6.08 l -0.15,0.08 a 2,2 0 0 1 -2,0 L 15.22,5.91 a 2,2 0 0 1 -1,-1.73 V 4 a 2,2 0 0 0 -2,-2 z"
					/>
				</svg>
			</Button>
		</DropdownMenu.Trigger>

		<DropdownMenu.Content class="min-w-56 p-3 text-sm font-light" {side} align="end">
			<!-- Playback trigger -->
			<DropdownMenu.Item
				class={cn(
					'flex w-full cursor-pointer items-center gap-3 px-2 py-3',
					section === 'top' && 'justify-between'
				)}
				on:click={() => {
					section = section === 'top' ? 'playback' : 'top';
				}}
			>
				<ArrowLeft
					class={cn('hidden size-4 text-white/80', section === 'playback' && 'inline-flex')}
				/>

				<div class="flex items-center gap-1.5">
					<CirclePlay class="size-3.5" />
					<span class="font-semibold leading-3">Playback</span>
				</div>

				<ChevronRight
					class={cn('inline-flex size-4 text-white/70', section === 'playback' && 'hidden')}
				/>
			</DropdownMenu.Item>

			<!-- Playback content -->
			<div
				class={cn(
					'hidden w-full min-w-64 items-center justify-between py-3',
					section === 'playback' && 'flex'
				)}
			>
				<div class="flex w-full flex-col gap-5">
					<!-- Auto play -->
					<div class="bg-muted/60 flex flex-row justify-between px-3 py-3">
						<Label id="autoplay-label" for="autoplay" class="flex grow cursor-pointer text-sm">
							Autoplay
						</Label>
						<Checkbox
							id="autoplay"
							bind:checked={$preferences.autoplay}
							aria-labelledby="autoplay-label"
							class="data-[state=checked]:text-secondary data-[state=checked]:border-secondary border-white data-[state=checked]:bg-transparent"
							on:click={() => {
								preferences.set({ ...$preferences, autoplay: $preferences.autoplay });
							}}
						/>
					</div>

					<div class="bg-muted/60 flex flex-row justify-between px-3 py-3">
						<Label
							id="autoplay-next-label"
							for="autoplay-next"
							class="flex grow cursor-pointer text-sm"
						>
							Autoplay Next
						</Label>
						<Checkbox
							id="autoplay-next"
							bind:checked={$preferences.autoplayNext}
							aria-labelledby="autoplay-next-label"
							class="data-[state=checked]:text-secondary data-[state=checked]:border-secondary border-white data-[state=checked]:bg-transparent"
							on:click={() => {
								preferences.set({ ...$preferences, autoplayNext: !$preferences.autoplayNext });
							}}
						/>
					</div>

					<!-- Speed -->
					<div class="flex w-full flex-col">
						<div class="text-muted-foreground/80 flex flex-row justify-between py-2 text-xs">
							<span>Speed</span>
							<span>{playbackRate === 1 ? 'Normal' : playbackRate + 'x'}</span>
						</div>

						<div class="bg-muted/60 flex w-full flex-row px-2 py-3">
							<ChevronsLeft class="size-4 text-white/70" />
							<media-speed-slider
								class="group relative mx-[7.5px] inline-flex w-full cursor-pointer touch-none select-none items-center outline-none aria-hidden:hidden"
							>
								<!-- Track -->
								<div
									class="relative z-0 h-[5px] w-full rounded-sm bg-white/30 ring-sky-400 group-data-[focus]:ring-[3px]"
								>
									<!-- Fill -->
									<div
										class="bg-secondary absolute h-full w-[var(--slider-fill)] rounded-sm opacity-100 transition-opacity duration-300 will-change-[width] group-data-[active]:opacity-0"
									/>
								</div>

								<!-- Thumb -->
								<div
									class="absolute left-[var(--slider-fill)] top-1/2 z-20 h-[15px] w-[15px] -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#cacaca] bg-white opacity-0 ring-white/40 transition-opacity duration-300 will-change-[left] group-data-[active]:opacity-100 group-data-[dragging]:ring-4"
								/>

								<!-- Steps -->
								<media-slider-steps
									class="absolute left-0 top-0 flex h-full w-full items-center justify-between"
								>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
									<div
										class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
									/>
								</media-slider-steps>
							</media-speed-slider>
							<ChevronsRight class="size-4 text-white/70" />
						</div>
					</div>
				</div>
			</div>
			<!-- {/if} -->
		</DropdownMenu.Content>
	</DropdownMenu.Root>
</div>
