<script lang="ts">
	import { cn } from '$lib/utils';
	import { Logo } from './icons';

	interface Props {
		size?: 'small' | 'medium' | 'large';
	}

	let { size = 'medium' }: Props = $props();
</script>

<div class="flex flex-row items-center gap-3">
	<div class="bg-foreground rounded-md p-1.5">
		<Logo
			class={cn(
				'fill-background',
				{ 'w-10': size === 'small' },
				{ 'w-14': size === 'medium' },
				{ 'w-20': size === 'large' }
			)}
		/>
	</div>
	<div
		class={cn(
			'font-logo flex font-normal tracking-wide',
			{ 'text-3xl': size === 'small' },
			{ 'text-4xl': size === 'medium' },
			{ 'text-5xl': size === 'large' }
		)}
	>
		<span>Off</span>
		<span class="font-semibold">Course</span>
	</div>
</div>
