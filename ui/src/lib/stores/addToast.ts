import type { ToastData } from '$lib/types/general';
import type { AddToastProps, Toast } from '@melt-ui/svelte';
import { writable } from 'svelte/store';

type AddToastFnType = (props: AddToastProps<ToastData>) => Toast<ToastData>;
export const addToast = writable<AddToastFnType>();
