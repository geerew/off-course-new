<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as Tooltip from '$components/ui/tooltip';
	import { flyAndScale } from '$lib/utils';
	import { Pause, Play } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';

	// ----------------------
	// Variables
	// ----------------------

	export let type: 'small' | 'big' = 'small';
	// ----------------------

	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();
	let divEl: HTMLDivElement;
	let paused = true;

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(divEl);
		if (!player) return;

		const unsub = player.subscribe(({ paused: playerPaused }) => {
			paused = playerPaused;
		});

		return () => {
			unsub();
		};
	});
</script>

<div bind:this={divEl}>
	{#if type === 'small'}
		<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={false}>
			<Tooltip.Trigger class="inline-flex">
				<Button
					variant="ghost"
					class="ring-media-focus hover:bg-secondary relative inline-flex h-auto cursor-pointer items-center justify-center rounded-md p-1.5 outline-none ring-inset data-[focus]:ring-4"
					on:click={() => {
						paused ? remote.play() : remote.pause();
					}}
				>
					{#if paused}
						<Play class="media-playing:hidden size-[18px] fill-white text-white" />
					{:else}
						<Pause class="media-paused:hidden size-[18px] fill-white text-white" />
					{/if}
				</Button>
			</Tooltip.Trigger>

			<Tooltip.Content
				class="bg-background text-foreground rounded-sm border-none px-1.5 py-1 text-xs"
				transition={flyAndScale}
				transitionConfig={{ y: 8, duration: 100 }}
				sideOffset={5}
			>
				<Tooltip.Arrow />
				{#if paused}
					<span class="media-paused:block hidden">Play</span>
				{:else}
					<span class="media-paused:hidden">Pause</span>
				{/if}
			</Tooltip.Content>
		</Tooltip.Root>
	{:else if type === 'big'}
		<Button
			variant="ghost"
			class="media-playing:hidden bg-secondary hover:bg-secondary flex h-20 w-20 cursor-pointer items-center justify-center rounded-full hover:brightness-110"
			on:click={() => {
				remote.play();
			}}
		>
			<Play class="ml-1 size-10 fill-white text-white" />
		</Button>
	{/if}
</div>
