import { array, boolean, number, object, optional, string, type Output } from 'valibot';

const FileInfoSchema = object({
	title: string(),
	path: string(),
	isSelected: optional(boolean()),
	isExistingCourse: optional(boolean()),
	isParent: optional(boolean())
});

export type FileInfo = Output<typeof FileInfoSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FileSystemSchema = object({
	count: number(),
	directories: array(FileInfoSchema),
	files: array(FileInfoSchema)
});

export type FileSystem = Output<typeof FileSystemSchema>;
