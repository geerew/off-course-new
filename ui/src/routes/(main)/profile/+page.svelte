<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Input, SubmitButton } from '$lib/components/form';
	import { Edit } from '$lib/components/icons';
	import { AlertDialog, Dialog, Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	let displayNameDialogOpen = $state(false);
	let displayNameInputEl = $state<HTMLInputElement>();
	let displayNameValue = $state<string>('');
	let displayNamePosting = $state(false);

	let deleteDialogOpen = $state(false);

	function deleteAccount() {
		auth.delete().finally(() => {
			deleteDialogOpen = false;
		});
	}

	async function submitDisplayNameForm(event: Event) {
		event.preventDefault();
		displayNamePosting = true;

		const response = await fetch('/api/auth/me', {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				display_name: displayNameValue
			})
		});

		if (response.ok) {
			await auth.me();
			displayNameDialogOpen = false;
		} else {
			const data = await response.json();
			toast.error(data.message);
			displayNamePosting = false;
		}
	}
</script>

{#if auth.user !== null}
	<div class="mx-auto flex max-w-2xl flex-col place-content-center items-start gap-5">
		<!-- Username -->
		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-[15px] uppercase">Username</div>
			<span class="text-background-primary text-2xl">{auth.user.username}</span>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Display name -->
		<div class="flex flex-col gap-3">
			<div class="flex flex-row items-center gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Display Name</div>

				<Dialog.Root
					bind:open={displayNameDialogOpen}
					onOpenChange={(e) => (displayNameValue = '')}
				>
					<Dialog.Trigger
						class="text-foreground-alt-2 hover:text-foreground-alt-1 mb-0.5 cursor-pointer duration-200"
					>
						<Edit class="size-4.5 stroke-2" />
					</Dialog.Trigger>

					<Dialog.Portal>
						<Dialog.Overlay
							class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
						/>
						<Dialog.Content
							class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-[20rem] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200"
							onOpenAutoFocus={(e) => {
								e.preventDefault();
								displayNameInputEl?.focus();
							}}
							onCloseAutoFocus={(e) => {
								e.preventDefault();
							}}
						>
							<form onsubmit={submitDisplayNameForm}>
								<div class="flex flex-col gap-2.5 p-5">
									<div class="flex flex-row gap-2">Display Name:</div>
									<Input
										bind:ref={displayNameInputEl}
										bind:value={displayNameValue}
										name="display name"
										type="text"
										placeholder={auth.user.displayName}
									/>
								</div>

								<div
									class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
								>
									<Dialog.Close
										type="button"
										class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
									>
										Cancel
									</Dialog.Close>

									<SubmitButton
										type="submit"
										disabled={displayNameValue === ''}
										class="h-auto w-24 py-2"
									>
										Update
									</SubmitButton>
								</div>
							</form>
						</Dialog.Content>
					</Dialog.Portal>
				</Dialog.Root>
			</div>
			<span class="text-background-primary text-2xl">{auth.user.displayName}</span>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Delete account -->
		<div class="flex flex-col gap-3">
			<div class="text-foreground-alt-2 text-[15px] uppercase">Delete Account</div>
			<AlertDialog.Root bind:open={deleteDialogOpen}>
				<AlertDialog.Trigger
					class="bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground w-36 cursor-pointer rounded-md py-2 duration-200 select-none"
				>
					Delete Account
				</AlertDialog.Trigger>

				<AlertDialog.Portal>
					<AlertDialog.Overlay
						class="bg-background/70 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50"
					/>

					<AlertDialog.Content
						interactOutsideBehavior="close"
						class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95 data-[state=open]:slide-in-from-top-5 bg-background-alt-1 data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-top-5 fixed top-20 left-1/2 z-50 w-full max-w-[20rem] min-w-[20rem] -translate-x-1/2 overflow-hidden rounded-lg data-[state=closed]:duration-200 data-[state=open]:duration-200 md:max-w-lg"
					>
						<div class="flex flex-col gap-2.5 p-5">
							<div class="flex items-center justify-center">
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="1.5"
									stroke="currentColor"
									class="text-foreground-error size-14"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
									/>
								</svg>
							</div>

							<AlertDialog.Description
								class="text-foreground-alt-1 flex flex-col gap-2 text-center"
							>
								<span class="text-lg">Are you sure you want to delete your account?</span>
								<span class="text-foreground-alt-2">All associated data will be deleted</span>
							</AlertDialog.Description>
						</div>

						<div
							class="bg-background-alt-2 border-background-alt-3 flex w-full items-center justify-end gap-2 border-t px-5 py-2.5"
						>
							<AlertDialog.Cancel
								class="border-background-alt-4 text-foreground-alt-1 hover:bg-background-alt-4 hover:text-foreground w-24 cursor-pointer rounded-md border py-2 duration-200 select-none"
							>
								Cancel
							</AlertDialog.Cancel>

							<AlertDialog.Action
								onclick={deleteAccount}
								class="bg-background-error hover:bg-background-error-alt-1 text-foreground-alt-1 hover:text-foreground w-24 cursor-pointer rounded-md py-2 duration-200 select-none"
							>
								Delete
							</AlertDialog.Action>
						</div>
					</AlertDialog.Content>
				</AlertDialog.Portal>
			</AlertDialog.Root>
		</div>
	</div>
{/if}
