import { array, boolean, enum_, number, object, optional, string, type Output } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const PathClassification = {
	None: 0,
	Ancestor: 1,
	Course: 2,
	Descendant: 3
} as const;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const FileInfoSchema = object({
	title: string(),
	path: string()
});

export type FileInfo = Output<typeof FileInfoSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const DirInfoSchema = object({
	title: string(),
	path: string(),
	classification: enum_(PathClassification),
	isSelected: optional(boolean()),
	isMovingInto: optional(boolean())
});

export type DirInfo = Output<typeof DirInfoSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FileSystemSchema = object({
	count: number(),
	directories: array(DirInfoSchema),
	files: array(FileInfoSchema)
});

export type FileSystem = Output<typeof FileSystemSchema>;
