<script lang="ts">
	import { cn, FlyAndScale } from '$lib/utils';
	import { Popover as PopoverPrimitive } from 'bits-ui';

	type $$Props = PopoverPrimitive.ContentProps;
	let className: $$Props['class'] = undefined;
	export let transition: $$Props['transition'] = FlyAndScale;
	export let transitionConfig: $$Props['transitionConfig'] = undefined;
	export { className as class };
</script>

<PopoverPrimitive.Content
	{transition}
	{transitionConfig}
	class={cn(
		'bg-popover text-popover-foreground z-50 w-72 rounded-md border p-4 shadow-md outline-none',
		className
	)}
	{...$$restProps}
>
	<slot />
</PopoverPrimitive.Content>
