<script lang="ts">
	import { Button } from '$components/ui/button';
	import * as Tooltip from '$components/ui/tooltip';
	import { flyAndScale } from '$lib/utils';
	import { onMount } from 'svelte';
	import { MediaRemoteControl } from 'vidstack';
	import { getCtx } from './context';

	// ----------------------
	// Variables
	// ----------------------

	const remote = new MediaRemoteControl();
	let playerEl: HTMLDivElement;

	// True when the video is paused
	let paused = true;
	let ended = false;

	// Video context
	const ctx = getCtx();

	// ----------------------
	// Lifecycle
	// ----------------------

	onMount(() => {
		// Find the player
		const player = remote.getPlayer(playerEl);
		if (!player) return;

		const pausedUnsub = player.subscribe(({ paused: playerPaused, ended: playerEnded }) => {
			paused = playerPaused;
			ended = playerEnded;
		});

		return () => {
			pausedUnsub();
		};
	});
</script>

<div bind:this={playerEl} class="inline-flex">
	<Tooltip.Root openDelay={100} portal={null} closeOnPointerDown={false}>
		<Tooltip.Trigger class="inline-flex">
			<Button
				variant="ghost"
				class="hover:bg-secondary inline-flex h-auto cursor-pointer rounded-sm bg-black px-4 py-1.5 text-white"
				on:click={() => {
					if (ended) {
						if ($ctx.ended) ctx.set({ ...$ctx, ended: false });

						remote.seek(0);
						remote.play();
					} else if (paused) {
						remote.play();
					} else {
						remote.pause();
					}
				}}
			>
				{#if ended}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="currentColor"
						class="size-[22px]"
					>
						<path
							fill-rule="evenodd"
							d="M4.755 10.059a7.5 7.5 0 0 1 12.548-3.364l1.903 1.903h-3.183a.75.75 0 1 0 0 1.5h4.992a.75.75 0 0 0 .75-.75V4.356a.75.75 0 0 0-1.5 0v3.18l-1.9-1.9A9 9 0 0 0 3.306 9.67a.75.75 0 1 0 1.45.388Zm15.408 3.352a.75.75 0 0 0-.919.53 7.5 7.5 0 0 1-12.548 3.364l-1.902-1.903h3.183a.75.75 0 0 0 0-1.5H2.984a.75.75 0 0 0-.75.75v4.992a.75.75 0 0 0 1.5 0v-3.18l1.9 1.9a9 9 0 0 0 15.059-4.035.75.75 0 0 0-.53-.918Z"
							clip-rule="evenodd"
						/>
					</svg>
				{:else if paused}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="currentColor"
						class="size-[22px]"
					>
						<path
							fill-rule="evenodd"
							d="M4.5 5.653c0-1.427 1.529-2.33 2.779-1.643l11.54 6.347c1.295.712 1.295 2.573 0 3.286L7.28 19.99c-1.25.687-2.779-.217-2.779-1.643V5.653Z"
							clip-rule="evenodd"
						/>
					</svg>
				{:else}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="currentColor"
						class="size-[22px]"
					>
						<path
							fill-rule="evenodd"
							d="M6.75 5.25a.75.75 0 0 1 .75-.75H9a.75.75 0 0 1 .75.75v13.5a.75.75 0 0 1-.75.75H7.5a.75.75 0 0 1-.75-.75V5.25Zm7.5 0A.75.75 0 0 1 15 4.5h1.5a.75.75 0 0 1 .75.75v13.5a.75.75 0 0 1-.75.75H15a.75.75 0 0 1-.75-.75V5.25Z"
							clip-rule="evenodd"
						/>
					</svg>
				{/if}
			</Button>
		</Tooltip.Trigger>

		<Tooltip.Content
			class="bg-background text-foreground rounded-sm border-none px-1.5 py-1 text-xs"
			transition={flyAndScale}
			transitionConfig={{ y: 8, duration: 100 }}
			sideOffset={5}
		>
			{#if ended}
				<span>Replay</span>
			{:else if paused}
				<span>Play</span>
			{:else}
				<span>Pause</span>
			{/if}
		</Tooltip.Content>
	</Tooltip.Root>
</div>
