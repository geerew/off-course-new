import { array, boolean, number, object, optional, string, type Output } from 'valibot';

const FileInfoSchema = object({
	title: string(),
	path: string()
});

export type FileInfo = Output<typeof FileInfoSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const DirInfoSchema = object({
	title: string(),
	path: string(),
	isSelected: optional(boolean()),
	isExistingCourse: optional(boolean()),
	isParent: optional(boolean())
});

export type DirInfo = Output<typeof DirInfoSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FileSystemSchema = object({
	count: number(),
	directories: array(DirInfoSchema),
	files: array(FileInfoSchema)
});

export type FileSystem = Output<typeof FileSystemSchema>;
