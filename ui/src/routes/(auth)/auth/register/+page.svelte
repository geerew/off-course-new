<script lang="ts">
	import { Logo } from '$lib/components';
	import { FormRegister } from '$lib/components/form';
	import { Button } from 'bits-ui';
</script>

<div class="flex h-full min-h-110 flex-col items-center justify-center gap-8">
	<Logo />

	<div class="flex min-w-sm flex-col gap-5">
		<div class="mb-2.5 flex flex-col gap-2">
			<div class="text-foreground-alt-1 text-center text-lg">Create an account</div>
		</div>

		<FormRegister endpoint="/api/auth/register" />

		<div class="text-foreground-alt-2 text-center">
			Already have an account?
			<Button.Root
				href="/auth/login/"
				class="hover:text-background-primary font-semibold duration-200"
			>
				Login
			</Button.Root>
		</div>
	</div>
</div>
