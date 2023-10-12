export type FileInfo = {
	title: string;
	path: string;
	isSelected?: boolean;
	isExistingCourse?: boolean;
	isParent?: boolean;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type FileSystemInfo = {
	count: number;
	directories: FileInfo[];
	files: FileInfo[];
};
