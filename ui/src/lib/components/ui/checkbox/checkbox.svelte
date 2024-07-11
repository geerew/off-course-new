<script lang="ts">
	import { Icons } from '$components/icons';
	import { cn } from '$lib/utils.js';
	import { Checkbox as CheckboxPrimitive } from 'bits-ui';

	type $$Props = CheckboxPrimitive.Props;
	type $$Events = CheckboxPrimitive.Events;

	let className: $$Props['class'] = undefined;
	export let checked: $$Props['checked'] = false;
	export { className as class };
</script>

<CheckboxPrimitive.Root
	class={cn(
		'peer box-content flex size-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[disabled=true]:cursor-not-allowed data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground data-[disabled=true]:opacity-50',
		className
	)}
	bind:checked
	{...$$restProps}
	on:click
>
	<CheckboxPrimitive.Indicator
		class={cn('flex size-4 items-center justify-center text-current')}
		let:isChecked
		let:isIndeterminate
	>
		{#if isChecked}
			<Icons.Check class="size-3.5" />
		{:else if isIndeterminate}
			<Icons.Minus class="size-3.5" />
		{/if}
	</CheckboxPrimitive.Indicator>
</CheckboxPrimitive.Root>
