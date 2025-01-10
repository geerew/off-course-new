<script lang="ts">
	import { cn } from '$lib/utils';
	import { Button } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	type Props = HTMLButtonAttributes & {
		ref?: HTMLButtonElement;
		children?: Snippet;
	};

	let { class: containerClass = '', children, ...restProps }: Props = $props();
</script>

<Button.Root
	type="submit"
	class={cn(
		'bg-background-primary text-foreground-alt-3 disabled:bg-background-primary-alt-1 group flex h-11 w-full cursor-pointer flex-row place-content-center items-center gap-2 rounded-md enabled:hover:brightness-105 disabled:cursor-not-allowed',
		containerClass
	)}
	{...restProps}
>
	{@render children?.()}
</Button.Root>
