<script lang="ts">
	import { Button } from '$components/ui/button';
	import { Checkbox } from '$components/ui/checkbox';
	import { Label } from '$components/ui/label';
	import { cn } from '$lib/utils';
	import { ArrowLeft, ChevronRight, ChevronsLeft, ChevronsRight, CirclePlay } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { preferences } from '../../store';

	// ----------------------
	// Exports
	// ----------------------
	export let show: boolean;

	// ----------------------
	// Variables
	// ----------------------
	const dispatch = createEventDispatcher();
</script>

<!-- Trigger -->
<Button
	variant="ghost"
	class={cn(
		'flex w-full cursor-pointer items-center justify-start gap-3 px-2 py-3',
		!show && 'justify-between'
	)}
	on:click={() => {
		dispatch(show ? 'close' : 'open');
	}}
>
	<ArrowLeft class={cn('hidden size-4 text-white/80', show && 'inline-flex')} />

	<div class="flex items-center gap-1.5">
		<CirclePlay class="size-3.5" />
		<span class="font-semibold leading-3">Playback</span>
	</div>

	<ChevronRight class={cn('inline-flex size-4 text-white/70', show && 'hidden')} />
</Button>

<!-- Content -->
<div class={cn('hidden w-full min-w-64 items-center justify-between py-3', show && 'flex')}>
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
			<Label id="autoload-next-label" for="autoload-next" class="flex grow cursor-pointer text-sm">
				Autoload Next
			</Label>
			<Checkbox
				id="autoload-next"
				bind:checked={$preferences.autoloadNext}
				aria-labelledby="autoload-next-label"
				class="data-[state=checked]:text-secondary data-[state=checked]:border-secondary border-white data-[state=checked]:bg-transparent"
				on:click={() => {
					preferences.set({ ...$preferences, autoloadNext: !$preferences.autoloadNext });
				}}
			/>
		</div>

		<!-- Speed -->
		<div class="flex w-full flex-col">
			<div class="text-muted-foreground/80 flex flex-row justify-between py-2 text-xs">
				<span>Speed</span>
				<span>{$preferences.playbackRate === 1 ? 'Normal' : $preferences.playbackRate + 'x'}</span>
			</div>

			<div class="bg-muted/60 flex w-full flex-row px-2 py-3">
				<ChevronsLeft class="size-4 text-white/70" />
				<media-speed-slider
					class="group relative mx-[7.5px] inline-flex w-full cursor-pointer touch-none select-none items-center outline-none aria-hidden:hidden"
					data-vaul-no-drag=""
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
						<template>
							<div
								class="h-1.5 w-0.5 bg-white/50 opacity-0 transition-opacity group-data-[active]:opacity-100"
							/>
						</template>
					</media-slider-steps>
				</media-speed-slider>
				<ChevronsRight class="size-4 text-white/70" />
			</div>
		</div>
	</div>
</div>
