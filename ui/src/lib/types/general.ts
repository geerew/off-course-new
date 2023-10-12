export type ClassName = string | undefined | null;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ToastData = {
	message: string;
	status: 'success' | 'error' | 'warning' | 'info';
};
